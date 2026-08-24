// Package local은 로컬 SQLite 기반 노트 저장소를 제공한다.
package local

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"time"

	"note_cli/api"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS notes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    category TEXT NOT NULL,
    images TEXT NOT NULL DEFAULT '[]',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    content_type TEXT NOT NULL,
    local_path TEXT NOT NULL,
    created_at TEXT NOT NULL
);
`

// Store 로컬 노트 데이터베이스.
type Store struct {
	db       *sql.DB
	filesDir string
}

// Open 지정된 경로의 SQLite 데이터베이스를 열고 스키마를 초기화한다.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("로컬 데이터베이스를 열지 못했습니다: %w", err)
	}

	filesDir := filepath.Join(filepath.Dir(path), "files")
	if err := os.MkdirAll(filesDir, 0700); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("로컬 첨부파일 디렉토리를 생성하지 못했습니다: %w", err)
	}
	store := &Store{db: db, filesDir: filesDir}
	if err := store.initialize(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) initialize() error {
	statements := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
		schema,
	}
	for _, statement := range statements {
		if _, err := s.db.Exec(statement); err != nil {
			return fmt.Errorf("로컬 데이터베이스를 초기화하지 못했습니다: %w", err)
		}
	}
	return nil
}

// Close 데이터베이스 연결을 닫는다.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// GetBoards 현재 저장된 모든 노트를 수정 시간 역순으로 반환한다.
func (s *Store) GetBoards() ([]api.BoardRead, error) {
	rows, err := s.db.Query(`
		SELECT id, title, content, category, images, created_at, updated_at
		FROM notes ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("로컬 노트 목록을 조회하지 못했습니다: %w", err)
	}
	defer func() { _ = rows.Close() }()

	boards := make([]api.BoardRead, 0)
	for rows.Next() {
		board, err := scanBoard(rows)
		if err != nil {
			return nil, err
		}
		boards = append(boards, board)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("로컬 노트 목록을 읽지 못했습니다: %w", err)
	}
	return boards, nil
}

