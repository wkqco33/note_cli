package cmd

import (
	"fmt"
	"note_cli/api"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/wkqco33/wcli"
)

var (
	searchTitle   string
	searchContent string
	searchFile    string
	searchFormat  string
)

var searchCmd = &wcli.Command{
	Use:   "search",
	Short: "노트 검색 (제목, 내용, 파일명)",
	Run: func(ctx *wcli.Context) error {
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

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

		results := filterBoards(notes, files, searchTitle, searchContent, searchFile)

		format, err := parseOutputFormat(searchFormat)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			if format == formatText {
				fmt.Println("검색 결과가 없습니다.")
			} else {
				rendered, renderErr := renderNotes(nil, format)
				if renderErr != nil {
					return renderErr
				}
				printRenderedNotes(rendered)
			}
			return nil
		}

		if format == formatText {
			fmt.Printf("총 %d개의 노트를 찾았습니다.\n", len(results))
		}
		rendered, err := renderNotes(results, format)
		if err != nil {
			return err
		}
		printRenderedNotes(rendered)
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

// filterBoards 제목/내용/첨부파일명 조건으로 노트를 필터링한다. 빈 조건은 무시.
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
	searchCmd.Flags().StringVar(&searchTitle, "title", "t", "", "노트 제목으로 검색")
	searchCmd.Flags().StringVar(&searchContent, "content", "c", "", "노트 내용으로 검색")
	searchCmd.Flags().StringVar(&searchFile, "file", "f", "", "첨부 파일명으로 검색")
	searchCmd.Flags().StringVar(&searchFormat, "format", "", "", "출력 형식 지정 (text, json, yaml)")
	rootCmd.AddCommand(searchCmd)
}
