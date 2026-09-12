package cmd

import (
	"encoding/json"
	"strings"
	"testing"

	"note_cli/api"

	"gopkg.in/yaml.v3"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    outputFormat
		wantErr bool
	}{
		{name: "빈 값은 text 기본", input: "", want: formatText},
		{name: "text", input: "text", want: formatText},
		{name: "json", input: "json", want: formatJSON},
		{name: "yaml", input: "yaml", want: formatYAML},
		{name: "대소문자 무시", input: "JSON", want: formatJSON},
		{name: "공백 제거", input: " yaml ", want: formatYAML},
		{name: "지원하지 않는 형식은 에러", input: "xml", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutputFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseOutputFormat(%q) error = %v, wantErr = %v", tt.input, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("parseOutputFormat(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRenderNotesJSON(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "할 일", Category: "work", Content: "내용"},
	}

	got, err := renderNotes(notes, formatJSON)
	if err != nil {
		t.Fatalf("renderNotes() error = %v", err)
	}

	var parsed []api.BoardRead
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("JSON 파싱 실패: %v\n%s", err, got)
	}
	if len(parsed) != 1 || parsed[0].Title != "할 일" {
		t.Fatalf("unexpected parsed notes: %+v", parsed)
	}
}

func TestRenderNotesYAML(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "할 일", Category: "work", Content: "내용"},
	}

	got, err := renderNotes(notes, formatYAML)
	if err != nil {
		t.Fatalf("renderNotes() error = %v", err)
	}

	var parsed []api.BoardRead
	if err := yaml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("YAML 파싱 실패: %v\n%s", err, got)
	}
	if len(parsed) != 1 || parsed[0].Title != "할 일" {
		t.Fatalf("unexpected parsed notes: %+v", parsed)
	}
}

func TestRenderNotesTextDelegatesToTable(t *testing.T) {
	notes := []api.BoardRead{
		{ID: 1, Title: "할 일", Category: "work", UpdatedAt: "2024-01-02T10:00:00"},
	}

	got, err := renderNotes(notes, formatText)
	if err != nil {
		t.Fatalf("renderNotes() error = %v", err)
	}
	if got != buildBoardTable(notes) {
		t.Fatalf("text 포맷은 표 출력과 동일해야 합니다")
	}
}

func TestRenderNoteJSONRoundTrip(t *testing.T) {
	note := &api.BoardRead{ID: 3, Title: "회의 메모", Content: "# 본문", Category: "work", UpdatedAt: "2024-01-02T10:00:00"}

	got, err := renderNote(note, formatJSON)
	if err != nil {
		t.Fatalf("renderNote() error = %v", err)
	}

	var parsed api.BoardRead
	if err := json.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("JSON 파싱 실패: %v\n%s", err, got)
	}
	if parsed.ID != 3 || parsed.Title != "회의 메모" || parsed.Content != "# 본문" {
		t.Fatalf("unexpected parsed note: %+v", parsed)
	}
}

func TestRenderNoteYAML(t *testing.T) {
	note := &api.BoardRead{ID: 3, Title: "회의 메모", Content: "# 본문", Category: "work"}

	got, err := renderNote(note, formatYAML)
	if err != nil {
		t.Fatalf("renderNote() error = %v", err)
	}

	var parsed api.BoardRead
	if err := yaml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("YAML 파싱 실패: %v\n%s", err, got)
	}
	if parsed.ID != 3 || parsed.Title != "회의 메모" {
		t.Fatalf("unexpected parsed note: %+v", parsed)
	}
}

func TestRenderNotesUnsupportedFormat(t *testing.T) {
	if _, err := renderNotes(nil, outputFormat("xml")); err == nil {
		t.Fatal("renderNotes() expected an error for unsupported format")
	}
	if !strings.Contains(renderAICreateDryRun(aiCreateDraft{}), "--dry-run") {
		t.Fatal("renderAICreateDryRun sanity check failed")
	}
}

func TestOutputFormatFromFlags(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		json    bool
		want    outputFormat
		wantErr bool
	}{
		{name: "미지정은 text", format: "", json: false, want: formatText},
		{name: "--format yaml", format: "yaml", json: false, want: formatYAML},
		{name: "--json은 json", format: "", json: true, want: formatJSON},
		{name: "--json이 --format보다 우선", format: "yaml", json: true, want: formatJSON},
		{name: "--json이면 잘못된 --format도 무시", format: "bad", json: true, want: formatJSON},
		{name: "잘못된 --format은 에러", format: "bad", json: false, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := outputFormatFromFlags(tt.format, tt.json)
			if (err != nil) != tt.wantErr {
				t.Fatalf("outputFormatFromFlags() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("outputFormatFromFlags(%q, %v) = %v, want %v", tt.format, tt.json, got, tt.want)
			}
		})
	}
}
