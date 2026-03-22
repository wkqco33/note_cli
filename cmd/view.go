package cmd

import (
	"fmt"
	"note_cli/api"
	"note_cli/config"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/wkqco/tdraw"
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

var viewCmd = &cobra.Command{
	Use:   "view [id]",
	Short: "ID로 노트 조회",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil || cfg.AccessToken == "" {
			fmt.Println("Please login first using 'note_cli login'")
			return
		}

		client := api.NewClient(cfg)

		var id int
		if len(args) == 1 {
			id, err = strconv.Atoi(args[0])
			if err != nil {
				fmt.Println("Invalid ID: must be an integer")
				return
			}
		} else {
			notes, err := client.GetBoards()
			if err != nil {
				fmt.Printf("Failed to get notes: %v\n", err)
				return
			}
			if len(notes) == 0 {
				fmt.Println("No notes found to view.")
				return
			}

			var options []huh.Option[int]
			for _, note := range notes {
				title := note.Title
				if len(title) > 40 {
					title = title[:37] + "..."
				}
				label := fmt.Sprintf("[%d] %s", note.ID, title)
				options = append(options, huh.NewOption(label, note.ID))
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[int]().
						Title("조회할 노트를 선택하세요").
						Options(options...).
						Value(&id),
				),
			)
			if err := form.Run(); err != nil {
				fmt.Println("취소되었습니다.")
				return
			}
		}
		note, err := client.GetBoard(id)
		if err != nil {
			fmt.Printf("Failed to get note: %v\n", err)
			return
		}

		updatedStr := note.UpdatedAt
		if len(updatedStr) >= 19 {
			updatedStr = strings.Replace(updatedStr[:19], "T", " ", 1)
		}

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

		rendered, err := glamour.Render(note.Content, "dark")
		if err != nil {
			fmt.Println(note.Content)
		} else {
			fmt.Print(rendered)
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

				filename := ""
				if hasFile {
					filename = f.OriginalFilename
					if filename == "" {
						filename = f.Filename
					}
				}
				if filename == "" {
					idx := strings.LastIndex(urlStr, "/")
					filename = urlStr
					if idx != -1 {
						filename = urlStr[idx+1:]
					}
				}

				fmt.Printf("  %d. %s\n", i+1, fileStyle.Render(filename))

				if !noImage && hasFile && isImageFile(f) {
					ext := filepath.Ext(strings.ToLower(f.OriginalFilename))
					if ext == "" {
						ext = filepath.Ext(strings.ToLower(f.Filename))
					}
					tmpPath, err := client.DownloadFileTemp(f.ID, ext)
					if err == nil {
						fmt.Println()
						if drawErr := tdraw.DrawFile(os.Stdout, tmpPath, tdraw.Options{}); drawErr != nil {
							fmt.Printf("     %s\n", urlStyle.Render(urlStr))
						}
						fmt.Println()
						os.Remove(tmpPath)
					} else {
						fmt.Printf("     %s\n", urlStyle.Render(urlStr))
					}
				} else {
					fmt.Printf("     %s\n", urlStyle.Render(urlStr))
				}
			}
			fmt.Println(div)
		}
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&noImage, "no-image", false, "이미지 파일을 터미널에 렌더링하지 않음")
}
