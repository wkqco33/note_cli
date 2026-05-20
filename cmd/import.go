package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"note_cli/api"

	"github.com/spf13/cobra"
)

var cleanImport bool

var importCmd = &cobra.Command{
	Use:   "import [PATH]",
	Short: "백업 파일에서 노트 및 첨부파일 복원",
	Long:  `지정된 ZIP 백업 아카이브에서 노트와 첨부파일을 가져와 서버에 복원합니다.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		client, err := newAuthenticatedClient()
		if err != nil {
			fmt.Println(err)
			return
		}

		importPath := args[0]
		fmt.Printf("복원 준비 중... 백업 파일: %s\n", importPath)

		// 1. zip 아카이브 열기
		r, err := zip.OpenReader(importPath)
		if err != nil {
			fmt.Printf("백업 파일을 열 수 없습니다: %v\n", err)
			return
		}
		defer r.Close()

		// notes.json 및 files.json 찾기
		var notesFile, filesFile *zip.File
		for _, f := range r.File {
			switch f.Name {
			case "notes.json":
				notesFile = f
			case "files.json":
				filesFile = f
			}
		}

		if notesFile == nil || filesFile == nil {
			fmt.Println("오류: 올바른 백업 파일 형식이 아닙니다 (notes.json 또는 files.json 누락)")
			return
		}

		// notes.json 파싱
		notesReader, err := notesFile.Open()
		if err != nil {
			fmt.Printf("notes.json 열기 실패: %v\n", err)
			return
		}
		var boards []api.BoardRead
		if err := json.NewDecoder(notesReader).Decode(&boards); err != nil {
			notesReader.Close()
			fmt.Printf("notes.json 파싱 실패: %v\n", err)
			return
		}
		notesReader.Close()

		// files.json 파싱
		filesReader, err := filesFile.Open()
		if err != nil {
			fmt.Printf("files.json 열기 실패: %v\n", err)
			return
		}
		var backupFiles []api.FileRead
		if err := json.NewDecoder(filesReader).Decode(&backupFiles); err != nil {
			filesReader.Close()
			fmt.Printf("files.json 파싱 실패: %v\n", err)
			return
		}
		filesReader.Close()

		// --clean 옵션이 켜져있다면 기존 데이터 일괄 삭제
		if cleanImport {
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
				var zipEntry *zip.File
				for _, f := range r.File {
					if f.Name == entryPath {
						zipEntry = f
						break
					}
				}

				if zipEntry == nil {
					fmt.Printf("경고: 백업본 내부에서 파일을 찾을 수 없습니다: %s. 건너뜁니다.\n", entryPath)
					continue
				}

				// 임시 파일 생성
				tmpFile, err := os.CreateTemp("", "note_cli_restore_*")
				if err != nil {
					fmt.Printf("임시 파일 생성 실패: %v\n", err)
					return
				}
				tmpPath := tmpFile.Name()

				// zip 엔트리 데이터를 임시 파일로 복사
				entryReader, err := zipEntry.Open()
				if err != nil {
					tmpFile.Close()
					os.Remove(tmpPath)
					fmt.Printf("백업 파일 읽기 실패 (%s): %v\n", entryPath, err)
					return
				}

				_, err = io.Copy(tmpFile, entryReader)
				entryReader.Close()
				tmpFile.Close()
				if err != nil {
					os.Remove(tmpPath)
					fmt.Printf("임시 파일 쓰기 실패: %v\n", err)
					return
				}

				// 업로드
				fmt.Printf("파일 업로드 중: %s -> ", bf.OriginalFilename)
				uploaded, err := client.UploadFile(tmpPath)
				os.Remove(tmpPath) // 업로드 후 즉시 제거
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
		for _, b := range boards {
			var newImages []string
			for _, img := range b.Images {
				if newURL, ok := oldURLToNewURL[img]; ok {
					newImages = append(newImages, newURL)
				} else {
					newImages = append(newImages, img)
				}
			}

			newBoard := api.BoardCreate{
				Title:    b.Title,
				Content:  b.Content,
				Category: b.Category,
				Images:   newImages,
			}

			fmt.Printf("노트 생성 중: %s... ", b.Title)
			created, err := client.CreateBoard(newBoard)
			if err != nil {
				fmt.Printf("실패 (%v)\n", err)
			} else {
				fmt.Printf("성공 (새 ID: %d)\n", created.ID)
			}
		}

		fmt.Println("복원이 성공적으로 완료되었습니다!")
	},
}

func init() {
	importCmd.Flags().BoolVar(&cleanImport, "clean", false, "복원하기 전에 서버의 모든 기존 노트와 파일을 삭제합니다.")
	rootCmd.AddCommand(importCmd)
}
