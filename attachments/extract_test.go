package attachments

import (
	"strings"
	"testing"
)

func TestExtractTextFormatsJSON(t *testing.T) {
	got, err := ExtractText("data.json", "", strings.NewReader(`{"name":"note","items":[1,2]}`))
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}
	if !strings.Contains(got, "\n  \"items\"") {
		t.Fatalf("JSON was not formatted: %s", got)
	}
}

func TestExtractTextRejectsUnsupportedAndOversizedFiles(t *testing.T) {
	if _, err := ExtractText("image.png", "image/png", strings.NewReader("data")); err == nil {
		t.Fatal("unsupported file was accepted")
	}
	tooLarge := strings.NewReader(strings.Repeat("a", int(MaxAnalysisBytes)+1))
	if _, err := ExtractText("large.txt", "", tooLarge); err == nil {
		t.Fatal("oversized file was accepted")
	}
}

func TestExtractTextUsesContentType(t *testing.T) {
	got, err := ExtractText("unknown", "text/plain; charset=utf-8", strings.NewReader("hello"))
	if err != nil || got != "hello" {
		t.Fatalf("ExtractText() = %q, error = %v", got, err)
	}
}
