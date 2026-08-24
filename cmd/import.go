package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"note_cli/api"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var cleanImport bool

// maxRestoreFileSize 복원 시 개별 파일 최대 크기 (서버 업로드 제한 10MB와 일치).
const maxRestoreFileSize = 10 * 1024 * 1024

var importCmd = &cobra.Command{
	Use:   "import [PATH]",
	Short: "백업 파일에서 노트 및 첨부파일 복원",
	Long:  `지정된 ZIP 백업 아카이브에서 노트와 첨부파일을 가져와 서버에 복원합니다.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		importPath := args[0]
		fmt.Printf("복원 준비 중... 백업 파일: %s\n", importPath)

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
			confirm := false
			err := huh.NewConfirm().
				Title("--clean: 서버의 모든 기존 노트와 파일을 삭제한 뒤 복원합니다. 계속하시겠습니까?").
				Affirmative("예 (전체 삭제 후 복원)").
				Negative("아니오 (취소)").
				Value(&confirm).
				Run()
			if err != nil || !confirm {
				fmt.Println("복원이 취소되었습니다.")
				return nil
			}

			fmt.Println("기존 데이터 삭제 요청(--clean) 처리 중...")

			// 기존 노트 삭제
			fmt.Print("기존 노트를 조회하는 중... ")
			existingBoards, err := client.GetBoards()
			if err == nil {
				fmt.Printf("성공 (%d개 삭제 시작)\n", len(existingBoards))
				for _, eb := range existingBoards {
					if err := client.DeleteBoard(eb.ID); err != nil {
						fmt.Printf("노트 삭제 실패 (ID: %d): %v\n", eb.ID, err)
					}
				}
			} else {
				fmt.Printf("실패 (계속 진행): %v\n", err)
			}

			// 기존 파일 삭제
			fmt.Print("기존 파일을 조회하는 중... ")
			existingFiles, err := client.GetFiles()
			if err == nil {
				fmt.Printf("성공 (%d개 삭제 시작)\n", len(existingFiles))
				for _, ef := range existingFiles {
					if err := client.DeleteFile(ef.ID); err != nil {
						fmt.Printf("파일 삭제 실패 (ID: %d): %v\n", ef.ID, err)
					}
				}
			} else {
				fmt.Printf("실패 (계속 진행): %v\n", err)
			}
			fmt.Println("기존 데이터 삭제 완료.")
		}

		// 2. 이미지/첨부파일 복원 및 URL 매핑 생성
		oldURLToNewURL := make(map[string]string)
		if len(backupFiles) > 0 {
			fmt.Println("첨부파일 복원 시작...")
			for _, bf := range backupFiles {
				entryPath := fmt.Sprintf("files/%d_%s", bf.ID, bf.OriginalFilename)
				zipEntry := findBackupFileEntry(r.File, entryPath)

				if zipEntry == nil {
					fmt.Printf("경고: 백업본 내부에서 파일을 찾을 수 없습니다: %s. 건너뜁니다.\n", entryPath)
					continue
				}

				// 서버 업로드 제한을 초과하는 파일은 사전에 건너뛴다
				if exceedsRestoreFileSize(zipEntry) {
					fmt.Printf("경고: 파일 크기가 제한(%dMB)을 초과하여 건너뜁니다: %s\n", maxRestoreFileSize/(1024*1024), bf.OriginalFilename)
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
				fmt.Printf("파일 업로드 중: %s -> ", bf.OriginalFilename)
				uploaded, err := client.UploadFile(tmpPath)
				_ = os.Remove(tmpPath) // 업로드 후 즉시 제거
				if err != nil {
					fmt.Printf("실패 (%v)\n", err)
					continue
				}
				fmt.Printf("성공 (새 ID: %d)\n", uploaded.ID)

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
	importCmd.Flags().BoolVar(&cleanImport, "clean", false, "복원하기 전에 서버의 모든 기존 노트와 파일을 삭제합니다.")
	rootCmd.AddCommand(importCmd)
}
