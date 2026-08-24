package cmd

import (
	"io"

	"note_cli/api"
)

// NoteStore 노트와 첨부파일을 저장하고 조회하는 공통 인터페이스.
// 원격 API와 로컬 SQLite 저장소가 동일한 명령어에서 사용되도록 한다.
type NoteStore interface {
	Close() error
	GetBoards() ([]api.BoardRead, error)
	CreateBoard(board api.BoardCreate) (*api.BoardRead, error)
	GetBoard(id int) (*api.BoardRead, error)
	UpdateBoard(id int, update api.BoardUpdate) (*api.BoardRead, error)
	DeleteBoard(id int) error

	GetFiles() ([]api.FileRead, error)
	UploadFile(filePath string) (*api.FileRead, error)
	DownloadFile(fileID int, destDir string, filename string) (string, error)
	DownloadFileTemp(fileID int, ext string) (string, error)
	DeleteFile(id int) error
	GetFileStream(fileID int) (io.ReadCloser, int64, error)
}
