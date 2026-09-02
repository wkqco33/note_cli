// Package attachments는 LLM 분석을 위한 첨부파일 텍스트 추출을 제공한다.
package attachments

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"strings"
)

const MaxAnalysisBytes int64 = 2 * 1024 * 1024

// ExtractText 지원되는 텍스트 파일의 내용을 읽는다.
func ExtractText(filename, contentType string, r io.Reader) (string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !isSupported(ext, contentType) {
		return "", fmt.Errorf("지원하지 않는 첨부파일 형식입니다: %s", filename)
	}
	limited := io.LimitReader(r, MaxAnalysisBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return "", fmt.Errorf("첨부파일을 읽지 못했습니다: %w", err)
	}
	if int64(len(data)) > MaxAnalysisBytes {
		return "", fmt.Errorf("LLM 분석 대상 파일은 %dMB 이하만 지원합니다", MaxAnalysisBytes/(1024*1024))
	}
	text := string(data)
	switch ext {
	case ".json":
		var value any
		if err := json.Unmarshal(data, &value); err != nil {
			return "", fmt.Errorf("JSON 파일을 해석하지 못했습니다: %w", err)
		}
		formatted, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return "", fmt.Errorf("JSON 파일을 포맷하지 못했습니다: %w", err)
		}
		text = string(formatted)
	case ".csv":
		text = normalizeLines(text)
	}
	if strings.TrimSpace(text) == "" {
		return "", fmt.Errorf("첨부파일 내용이 비어 있습니다")
	}
	return text, nil
}

func isSupported(ext, contentType string) bool {
	if ext == ".txt" || ext == ".md" || ext == ".markdown" || ext == ".csv" || ext == ".json" || ext == ".log" {
		return true
	}
	baseType, _, _ := mime.ParseMediaType(contentType)
	return baseType == "text/plain" || baseType == "text/markdown" || baseType == "text/csv" || baseType == "application/json"
}

func normalizeLines(value string) string {
	scanner := bufio.NewScanner(strings.NewReader(value))
	var lines []string
	for scanner.Scan() {
		lines = append(lines, strings.TrimSpace(scanner.Text()))
	}
	return strings.Join(lines, "\n")
}
