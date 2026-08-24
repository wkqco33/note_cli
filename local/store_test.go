package local

import (
	"os"
	"path/filepath"
	"testing"

	"note_cli/api"
)

func TestStoreBoardCRUD(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "notes.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() { _ = store.Close() }()

	created, err := store.CreateBoard(api.BoardCreate{
		Title:    "첫 노트",
		Content:  "내용",
		Category: "personal",
		Images:   []string{"file:///tmp/image.png"},
	})
	if err != nil {
		t.Fatalf("CreateBoard() error = %v", err)
	}
	if created.ID != 1 {
		t.Fatalf("created ID = %d, want 1", created.ID)
	}
	if len(created.Images) != 1 || created.Images[0] != "file:///tmp/image.png" {
		t.Fatalf("created images = %#v", created.Images)
	}

	got, err := store.GetBoard(created.ID)
	if err != nil {
		t.Fatalf("GetBoard() error = %v", err)
	}
	if got.Title != "첫 노트" || got.Content != "내용" {
		t.Fatalf("GetBoard() = %#v", got)
	}

	newTitle := "수정한 노트"
	updated, err := store.UpdateBoard(created.ID, api.BoardUpdate{Title: &newTitle})
	if err != nil {
		t.Fatalf("UpdateBoard() error = %v", err)
	}
	if updated.Title != newTitle || updated.Content != "내용" {
		t.Fatalf("UpdateBoard() = %#v", updated)
	}

	boards, err := store.GetBoards()
	if err != nil {
		t.Fatalf("GetBoards() error = %v", err)
	}
	if len(boards) != 1 || boards[0].ID != created.ID {
		t.Fatalf("GetBoards() = %#v", boards)
	}

	if err := store.DeleteBoard(created.ID); err != nil {
		t.Fatalf("DeleteBoard() error = %v", err)
	}
	boards, err = store.GetBoards()
	if err != nil {
		t.Fatalf("GetBoards() after delete error = %v", err)
	}
	if len(boards) != 0 {
		t.Fatalf("GetBoards() after delete = %#v", boards)
	}
}

func TestStoreRejectsMissingBoard(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "notes.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() { _ = store.Close() }()

	if _, err := store.GetBoard(999); err == nil {
		t.Fatal("GetBoard() error = nil, want error")
	}
}

func TestStoreFileLifecycle(t *testing.T) {
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "메모.txt")
	content := []byte("첨부파일 내용")
	if err := os.WriteFile(sourcePath, content, 0600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	store, err := Open(filepath.Join(dir, "notes.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer func() { _ = store.Close() }()

	file, err := store.UploadFile(sourcePath)
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}
	if file.ID != 1 || file.OriginalFilename != "메모.txt" || file.FileSize != len(content) {
		t.Fatalf("UploadFile() = %#v", file)
	}
	if file.URL == "" {
		t.Fatal("UploadFile() URL is empty")
	}

	destDir := filepath.Join(dir, "download")
	if err := os.Mkdir(destDir, 0700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	destPath, err := store.DownloadFile(file.ID, destDir, "copy.txt")
	if err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}
	got, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("downloaded content = %q, want %q", got, content)
	}

	if err := store.DeleteFile(file.ID); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}
	files, err := store.GetFiles()
	if err != nil {
		t.Fatalf("GetFiles() error = %v", err)
	}
	if len(files) != 0 {
		t.Fatalf("GetFiles() after delete = %#v", files)
	}
}
