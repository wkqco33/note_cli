package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"note_cli/api"
	"note_cli/config"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
)

const maxAttachedFileSize = 500 * 1024 * 1024

func loadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("설정을 불러오지 못했습니다: %w", err)
	}

	return cfg, nil
}

func newClient() (*api.Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return api.NewClient(cfg), nil
}

func newAuthenticatedClient() (*api.Client, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if cfg.AccessToken == "" {
		return nil, fmt.Errorf("로그인이 필요합니다. 'note_cli login'을 먼저 실행하세요")
	}

	return api.NewClient(cfg), nil
}

func parseIDArg(args []string) (int, error) {
	id, err := strconv.Atoi(args[0])
	if err != nil {
		return 0, fmt.Errorf("ID는 정수여야 합니다")
	}

	return id, nil
}

func validateAttachedFiles(paths []string) error {
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("파일을 찾을 수 없습니다: %s", path)
		}
		if info.Size() > maxAttachedFileSize {
			return fmt.Errorf("500MB 제한을 초과한 파일이 있습니다: %s", path)
		}
	}

	return nil
}

func truncateText(value string, max int) string {
	if max <= 0 || len(value) <= max {
		return value
	}
	if max <= 3 {
		return value[:max]
	}

	return value[:max-3] + "..."
}

func formatTimestamp(value string, length int) string {
	if len(value) < length {
		return value
	}

	return strings.Replace(value[:length], "T", " ", 1)
}

func fileNameFromURL(rawURL string) string {
	idx := strings.LastIndex(rawURL, "/")
	if idx == -1 {
		return rawURL
	}

	return rawURL[idx+1:]
}

func displayFileName(file api.FileRead) string {
	if file.OriginalFilename != "" {
		return file.OriginalFilename
	}

	return file.Filename
}

func fileExtension(file api.FileRead) string {
	ext := filepath.Ext(strings.ToLower(file.OriginalFilename))
	if ext != "" {
		return ext
	}

	return filepath.Ext(strings.ToLower(file.Filename))
}

func boardOptions(boards []api.BoardRead) []huh.Option[int] {
	options := make([]huh.Option[int], 0, len(boards))
	for _, board := range boards {
		label := fmt.Sprintf("[%d] %s", board.ID, truncateText(board.Title, 40))
		options = append(options, huh.NewOption(label, board.ID))
	}

	return options
}

func selectBoardID(title string, boards []api.BoardRead) (int, error) {
	var id int
	err := huh.NewSelect[int]().
		Title(title).
		Options(boardOptions(boards)...).
		Value(&id).
		Run()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func boardTitleByFileName(boards []api.BoardRead) map[string]string {
	titles := make(map[string]string, len(boards))
	for _, board := range boards {
		for _, imageURL := range board.Images {
			titles[fileNameFromURL(imageURL)] = board.Title
		}
	}

	return titles
}

func fileOptions(files []api.FileRead, boardTitles map[string]string) []huh.Option[int] {
	options := make([]huh.Option[int], 0, len(files))
	for _, file := range files {
		label := fmt.Sprintf("[%d] %s", file.ID, displayFileName(file))
		noteTitle := truncateText(boardTitles[file.Filename], 15)
		if noteTitle != "" {
			label += fmt.Sprintf(" (노트: %s, 크기: %s)", noteTitle, humanize.Bytes(uint64(file.FileSize)))
		} else {
			label += fmt.Sprintf(" (크기: %s)", humanize.Bytes(uint64(file.FileSize)))
		}
		options = append(options, huh.NewOption(label, file.ID))
	}

	return options
}

func selectFileID(title string, files []api.FileRead, boardTitles map[string]string) (int, error) {
	var id int
	err := huh.NewSelect[int]().
		Title(title).
		Options(fileOptions(files, boardTitles)...).
		Value(&id).
		Run()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func findFileByID(files []api.FileRead, id int) (api.FileRead, bool) {
	for _, file := range files {
		if file.ID == id {
			return file, true
		}
	}

	return api.FileRead{}, false
}

func categorySelectOptions() []huh.Option[string] {
	return []huh.Option[string]{
		huh.NewOption("Work", "work"),
		huh.NewOption("Personal", "personal"),
		huh.NewOption("Idea", "idea"),
		huh.NewOption("Other", "other"),
	}
}
