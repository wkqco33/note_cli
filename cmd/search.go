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
	Run: func(cmd *cobra.Command, args []string) {
		client, err := newAuthenticatedClient()
		if err != nil {
			fmt.Println(err)
			return
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
				return
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
			fmt.Printf("노트 목록을 불러오지 못했습니다: %v\n", err)
			return
		}

		var files []api.FileRead
		if searchFile != "" {
			files, err = client.GetFiles()
			if err != nil {
				fmt.Printf("검색용 파일 목록을 불러오지 못했습니다: %v\n", err)
				return
			}
		}

		// URL을 통해 파일명을 찾기 위한 맵 생성
		urlToFilename := make(map[string]string)
		if searchFile != "" {
			for _, f := range files {
				urlToFilename[f.URL] = displayFileName(f)
			}
		}

		var results []api.BoardRead

		for _, note := range notes {
			match := true

			// 1. 제목 검색
			if searchTitle != "" && !strings.Contains(strings.ToLower(note.Title), strings.ToLower(searchTitle)) {
				match = false
			}
			// 2. 내용 검색
			if searchContent != "" && !strings.Contains(strings.ToLower(note.Content), strings.ToLower(searchContent)) {
				match = false
			}
			// 3. 첨부파일 검색
			if searchFile != "" {
				fileMatch := false
				for _, imgUrl := range note.Images {
					filename := urlToFilename[imgUrl]
					if filename == "" {
						filename = fileNameFromURL(imgUrl)
					}
					if strings.Contains(strings.ToLower(filename), strings.ToLower(searchFile)) {
						fileMatch = true
						break
					}
				}
				if !fileMatch {
					match = false
				}
			}

			if match {
				results = append(results, note)
			}
		}

		if len(results) == 0 {
			fmt.Println("검색 결과가 없습니다.")
			return
		}

		fmt.Printf("총 %d개의 노트를 찾았습니다.\n", len(results))
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tUPDATED")
		for _, note := range results {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", note.ID, truncateText(note.Title, 40), note.Category, formatTimestamp(note.UpdatedAt, 16))
		}
		w.Flush()
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchTitle, "title", "t", "", "노트 제목으로 검색")
	searchCmd.Flags().StringVarP(&searchContent, "content", "c", "", "노트 내용으로 검색")
	searchCmd.Flags().StringVarP(&searchFile, "file", "f", "", "첨부 파일명으로 검색")
	rootCmd.AddCommand(searchCmd)
}
