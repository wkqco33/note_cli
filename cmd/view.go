package cmd

import (
	"fmt"
	"note_cli/api"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/wkqco33/tdraw"
	"github.com/wkqco33/wcli"
)

var noImage bool

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
		note, err := client.GetBoard(id)
		if err != nil {
			return fmt.Errorf("노트를 불러오지 못했습니다: %w", err)
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
			glamour.WithAutoStyle(),
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
}
