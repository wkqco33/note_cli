package cmd

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export [PATH]",
	Short: "노트 및 첨부파일 백업 내보내기",
	Long:  `현재 사용자의 모든 노트와 업로드된 첨부파일을 지정한 ZIP 아카이브 경로로 백업합니다.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		var exportPath string
		if len(args) > 0 {
			exportPath = args[0]
		} else {
			exportPath = defaultExportPath(time.Now())
		}

		fmt.Printf("백업 준비 중... 대상 파일: %s\n", exportPath)

		// 1. 노트 목록 조회
		fmt.Print("노트 데이터를 조회하는 중... ")
		boards, err := client.GetBoards()
		if err != nil {
			fmt.Printf("실패\n")
			return fmt.Errorf("에러: %w", err)
		}
		fmt.Printf("성공 (%d개)\n", len(boards))

		// 2. 첨부파일 목록 조회
		fmt.Print("첨부파일 데이터를 조회하는 중... ")
		files, err := client.GetFiles()
		if err != nil {
			fmt.Printf("실패\n")
			return fmt.Errorf("에러: %w", err)
		}
		fmt.Printf("성공 (%d개)\n", len(files))

		// 3. zip 파일 생성
		zipFile, err := os.Create(exportPath)
		if err != nil {
			return fmt.Errorf("백업 파일을 생성하지 못했습니다: %w", err)
		}
		defer func() {
			_ = zipFile.Close()
		}()

		archive := zip.NewWriter(zipFile)
		defer func() {
			_ = archive.Close()
		}()

		// notes.json 작성
		notesData, err := marshalBackupJSON(boards)
		if err != nil {
			return fmt.Errorf("notes.json 직렬화 실패: %w", err)
		}
		notesEntry, err := archive.Create("notes.json")
		if err != nil {
			return fmt.Errorf("notes.json 생성 실패: %w", err)
		}
		if _, err := notesEntry.Write(notesData); err != nil {
			return fmt.Errorf("notes.json 쓰기 실패: %w", err)
		}

		// files.json 작성
		filesData, err := marshalBackupJSON(files)
		if err != nil {
			return fmt.Errorf("files.json 직렬화 실패: %w", err)
		}
		filesEntry, err := archive.Create("files.json")
		if err != nil {
			return fmt.Errorf("files.json 생성 실패: %w", err)
		}
		if _, err := filesEntry.Write(filesData); err != nil {
			return fmt.Errorf("files.json 쓰기 실패: %w", err)
		}

		// 4. 첨부파일 개별 다운로드 및 압축 저장
		if len(files) > 0 {
			var totalSize int64
			for _, f := range files {
				totalSize += int64(f.FileSize)
			}

			fmt.Println("첨부파일 백업 다운로드 시작...")
			bar := progressbar.DefaultBytes(totalSize, "다운로드 및 압축")

			for _, f := range files {
				body, _, err := client.GetFileStream(f.ID)
				if err != nil {
					fmt.Printf("\n파일 다운로드 실패 (ID: %d, 파일명: %s): %v. 계속 진행합니다.\n", f.ID, f.OriginalFilename, err)
					continue
				}

				entryPath := fmt.Sprintf("files/%d_%s", f.ID, f.OriginalFilename)
				fileEntry, err := archive.Create(entryPath)
				if err != nil {
					_ = body.Close()
					fmt.Printf("\nZIP 내 파일 생성 실패 (%s): %v. 계속 진행합니다.\n", entryPath, err)
					continue
				}

				_, err = io.Copy(io.MultiWriter(fileEntry, bar), body)
				_ = body.Close()
				if err != nil {
					fmt.Printf("\nZIP 복사 중 오류 발생 (%s): %v. 계속 진행합니다.\n", entryPath, err)
					continue
				}
			}
			fmt.Println()
		}

		fmt.Printf("백업이 성공적으로 완료되었습니다! 파일 경로: %s\n", exportPath)
		return nil
	},
}

// marshalBackupJSON 백업 아카이브용으로 들여쓰기된 JSON으로 직렬화
func marshalBackupJSON(v any) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
}

// defaultExportPath 인자가 없을 때 사용하는 기본 백업 파일명
func defaultExportPath(now time.Time) string {
	return fmt.Sprintf("note_backup_%s.zip", now.Format("20060102_150405"))
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