// CreateBoard 새 노트를 저장한다.
func (s *Store) CreateBoard(board api.BoardCreate) (*api.BoardRead, error) {
	images, err := json.Marshal(board.Images)
	if err != nil {
		return nil, fmt.Errorf("첨부파일 목록을 직렬화하지 못했습니다: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.Exec(`
		INSERT INTO notes (title, content, category, images, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`, board.Title, board.Content, board.Category, string(images), now, now)
	if err != nil {
		return nil, fmt.Errorf("로컬 노트를 생성하지 못했습니다: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("생성된 노트 ID를 가져오지 못했습니다: %w", err)
	}
	return s.GetBoard(int(id))
}

// GetBoard ID로 노트를 조회한다.
func (s *Store) GetBoard(id int) (*api.BoardRead, error) {
	row := s.db.QueryRow(`
		SELECT id, title, content, category, images, created_at, updated_at
		FROM notes WHERE id = ?`, id)
	board, err := scanBoard(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("노트를 찾을 수 없습니다: %d", id)
		}
		return nil, err
	}
	return &board, nil
}

// UpdateBoard 지정된 필드만 노트를 수정한다.
func (s *Store) UpdateBoard(id int, update api.BoardUpdate) (*api.BoardRead, error) {
	current, err := s.GetBoard(id)
	if err != nil {
		return nil, err
	}
	if update.Title != nil {
		current.Title = *update.Title
	}
	if update.Content != nil {
		current.Content = *update.Content
	}
	if update.Category != nil {
		current.Category = *update.Category
	}
	if update.Images != nil {
		current.Images = *update.Images
	}
	images, err := json.Marshal(current.Images)
	if err != nil {
		return nil, fmt.Errorf("첨부파일 목록을 직렬화하지 못했습니다: %w", err)
	}
	updatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.Exec(`
		UPDATE notes SET title = ?, content = ?, category = ?, images = ?, updated_at = ?
		WHERE id = ?`, current.Title, current.Content, current.Category, string(images), updatedAt, id)
	if err != nil {
		return nil, fmt.Errorf("로컬 노트를 수정하지 못했습니다: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil, fmt.Errorf("노트를 찾을 수 없습니다: %d", id)
	}
	return s.GetBoard(id)
}

// DeleteBoard ID로 노트를 삭제한다.
func (s *Store) DeleteBoard(id int) error {
	result, err := s.db.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("로컬 노트를 삭제하지 못했습니다: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("삭제된 노트 수를 확인하지 못했습니다: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("노트를 찾을 수 없습니다: %d", id)
	}
	return nil
}

// GetFiles 현재 저장된 모든 첨부파일을 생성 시간 역순으로 반환한다.
func (s *Store) GetFiles() ([]api.FileRead, error) {
	rows, err := s.db.Query(`
		SELECT id, filename, original_filename, file_size, content_type, local_path, created_at
		FROM files ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("로컬 첨부파일 목록을 조회하지 못했습니다: %w", err)
	}
	defer func() { _ = rows.Close() }()

	files := make([]api.FileRead, 0)
	for rows.Next() {
		localFile, err := scanFile(rows)
		if err != nil {
			return nil, err
		}
		files = append(files, localFile.FileRead)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("로컬 첨부파일 목록을 읽지 못했습니다: %w", err)
	}
	return files, nil
}

// UploadFile 로컬 파일을 데이터 디렉토리에 복사하고 메타데이터를 저장한다.
func (s *Store) UploadFile(filePath string) (*api.FileRead, error) {
	source, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("파일을 열지 못했습니다: %w", err)
	}
	defer func() { _ = source.Close() }()

	stat, err := source.Stat()
	if err != nil {
		return nil, fmt.Errorf("파일 정보를 가져오지 못했습니다: %w", err)
	}
	if !stat.Mode().IsRegular() {
		return nil, fmt.Errorf("일반 파일만 첨부할 수 있습니다: %s", filePath)
	}
	originalName := filepath.Base(filePath)
	if originalName == "." || originalName == string(filepath.Separator) || originalName == "" {
		return nil, fmt.Errorf("유효하지 않은 파일명입니다: %s", filePath)
	}
	contentType := mime.TypeByExtension(filepath.Ext(originalName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	result, err := s.db.Exec(`
		INSERT INTO files (filename, original_filename, file_size, content_type, local_path, created_at)
		VALUES (?, ?, ?, ?, '', ?)`, originalName, originalName, stat.Size(), contentType, now)
	if err != nil {
		return nil, fmt.Errorf("첨부파일 정보를 저장하지 못했습니다: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("생성된 첨부파일 ID를 가져오지 못했습니다: %w", err)
	}
	destPath := filepath.Join(s.filesDir, fmt.Sprintf("%d_%s", id, originalName))
	dest, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		_, _ = s.db.Exec("DELETE FROM files WHERE id = ?", id)
		return nil, fmt.Errorf("첨부파일을 저장하지 못했습니다: %w", err)
	}
	_, copyErr := io.Copy(dest, source)
	closeErr := dest.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(destPath)
		_, _ = s.db.Exec("DELETE FROM files WHERE id = ?", id)
		if copyErr != nil {
			return nil, fmt.Errorf("첨부파일을 복사하지 못했습니다: %w", copyErr)
		}
		return nil, fmt.Errorf("첨부파일을 닫지 못했습니다: %w", closeErr)
	}

	if _, err := s.db.Exec("UPDATE files SET local_path = ? WHERE id = ?", destPath, id); err != nil {
		_ = os.Remove(destPath)
		_, _ = s.db.Exec("DELETE FROM files WHERE id = ?", id)
		return nil, fmt.Errorf("첨부파일 경로를 저장하지 못했습니다: %w", err)
	}
	localFile, err := s.getFile(int(id))
	if err != nil {
		return nil, err
	}
	return &localFile.FileRead, nil
}

func (s *Store) getFile(id int) (*localFile, error) {
	row := s.db.QueryRow(`
		SELECT id, filename, original_filename, file_size, content_type, local_path, created_at
		FROM files WHERE id = ?`, id)
	file, err := scanFile(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("첨부파일을 찾을 수 없습니다: %d", id)
		}
		return nil, err
	}
	return &file, nil
}

// DeleteFile 첨부파일과 노트의 해당 첨부파일 참조를 함께 삭제한다.
func (s *Store) DeleteFile(id int) error {
	file, err := s.getFile(id)
	if err != nil {
		return err
	}
	if err := os.Remove(file.localPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("첨부파일을 삭제하지 못했습니다: %w", err)
	}
	if _, err := s.db.Exec("DELETE FROM files WHERE id = ?", id); err != nil {
		return fmt.Errorf("첨부파일 정보를 삭제하지 못했습니다: %w", err)
	}
	return nil
}

// DownloadFile 첨부파일을 지정된 디렉토리에 복사한다.
func (s *Store) DownloadFile(fileID int, destDir, filename string) (string, error) {
	file, err := s.getFile(fileID)
	if err != nil {
		return "", err
	}
	if filename == "" {
		filename = file.OriginalFilename
	}
	filename = filepath.Base(filepath.Clean(filename))
	if filename == "." || filename == "" {
		filename = fmt.Sprintf("downloaded_file_%d", fileID)
	}
	destPath := filepath.Join(destDir, filename)
	source, err := os.Open(file.localPath)
	if err != nil {
		return "", fmt.Errorf("첨부파일을 열지 못했습니다: %w", err)
	}
	defer func() { _ = source.Close() }()
	dest, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("다운로드 파일을 생성하지 못했습니다: %w", err)
	}
	if _, err := io.Copy(dest, source); err != nil {
		_ = dest.Close()
		return "", fmt.Errorf("첨부파일을 복사하지 못했습니다: %w", err)
	}
	if err := dest.Close(); err != nil {
		return "", fmt.Errorf("다운로드 파일을 닫지 못했습니다: %w", err)
	}
	return destPath, nil
}

