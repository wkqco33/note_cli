package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"note_cli/api"
	"note_cli/utils"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/wkqco33/tdraw"
	"github.com/wkqco33/wcli"
)

var (
	noImage    bool
	viewFormat string
	viewJSON   bool
)

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".bmp": true, ".webp": true,
}

func isImageFile(f api.FileRead) bool {
	if strings.HasPrefix(f.ContentType, "image/") {
		return true
	}
	orig := strings.ToLower(f.OriginalFilename)
	if orig == "" {
		orig = strings.ToLower(f.Filename)
	}
	return imageExts[filepath.Ext(orig)]
}

var viewCmd = &wcli.Command{
	Use:   "view [id]",
	Short: "ID로 노트 조회",
	Run: func(ctx *wcli.Context) error {
		if err := requireMaxArgs(ctx.Args, 1); err != nil {
			return err
		}
		args := ctx.Args
		client, err := newAuthenticatedClient()
		if err != nil {
			return err
		}
		defer func() { _ = client.Close() }()

		id, ok, err := resolveBoardID(client, args, "조회할 노트를 선택하세요", "조회할 노트가 없습니다.")
		if err != nil || !ok {
			return err
		}
		var note *api.BoardRead
		err = utils.WithSpinner("노트를 불러오는 중...", func() error {
			var innerErr error
			note, innerErr = client.GetBoard(id)
			return innerErr
		})
		if err != nil {
			return fmt.Errorf("노트를 불러오지 못했습니다: %w", err)
		}

		// 구조화 출력 형식이면 렌더링/이미지 처리 없이 노트 데이터만 출력
		format, err := outputFormatFromFlags(viewFormat, viewJSON)
		if err != nil {
			return err
		}
		if format != formatText {
			rendered, err := renderNote(note, format)
			if err != nil {
				return err
			}
			fmt.Println(rendered)
			return nil
		}

		updatedStr := formatTimestamp(note.UpdatedAt, 19)

		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("99")).
			Padding(0, 1)
		metaStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("244")).
			Italic(true)
		divStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

		div := divStyle.Render(strings.Repeat("─", 50))
		fmt.Println(div)
		fmt.Println(titleStyle.Render(note.Title))
		fmt.Println(metaStyle.Render(fmt.Sprintf("  %s  │  %s", note.Category, updatedStr)))
		fmt.Println(div)

		renderer, err := glamour.NewTermRenderer(
			glamour.WithStandardStyle(resolveMarkdownStyle(noColorMode)),
			glamour.WithWordWrap(0),
		)
		if err != nil {
			fmt.Println(note.Content)
		} else {
			rendered, err := renderer.Render(note.Content)
			if err != nil {
				fmt.Println(note.Content)
			} else {
				fmt.Print(rendered)
			}
		}
		fmt.Println(div)

		if len(note.Images) > 0 {
			fileHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
			fmt.Println(fileHeaderStyle.Render("📎 첨부파일 목록"))

			files, err := client.GetFiles()
			urlToFile := make(map[string]api.FileRead)
			if err == nil {
				for _, f := range files {
					urlToFile[f.URL] = f
				}
			}

			fileStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
			urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("244")).Italic(true)

			for i, urlStr := range note.Images {
				f, hasFile := urlToFile[urlStr]

				filename := fileNameFromURL(urlStr)
				if hasFile {
					filename = displayFileName(f)
				}

				fmt.Printf("  %d. %s\n", i+1, fileStyle.Render(filename))

				if !noImage && hasFile && isImageFile(f) {
					ext := fileExtension(f)
					tmpPath, err := client.DownloadFileTemp(f.ID, ext)
					if err == nil {
						fmt.Println()
						if drawErr := tdraw.DrawFile(os.Stdout, tmpPath, tdraw.Options{}); drawErr != nil {
							fmt.Printf("     %s\n", urlStyle.Render(urlStr))
						}
						fmt.Println()
						_ = os.Remove(tmpPath)
					} else {
						fmt.Printf("     %s\n", urlStyle.Render(urlStr))
					}
				} else {
					fmt.Printf("     %s\n", urlStyle.Render(urlStr))
				}
			}
			fmt.Println(div)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&noImage, "no-image", "", false, "이미지 파일을 터미널에 렌더링하지 않음")
	viewCmd.Flags().StringVar(&viewFormat, "format", "", "", "출력 형식 지정 (text, json, yaml)")
	viewCmd.Flags().BoolVar(&viewJSON, "json", "", false, "JSON 형식으로 출력 (--format json과 동일)")
}
