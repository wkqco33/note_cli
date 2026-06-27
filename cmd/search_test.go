package cmd

import (
	"testing"

	"note_cli/api"
)

func TestFilterBoardsByTitle(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "회의 메모", Content: "내용A"},
		{ID: 2, Title: "아이디어 정리", Content: "회의록"},
	}
	got := filterBoards(notes, nil, "회의", "", "")
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("expected only note 1, got %#v", got)
	}
}

func TestFilterBoardsByContent(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "A", Content: "hello world"},
		{ID: 2, Title: "B", Content: "goodbye"},
	}
	got := filterBoards(notes, nil, "", "world", "")
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("expected only note 1, got %#v", got)
	}
}

func TestFilterBoardsCaseInsensitive(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "GoLang Tips", Content: "x"},
	}
	got := filterBoards(notes, nil, "golang", "", "")
	if len(got) != 1 {
		t.Fatalf("expected case-insensitive match, got %#v", got)
	}
}

func TestFilterBoardsByFile(t *testing.T) {
	files := []api.FileRead{
		{ID: 10, URL: "http://h/family.png", OriginalFilename: "family.png"},
		{ID: 11, URL: "http://h/work_doc.pdf", OriginalFilename: "work_doc.pdf"},
	}
	notes := []api.BoardRead{
		{ID: 1, Title: "n1", Content: "c", Images: []string{"http://h/family.png"}},
		{ID: 2, Title: "n2", Content: "c", Images: []string{"http://h/work_doc.pdf"}},
		{ID: 3, Title: "n3", Content: "c", Images: []string{"http://h/other.png"}},
	}
	got := filterBoards(notes, files, "", "", "family")
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("expected only note 1, got %#v", got)
	}
}

func TestFilterBoardsByFileUsesURLFallback(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "n1", Content: "c", Images: []string{"http://h/screenshot_xyz.png"}},
	}
	got := filterBoards(notes, nil, "", "", "screenshot")
	if len(got) != 1 {
		t.Fatalf("expected url fallback match, got %#v", got)
	}
}

func TestFilterBoardsAllEmptyReturnsAll(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "a"},
		{ID: 2, Title: "b"},
	}
	got := filterBoards(notes, nil, "", "", "")
	if len(got) != 2 {
		t.Fatalf("expected all notes returned, got %d", len(got))
	}
}

func TestFilterBoardsCombinedConditionsAllMustMatch(t *testing.T) {
	files := []api.FileRead{{URL: "http://h/img.png", OriginalFilename: "img.png"}}
	notes := []api.BoardRead{
		{ID: 1, Title: "meeting", Content: "agenda", Images: []string{"http://h/img.png"}},
		{ID: 2, Title: "meeting", Content: "other", Images: []string{"http://h/img.png"}},
	}
	got := filterBoards(notes, files, "meeting", "agenda", "img")
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("expected only note 1 with all conditions met, got %#v", got)
	}
}

func TestBuildURLToFilename(t *testing.T) {
	files := []api.FileRead{
		{URL: "u1", Filename: "a.png", OriginalFilename: "alpha.png"},
		{URL: "u2", Filename: "b.png"},
	}
	m := buildURLToFilename(files)
	if m["u1"] != "alpha.png" {
		t.Fatalf("expected original filename, got %q", m["u1"])
	}
	if m["u2"] != "b.png" {
		t.Fatalf("expected fallback to filename, got %q", m["u2"])
	}
}
