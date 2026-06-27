package cmd

import (
	"encoding/json"
	"testing"
	"time"

	"note_cli/api"
)

func TestMarshalBackupJSON(t *testing.T) {
	boards := []api.BoardRead{{ID: 1, Title: "t"}}
	data, err := marshalBackupJSON(boards)
	if err != nil {
		t.Fatalf("marshalBackupJSON: %v", err)
	}

	var got []api.BoardRead
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(got) != 1 || got[0].ID != 1 || got[0].Title != "t" {
		t.Fatalf("roundtrip mismatch: %#v", got)
	}

	// 들여쓰기 포함 여부 확인
	if !contains(string(data), "  ") {
		t.Fatal("expected indented JSON output")
	}
}

func TestDefaultExportPath(t *testing.T) {
	ts := time.Date(2026, 3, 10, 14, 5, 6, 0, time.UTC)
	got := defaultExportPath(ts)
	want := "note_backup_20260310_140506.zip"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && indexOf(s, sub) >= 0
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