// DownloadFileTemp 첨부파일을 임시 파일로 복사한다.
func (s *Store) DownloadFileTemp(fileID int, ext string) (string, error) {
	file, err := s.getFile(fileID)
	if err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp("", fmt.Sprintf("note_img_*%s", ext))
	if err != nil {
		return "", fmt.Errorf("임시 파일을 생성하지 못했습니다: %w", err)
	}
	source, err := os.Open(file.localPath)
	if err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("첨부파일을 열지 못했습니다: %w", err)
	}
	_, copyErr := io.Copy(tmp, source)
	_ = source.Close()
	closeErr := tmp.Close()
	if copyErr != nil || closeErr != nil {
		_ = os.Remove(tmp.Name())
		if copyErr != nil {
			return "", fmt.Errorf("임시 파일을 쓰지 못했습니다: %w", copyErr)
		}
		return "", fmt.Errorf("임시 파일을 닫지 못했습니다: %w", closeErr)
	}
	return tmp.Name(), nil
}

// GetFileStream 첨부파일 스트림과 크기를 반환한다.
func (s *Store) GetFileStream(fileID int) (io.ReadCloser, int64, error) {
	file, err := s.getFile(fileID)
	if err != nil {
		return nil, 0, err
	}
	stream, err := os.Open(file.localPath)
	if err != nil {
		return nil, 0, fmt.Errorf("첨부파일을 열지 못했습니다: %w", err)
	}
	return stream, int64(file.FileSize), nil
}

type localFile struct {
	api.FileRead
	localPath string
}

type fileRowScanner interface {
	Scan(dest ...any) error
}

func scanFile(row fileRowScanner) (localFile, error) {
	var file localFile
	var localPath string
	if err := row.Scan(&file.ID, &file.Filename, &file.OriginalFilename,
		&file.FileSize, &file.ContentType, &localPath, &file.CreatedAt); err != nil {
		return localFile{}, fmt.Errorf("로컬 첨부파일을 읽지 못했습니다: %w", err)
	}
	file.localPath = localPath
	file.URL = "file://" + filepath.ToSlash(localPath)
	return file, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanBoard(row rowScanner) (api.BoardRead, error) {
	var board api.BoardRead
	var images string
	if err := row.Scan(&board.ID, &board.Title, &board.Content, &board.Category,
		&images, &board.CreatedAt, &board.UpdatedAt); err != nil {
		return api.BoardRead{}, fmt.Errorf("로컬 노트를 읽지 못했습니다: %w", err)
	}
	if err := json.Unmarshal([]byte(images), &board.Images); err != nil {
		return api.BoardRead{}, fmt.Errorf("첨부파일 목록을 역직렬화하지 못했습니다: %w", err)
	}
	return board, nil
}
