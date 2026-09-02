package cmd

import (
	"io"

	"note_cli/api"
	"note_cli/embedding"
)

// NoteStore 원격 API와 로컬 SQLite 저장소가 공유하는 공통 인터페이스.
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

type embeddingStore interface {
	GetEmbeddings() ([]embedding.Record, error)
	SaveEmbedding(embedding.Record) error
}
