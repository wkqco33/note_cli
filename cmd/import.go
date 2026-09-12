package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"note_cli/api"

	"github.com/wkqco33/wcli"
)

var cleanImport bool

// maxRestoreFileSize 복원 시 개별 파일 최대 크기 (서버 업로드 제한 10MB와 일치).
const maxRestoreFileSize = 10 * 1024 * 1024

var importCmd = &wcli.Command{
	Use:   "import [PATH]",
	Short: "백업 파일에서 노트 및 첨부파일 복원",
	Long: `지정된 ZIP 백업 아카이브에서 노트와 첨부파일을 가져와 서버에 복원합니다.

--clean을 지정하면 복원 전에 서버의 기존 노트와 파일을 모두 삭제합니다.
이 동작은 확인 프롬프트를 거치며, 비대화형 환경에서는 --yes/-y가 필요합니다.`,
	Run: func(ctx *wcli.Context) error {
		if err := requireExactArgs(ctx.Args, 1); err != nil {
			return err
		}
		args := ctx.Args
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		importPath := args[0]
		statusf("복원 준비 중... 백업 파일: %s", importPath)

		// 1. zip 아카이브 열기
		r, err := zip.OpenReader(importPath)
		if err != nil {
			return fmt.Errorf("백업 파일을 열 수 없습니다: %w", err)
		}
		defer func() {
			_ = r.Close()
		}()

		// notes.json 및 files.json 찾기
		notesFile, filesFile := findBackupEntries(r.File)

		if notesFile == nil || filesFile == nil {
			return fmt.Errorf("올바른 백업 파일 형식이 아닙니다 (notes.json 또는 files.json 누락)")
		}

		// notes.json 파싱
		notesReader, err := notesFile.Open()
		if err != nil {
			return fmt.Errorf("notes.json 열기 실패: %w", err)
		}
		var boards []api.BoardRead
		if err := json.NewDecoder(notesReader).Decode(&boards); err != nil {
			_ = notesReader.Close()
			return fmt.Errorf("notes.json 파싱 실패: %w", err)
		}
		_ = notesReader.Close()

		// files.json 파싱
		filesReader, err := filesFile.Open()
		if err != nil {
			return fmt.Errorf("files.json 열기 실패: %w", err)
		}
		var backupFiles []api.FileRead
		if err := json.NewDecoder(filesReader).Decode(&backupFiles); err != nil {
			_ = filesReader.Close()
			return fmt.Errorf("files.json 파싱 실패: %w", err)
		}
		_ = filesReader.Close()

		// --clean 옵션이 켜져있다면 기존 데이터 일괄 삭제 (파괴적 작업이므로 확인 필수)
		if cleanImport {
			confirm, err := askConfirm(
				"--clean: 서버의 모든 기존 노트와 파일을 삭제한 뒤 복원합니다. 계속하시겠습니까?",
				"예 (전체 삭제 후 복원)",
				"아니오 (취소)",
				"--clean을 실행하려면 --yes/-y 플래그를 지정하세요",
			)
			if err != nil {
				return err
			}
			if !confirm {
				fmt.Println("복원이 취소되었습니다.")
				return nil
			}

			statusf("기존 데이터 삭제 요청(--clean) 처리 중...")

			// 기존 노트 삭제
			existingBoards, err := client.GetBoards()
			if err == nil {
				statusf("기존 노트 %d개 삭제 시작", len(existingBoards))
				for _, eb := range existingBoards {
					if err := client.DeleteBoard(eb.ID); err != nil {
						statusf("노트 삭제 실패 (ID: %d): %v", eb.ID, err)
					}
				}
			} else {
				statusf("기존 노트 조회 실패 (계속 진행): %v", err)
			}

			// 기존 파일 삭제
			existingFiles, err := client.GetFiles()
			if err == nil {
				statusf("기존 파일 %d개 삭제 시작", len(existingFiles))
				for _, ef := range existingFiles {
					if err := client.DeleteFile(ef.ID); err != nil {
						statusf("파일 삭제 실패 (ID: %d): %v", ef.ID, err)
					}
				}
			} else {
				statusf("기존 파일 조회 실패 (계속 진행): %v", err)
			}
			statusf("기존 데이터 삭제 완료.")
		}

		// 2. 이미지/첨부파일 복원 및 URL 매핑 생성
		oldURLToNewURL := make(map[string]string)
		if len(backupFiles) > 0 {
			statusf("첨부파일 복원 시작...")
			for _, bf := range backupFiles {
				entryPath := fmt.Sprintf("files/%d_%s", bf.ID, bf.OriginalFilename)
				zipEntry := findBackupFileEntry(r.File, entryPath)

				if zipEntry == nil {
					statusf("경고: 백업본 내부에서 파일을 찾을 수 없습니다: %s. 건너뜁니다.", entryPath)
					continue
				}

				// 서버 업로드 제한을 초과하는 파일은 사전에 건너뛴다
				if exceedsRestoreFileSize(zipEntry) {
					statusf("경고: 파일 크기가 제한(%dMB)을 초과하여 건너뜁니다: %s", maxRestoreFileSize/(1024*1024), bf.OriginalFilename)
					continue
				}

				// 임시 파일 생성
				tmpFile, err := os.CreateTemp("", "note_cli_restore_*")
				if err != nil {
					return fmt.Errorf("임시 파일 생성 실패: %w", err)
				}
				tmpPath := tmpFile.Name()

				// zip 엔트리 데이터를 임시 파일로 복사
				entryReader, err := zipEntry.Open()
				if err != nil {
					_ = tmpFile.Close()
					_ = os.Remove(tmpPath)
					return fmt.Errorf("백업 파일 읽기 실패 (%s): %w", entryPath, err)
				}

				_, err = io.Copy(tmpFile, entryReader)
				_ = entryReader.Close()
				_ = tmpFile.Close()
				if err != nil {
					_ = os.Remove(tmpPath)
					return fmt.Errorf("임시 파일 쓰기 실패: %w", err)
				}

				// 업로드
				uploaded, err := client.UploadFile(tmpPath)
				_ = os.Remove(tmpPath) // 업로드 후 즉시 제거
				if err != nil {
					statusf("파일 업로드 실패 (%s): %v", bf.OriginalFilename, err)
					continue
				}
				statusf("파일 업로드 완료 (%s, 새 ID: %d)", bf.OriginalFilename, uploaded.ID)

				// URL 매핑 관계 기록
				oldURLToNewURL[bf.URL] = uploaded.URL
			}
		}

		// 3. 노트 복원
		fmt.Println("노트 복원 시작...")
		restored := rebuildBoardsForRestore(boards, oldURLToNewURL)
		for _, b := range restored {
			fmt.Printf("노트 생성 중: %s... ", b.Title)
			created, err := client.CreateBoard(b)
			if err != nil {
				fmt.Printf("실패 (%v)\n", err)
			} else {
				fmt.Printf("성공 (새 ID: %d)\n", created.ID)
			}
		}

		fmt.Println("복원이 성공적으로 완료되었습니다!")
		return nil
	},
}

