package cmd

import (
	"fmt"
	"note_cli/api"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
)

var (
	searchTitle   string
	searchContent string
	searchFile    string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "노트 검색 (제목, 내용, 파일명)",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}

		// 플래그가 하나도 입력되지 않았을 경우 TUI 표시
		if searchTitle == "" && searchContent == "" && searchFile == "" {
			var searchType string
			var query string

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("검색 대상을 선택하세요").
						Options(
							huh.NewOption("제목", "title"),
							huh.NewOption("내용", "content"),
							huh.NewOption("파일", "file"),
						).
						Value(&searchType),
					huh.NewInput().
						Title("검색어를 입력하세요").
						Value(&query).
						Validate(func(str string) error {
							if str == "" {
								return fmt.Errorf("검색어를 입력해야 합니다")
							}
							return nil
						}),
				),
			)

			if err := form.Run(); err != nil {
				fmt.Println("검색이 취소되었습니다.")
				return nil
			}

			switch searchType {
			case "title":
				searchTitle = query
			case "content":
				searchContent = query
			case "file":
				searchFile = query
			}
		}

		notes, err := client.GetBoards()
		if err != nil {
			return fmt.Errorf("노트 목록을 불러오지 못했습니다: %w", err)
		}

		var files []api.FileRead
		if searchFile != "" {
			files, err = client.GetFiles()
			if err != nil {
				return fmt.Errorf("검색용 파일 목록을 불러오지 못했습니다: %w", err)
			}
		}

		var results []api.BoardRead

		results = filterBoards(notes, files, searchTitle, searchContent, searchFile)

		if len(results) == 0 {
			fmt.Println("검색 결과가 없습니다.")
			return nil
		}

		fmt.Printf("총 %d개의 노트를 찾았습니다.\n", len(results))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tUPDATED")
		for _, note := range results {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", note.ID, truncateText(note.Title, 40), note.Category, formatTimestamp(note.UpdatedAt, 16))
		}
		w.Flush()
		return nil
	},
}

// buildURLToFilename 파일 목록에서 URL -> 표시 파일명 맵을 생성
func buildURLToFilename(files []api.FileRead) map[string]string {
	m := make(map[string]string, len(files))
	for _, f := range files {
		m[f.URL] = displayFileName(f)
	}
	return m
}

// filterBoards 제목/내용/첨부파일명 조건으로 노트를 필터링.
// 빈 조건은 해당 조건을 무시(모두 매칭)한다. 검색어는 대소문자 구분 없이 부분 일치.
func filterBoards(notes []api.BoardRead, files []api.FileRead, titleQuery, contentQuery, fileQuery string) []api.BoardRead {
	if titleQuery == "" && contentQuery == "" && fileQuery == "" {
		return append([]api.BoardRead(nil), notes...)
	}

	urlToFilename := buildURLToFilename(files)
	titleLower := strings.ToLower(titleQuery)
	contentLower := strings.ToLower(contentQuery)
	fileLower := strings.ToLower(fileQuery)

	var results []api.BoardRead
	for _, note := range notes {
		if titleQuery != "" && !strings.Contains(strings.ToLower(note.Title), titleLower) {
			continue
		}
		if contentQuery != "" && !strings.Contains(strings.ToLower(note.Content), contentLower) {
			continue
		}
		if fileQuery != "" {
			if !boardMatchesFile(note, urlToFilename, fileLower) {
				continue
			}
		}
		results = append(results, note)
	}
	return results
}

func boardMatchesFile(note api.BoardRead, urlToFilename map[string]string, fileLower string) bool {
	for _, imgUrl := range note.Images {
		filename := urlToFilename[imgUrl]
		if filename == "" {
			filename = fileNameFromURL(imgUrl)
		}
		if strings.Contains(strings.ToLower(filename), fileLower) {
			return true
		}
	}
	return false
}

func init() {
	searchCmd.Flags().StringVarP(&searchTitle, "title", "t", "", "노트 제목으로 검색")
	searchCmd.Flags().StringVarP(&searchContent, "content", "c", "", "노트 내용으로 검색")
	searchCmd.Flags().StringVarP(&searchFile, "file", "f", "", "첨부 파일명으로 검색")
	rootCmd.AddCommand(searchCmd)
}
