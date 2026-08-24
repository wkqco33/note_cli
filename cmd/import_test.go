package cmd

import (
	"archive/zip"
	"path/filepath"
	"testing"

	"note_cli/api"
)

func TestRemapImageURLs(t *testing.T) {
	mapping := map[string]string{"old1": "new1", "old2": "new2"}
	got := remapImageURLs([]string{"old1", "old3", "old2"}, mapping)
	want := []string{"new1", "old3", "new2"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("at %d: expected %q, got %q", i, want[i], got[i])
		}
	}
}

func TestRemapImageURLsEmpty(t *testing.T) {
	if got := remapImageURLs(nil, nil); got != nil {
		t.Fatalf("expected nil, got %#v", got)
	}
}

func TestRebuildBoardsForRestore(t *testing.T) {
	boards := []api.BoardRead{
		{Title: "t1", Content: "c1", Category: "work", Images: []string{"old1", "old2"}},
		{Title: "t2", Content: "c2", Category: "idea", Images: []string{"unknown"}},
	}
	mapping := map[string]string{"old1": "new1", "old2": "new2"}
	got := rebuildBoardsForRestore(boards, mapping)

	if len(got) != 2 {
		t.Fatalf("expected 2 boards, got %d", len(got))
	}
	if got[0].Title != "t1" || got[0].Category != "work" {
		t.Fatalf("unexpected board 1: %#v", got[0])
	}
	if len(got[0].Images) != 2 || got[0].Images[0] != "new1" || got[0].Images[1] != "new2" {
		t.Fatalf("expected remapped images, got %#v", got[0].Images)
	}
	if got[1].Images[0] != "unknown" {
		t.Fatalf("expected unmapped URL preserved, got %q", got[1].Images[0])
	}
}

func TestRebuildBoardsForRestoreEmptyImages(t *testing.T) {
	boards := []api.BoardRead{{Title: "t", Content: "c", Category: "w"}}
	got := rebuildBoardsForRestore(boards, nil)
	if len(got) != 1 || got[0].Images != nil {
		t.Fatalf("unexpected: %#v", got)
	}
}

func TestFindBackupEntries(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "backup.zip")
	if err := createTestZip(zipPath, map[string]string{
		"notes.json":    "[]",
		"files.json":    "[]",
		"files/1_a.txt": "x",
		"unrelated.txt": "y",
	}); err != nil {
		t.Fatalf("create zip: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer func() {
		_ = r.Close()
	}()

	notesFile, filesFile := findBackupEntries(r.File)
	if notesFile == nil || filesFile == nil {
		t.Fatalf("expected both notes.json and files.json to be found")
	}
}

func TestFindBackupEntriesMissing(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "backup.zip")
	if err := createTestZip(zipPath, map[string]string{
		"notes.json": "[]",
	}); err != nil {
		t.Fatalf("create zip: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer func() {
		_ = r.Close()
	}()

	notesFile, filesFile := findBackupEntries(r.File)
	if notesFile == nil {
		t.Fatal("expected notes.json to be found")
	}
	if filesFile != nil {
		t.Fatal("expected files.json to be missing")
	}
}

func TestFindBackupFileEntry(t *testing.T) {
	tmp := t.TempDir()
	zipPath := filepath.Join(tmp, "backup.zip")
	if err := createTestZip(zipPath, map[string]string{
		"files/1_a.txt": "x",
		"files/2_b.txt": "y",
	}); err != nil {
		t.Fatalf("create zip: %v", err)
	}

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	defer r.Close()

	if findBackupFileEntry(r.File, "files/1_a.txt") == nil {
		t.Fatal("expected entry found")
	}
	if findBackupFileEntry(r.File, "files/missing.txt") != nil {
		t.Fatal("expected nil for missing entry")
	}
}

func TestExceedsRestoreFileSize(t *testing.T) {
	tests := []struct {
		name string
		size uint64
		want bool
	}{
		{"limit 이하", maxRestoreFileSize, false},
		{"limit 초과", maxRestoreFileSize + 1, true},
		{"빈 파일", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &zip.File{FileHeader: zip.FileHeader{UncompressedSize64: tt.size}}
			if got := exceedsRestoreFileSize(f); got != tt.want {
				t.Fatalf("exceedsRestoreFileSize(%d) = %v, want %v", tt.size, got, tt.want)
			}
		})
	}
}