// rebuildBoardsForRestore 노트의 images URL을 새 업로드 URL로 치환해 생성 페이로드를 반환한다.
func rebuildBoardsForRestore(boards []api.BoardRead, urlMapping map[string]string) []api.BoardCreate {
	restored := make([]api.BoardCreate, 0, len(boards))
	for _, b := range boards {
		newImages := remapImageURLs(b.Images, urlMapping)
		restored = append(restored, api.BoardCreate{
			Title:    b.Title,
			Content:  b.Content,
			Category: b.Category,
			Images:   newImages,
		})
	}
	return restored
}

// remapImageURLs URL 리스트를 새 URL로 매핑. 매핑에 없으면 원본 유지.
func remapImageURLs(images []string, urlMapping map[string]string) []string {
	if len(images) == 0 {
		return nil
	}
	out := make([]string, 0, len(images))
	for _, img := range images {
		if newURL, ok := urlMapping[img]; ok {
			out = append(out, newURL)
		} else {
			out = append(out, img)
		}
	}
	return out
}

// findBackupEntries ZIP 아카이브에서 notes.json과 files.json 엔트리를 찾는다.
func findBackupEntries(files []*zip.File) (notesFile *zip.File, filesFile *zip.File) {
	for _, f := range files {
		switch f.Name {
		case "notes.json":
			notesFile = f
		case "files.json":
			filesFile = f
		}
	}
	return notesFile, filesFile
}

// findBackupFileEntry 백업 파일 엔트리 경로(files/<id>_<origname>)로 ZIP 엔트리를 찾는다.
func findBackupFileEntry(files []*zip.File, entryPath string) *zip.File {
	for _, f := range files {
		if f.Name == entryPath {
			return f
		}
	}
	return nil
}

// exceedsRestoreFileSize ZIP 엔트리의 압축 해제 크기가 복원 제한을 초과하는지 판단.
func exceedsRestoreFileSize(f *zip.File) bool {
	return f.UncompressedSize64 > maxRestoreFileSize
}

func init() {
	importCmd.Flags().BoolVar(&cleanImport, "clean", "", false, "복원하기 전에 서버의 모든 기존 노트와 파일을 삭제합니다.")
	rootCmd.AddCommand(importCmd)
}
