package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// findCleanCacheFiles 지정 디렉토리에서 임시 캐시 파일(note_img_*)을 찾는다.
func findCleanCacheFiles(dir string) ([]string, error) {
	return filepath.Glob(filepath.Join(dir, "note_img_*"))
}

// cleanCacheFiles 주어진 경로를 삭제하고 삭제한 파일 수와 총 크기를 반환한다.
// sourceDir가 아닌 위치의 경로와 삭제 실패한 파일은 집계에서 제외한다.
func cleanCacheFiles(paths []string, sourceDir string) (deleted int, totalSize int64) {
	for _, path := range paths {
		if filepath.Dir(path) != sourceDir {
			continue
		}
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if err := os.Remove(path); err != nil {
			continue
		}
		totalSize += info.Size()
		deleted++
	}
	return deleted, totalSize
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "임시 캐시 파일 삭제",
	Long:  "view 명령 실행 중 생성된 임시 이미지 파일을 정리합니다.",
	RunE: func(cmd *cobra.Command, args []string) error {
		tmpDir := os.TempDir()
		matches, err := findCleanCacheFiles(tmpDir)
		if err != nil {
			return fmt.Errorf("캐시 파일 검색 실패: %w", err)
		}

		if len(matches) == 0 {
			fmt.Println("정리할 캐시 파일이 없습니다.")
			return nil
		}

		deleted, totalSize := cleanCacheFiles(matches, tmpDir)
		fmt.Printf("캐시 파일 %d개 삭제 완료 (%.1f KB)\n", deleted, float64(totalSize)/1024)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}
