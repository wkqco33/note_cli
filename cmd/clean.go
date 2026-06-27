package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "임시 캐시 파일 삭제",
	Long:  "view 명령 실행 중 생성된 임시 이미지 파일을 정리합니다.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tmpDir := os.TempDir()
		matches, err := filepath.Glob(filepath.Join(tmpDir, "note_img_*"))
		if err != nil {
			return fmt.Errorf("캐시 파일 검색 실패: %w", err)
		}

		if len(matches) == 0 {
			fmt.Println("정리할 캐시 파일이 없습니다.")
			return nil
		}

		deleted := 0
		var totalSize int64
		for _, path := range matches {
			info, err := os.Stat(path)
			if err == nil {
				totalSize += info.Size()
			}
			if err := os.Remove(path); err != nil {
				fmt.Printf("삭제 실패: %s (%v)\n", path, err)
			} else {
				deleted++
			}
		}

		fmt.Printf("캐시 파일 %d개 삭제 완료 (%.1f KB)\n", deleted, float64(totalSize)/1024)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
