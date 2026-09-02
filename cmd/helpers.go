package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"note_cli/api"
	"note_cli/config"
	"note_cli/local"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"
	"github.com/mattn/go-runewidth"
)

const (
	// maxAttachedFileSize 첨부 파일 개별 최대 크기
	maxAttachedFileSize = 500 * 1024 * 1024
	// defaultNoteCategory 카테고리 미지정 시 기본값
	defaultNoteCategory = "other"
	// contentStdinMarker 내용을 stdin에서 읽는다는 표식
	contentStdinMarker = "-"
)

func requireMaxArgs(args []string, max int) error {
	if len(args) > max {
		return fmt.Errorf("인자는 최대 %d개까지 지정할 수 있습니다", max)
	}
	return nil
}

func requireExactArgs(args []string, expected int) error {
	if len(args) != expected {
		return fmt.Errorf("인자는 정확히 %d개를 지정해야 합니다", expected)
	}
	return nil
}

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

func newAuthenticatedClient() (NoteStore, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, err
	}
	if cfg.Mode == "local" {
		path, err := config.DatabasePath()
		if err != nil {
			return nil, fmt.Errorf("로컬 데이터베이스 경로를 확인하지 못했습니다: %w", err)
		}
		store, err := local.Open(path)
		if err != nil {
			return nil, err
		}
		return store, nil
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

// runNoteForm 노트 제목/카테고리 입력 폼 실행 (add/edit 공용).
// askTitle/askCategory가 false인 필드는 프롬프트를 건너뛰고 전달된 값을 유지한다.
func runNoteForm(askTitle, askCategory bool, title, category *string) error {
	var fields []huh.Field
	if askTitle {
		fields = append(fields, huh.NewInput().
			Title("제목").
			Value(title).
			Validate(requiredInput("제목을 입력해야 합니다")))
	}
	if askCategory {
		fields = append(fields, huh.NewSelect[string]().
			Title("카테고리").
			Options(categorySelectOptions()...).
			Value(category))
	}
	if len(fields) == 0 {
		return nil
	}

	return huh.NewForm(huh.NewGroup(fields...)).Run()
}

// providedNoteFields 플래그로 명시적으로 지정된 노트 필드.
// nil이면 해당 필드는 지정되지 않은 것으로 간주해 기존 값/폼 입력을 유지한다.
type providedNoteFields struct {
	Title    *string
	Category *string
	Content  *string
}

// contentFlagsChanged 내용 관련 플래그(--content/--content-file)가 지정되었는지 판단
func contentFlagsChanged(content, contentFile string) bool {
	return content != "" || contentFile != ""
}

// resolveNoteContent 플래그로 지정된 내용을 해석한다.
// content가 "-"면 stdin에서, contentFile이 있으면 파일에서 읽는다.
// 두 플래그 동시 지정과 빈 결과는 에러로 처리한다 (편집기 폴백 시 세션이 멈출 수 있음).
func resolveNoteContent(content, contentFile string, stdin io.Reader) (string, error) {
	if content != "" && contentFile != "" {
		return "", fmt.Errorf("--content와 --content-file은 동시에 지정할 수 없습니다")
	}

	switch {
	case content == contentStdinMarker:
		data, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("stdin을 읽지 못했습니다: %w", err)
		}
		content = string(data)
	case contentFile != "":
		data, err := os.ReadFile(contentFile)
		if err != nil {
			return "", fmt.Errorf("내용 파일을 읽지 못했습니다 (%s): %w", contentFile, err)
		}
		content = string(data)
	}

	content = strings.TrimRight(content, "\n")
	if content == "" {
		return "", fmt.Errorf("노트 내용이 비어 있습니다")
	}

	return content, nil
}

// categoryValues 선택 옵션에서 카테고리 값 목록을 파생한다
func categoryValues() []string {
	options := categorySelectOptions()
	values := make([]string, 0, len(options))
	for _, opt := range options {
		values = append(values, opt.Value)
	}

	return values
}

// normalizeNoteCategory 카테고리 플래그 값을 검증/정규화한다. 빈 값은 기본값으로 대체.
func normalizeNoteCategory(category string) (string, error) {
	trimmed := strings.TrimSpace(strings.ToLower(category))
	if trimmed == "" {
		return defaultNoteCategory, nil
	}

	for _, v := range categoryValues() {
		if v == trimmed {
			return trimmed, nil
		}
	}

	return "", fmt.Errorf("카테고리는 %s 중 하나여야 합니다", strings.Join(categoryValues(), ", "))
}

// mergeNoteFields 기존 노트에 플래그로 지정된 필드만 덮어써 반환한다 (edit용)
func mergeNoteFields(note *api.BoardRead, provided providedNoteFields) (title, category, content string) {
	title = note.Title
	category = note.Category
	content = note.Content

	if provided.Title != nil {
		title = *provided.Title
	}
	if provided.Category != nil {
		category = *provided.Category
	}
	if provided.Content != nil {
		content = strings.TrimRight(*provided.Content, "\n")
	}

	return title, category, content
}

// uploadAttachedFiles 첨부 파일들을 순서대로 업로드하고 URL 목록 반환
func uploadAttachedFiles(client NoteStore, paths []string) ([]string, error) {
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

// resolveBoardID ID 인자가 있으면 파싱하고, 없으면 목록에서 선택하게 한다.
func resolveBoardID(client NoteStore, args []string, prompt, emptyMsg string) (id int, ok bool, err error) {
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

// resolveFileID 파일 목록에서 선택하게 하고 파일 목록도 함께 반환한다.
func resolveFileID(client NoteStore, prompt, emptyMsg string) (id int, files []api.FileRead, ok bool, err error) {
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

// buildBoardTable 노트 목록을 표 텍스트로 구성한다. 한글 정렬을 위해 runewidth로 패딩한다.
func buildBoardTable(notes []api.BoardRead) string {
	if len(notes) == 0 {
		return ""
	}

	const (
		colID    = 5
		colTitle = 42
		colCat   = 10
		colDate  = 16
		sep      = "  "
	)

	var b strings.Builder

	header := padRight("ID", colID) + sep +
		padRight("TITLE", colTitle) + sep +
		padRight("CATEGORY", colCat) + sep +
		"UPDATED"
	b.WriteString(header)
	b.WriteString("\n")

	for _, note := range notes {
		row := padRight(fmt.Sprintf("%d", note.ID), colID) + sep +
			padRight(truncateText(note.Title, 40), colTitle) + sep +
			padRight(note.Category, colCat) + sep +
			formatTimestamp(note.UpdatedAt, colDate)
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}

// padRight 표시 폭 기준으로 우측에 공백을 채운다 (전각 문자 2칸 계산).
func padRight(s string, width int) string {
	pad := width - runewidth.StringWidth(s)
	if pad <= 0 {
		return s
	}
	return s + strings.Repeat(" ", pad)
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

// truncateText 표시 폭 기준으로 문자열을 자르고 말줄임표를 붙인다.
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
