package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/tabwriter"

	"note_cli/api"
	"note_cli/config"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-runewidth"
)

const maxAttachedFileSize = 500 * 1024 * 1024

// binaryName 안내 문구에 사용할 실제 실행 파일 이름 (예: ncli)
func binaryName() string {
	return strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
}

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
		return nil, fmt.Errorf("로그인이 필요합니다. '%s login'을 먼저 실행하세요", binaryName())
	}

	return api.NewClient(cfg), nil
}

// requiredInput 빈 입력을 거부하는 huh 검증 함수 생성
func requiredInput(message string) func(string) error {
	return func(str string) error {
		if str == "" {
			return errors.New(message)
		}
		return nil
	}
}

// runNoteForm 노트 제목/카테고리 입력 폼 실행 (add/edit 공용)
func runNoteForm(title, category *string) error {
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("제목").
				Value(title).
				Validate(requiredInput("제목을 입력해야 합니다")),
			huh.NewSelect[string]().
				Title("카테고리").
				Options(categorySelectOptions()...).
				Value(category),
		),
	).Run()
}

// uploadAttachedFiles 첨부 파일들을 순서대로 업로드하고 URL 목록 반환
func uploadAttachedFiles(client *api.Client, paths []string) ([]string, error) {
	var urls []string
	for _, path := range paths {
		fmt.Printf("파일 첨부 중: %s\n", path)
		uploaded, err := client.UploadFile(path)
		if err != nil {
			return nil, fmt.Errorf("업로드 실패 (%s): %w", path, err)
		}
		urls = append(urls, uploaded.URL)
		fmt.Println("업로드 완료!")
	}

	return urls, nil
}

// resolveBoardID args에 ID가 있으면 파싱하고, 없으면 목록에서 선택하게 한다.
// ok=false이고 err=nil이면 안내를 이미 출력했으므로 호출자는 정상 종료하면 된다.
func resolveBoardID(client *api.Client, args []string, prompt, emptyMsg string) (id int, ok bool, err error) {
	if len(args) == 1 {
		id, err = parseIDArg(args)
		return id, err == nil, err
	}

	notes, err := client.GetBoards()
	if err != nil {
		return 0, false, fmt.Errorf("노트 목록을 불러오지 못했습니다: %w", err)
	}
	if len(notes) == 0 {
		fmt.Println(emptyMsg)
		return 0, false, nil
	}

	id, err = selectBoardID(prompt, notes)
	if err != nil {
		fmt.Println("취소되었습니다.")
		return 0, false, nil
	}

	return id, true, nil
}

// resolveFileID 파일 목록에서 선택하게 한다. 선택 후 파일 정보 조회용으로
// 파일 목록도 함께 반환한다. ok=false, err=nil이면 정상 종료하면 된다.
func resolveFileID(client *api.Client, prompt, emptyMsg string) (id int, files []api.FileRead, ok bool, err error) {
	files, err = client.GetFiles()
	if err != nil {
		return 0, nil, false, fmt.Errorf("파일 목록을 불러오지 못했습니다: %w", err)
	}
	if len(files) == 0 {
		fmt.Println(emptyMsg)
		return 0, nil, false, nil
	}

	// 노트 제목은 표시용 부가 정보이므로 조회에 실패해도 계속 진행
	boardTitles := map[string]string{}
	if boards, boardsErr := client.GetBoards(); boardsErr == nil {
		boardTitles = boardTitleByFileName(boards)
	}

	id, err = selectFileID(prompt, files, boardTitles)
	if err != nil {
		fmt.Println("취소되었습니다.")
		return 0, nil, false, nil
	}

	return id, files, true, nil
}

// printBoardTable 노트 목록을 표 형태로 출력 (list/search 공용)
func printBoardTable(notes []api.BoardRead) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tUPDATED")
	for _, note := range notes {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", note.ID, truncateText(note.Title, 40), note.Category, formatTimestamp(note.UpdatedAt, 16))
	}
	w.Flush()
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

// truncateText 표시 폭(전각 문자 2칸) 기준으로 문자열을 자르고 말줄임표를 붙인다.
// 바이트 단위 슬라이스는 한글 등 멀티바이트 문자를 중간에서 깨뜨리므로 사용하지 않는다.
func truncateText(value string, max int) string {
	if max <= 0 || runewidth.StringWidth(value) <= max {
		return value
	}
	if max <= 3 {
		return runewidth.Truncate(value, max, "")
	}

	return runewidth.Truncate(value, max, "...")
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
