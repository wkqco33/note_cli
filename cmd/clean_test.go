package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanCacheTargets(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)

	// 대상 패턴과 무관한 파일
	others := []string{"keep.txt", "note_img.png", "note_imgsomething"}
	for _, name := range others {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte("x"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	// 매칭 대상
	target := filepath.Join(tmp, "note_img_20240101.png")
	if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	matches, err := findCleanCacheFiles(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 || matches[0] != target {
		t.Fatalf("expected only %q, got %v", target, matches)
	}
}

func TestCleanCacheNone(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	if err := os.WriteFile(filepath.Join(tmp, "note.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	matches, err := findCleanCacheFiles(tmp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected no matches, got %v", matches)
	}
}

func TestCleanCacheStats(t *testing.T) {
	tmp := t.TempDir()
	paths := []string{
		filepath.Join(tmp, "a.png"),
		filepath.Join(tmp, "b.png"),
	}
	_ = os.WriteFile(paths[0], make([]byte, 2048), 0644)
	_ = os.WriteFile(paths[1], make([]byte, 1024), 0644)

	count, size := cleanCacheFiles(paths, t.TempDir()) // missing dir → 삭제 실패는 무시
	if count != 0 {
		t.Fatalf("expected 0 deleted (temp dir differs), got %d", count)
	}

	count, size = cleanCacheFiles(paths, tmp)
	if count != 2 {
		t.Fatalf("expected 2 deleted, got %d", count)
	}
	if size != 3072 {
		t.Fatalf("expected total size 3072, got %d", size)
	}
	for _, p := range paths {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Fatalf("expected %s removed", p)
		}
	}
}
