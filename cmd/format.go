package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"note_cli/api"

	"gopkg.in/yaml.v3"
)

// outputFormat 조회 커맨드의 출력 형식
type outputFormat string

const (
	formatText outputFormat = "text"
	formatJSON outputFormat = "json"
	formatYAML outputFormat = "yaml"
)

// supportedOutputFormats 허용되는 출력 형식 목록
var supportedOutputFormats = []string{"text", "json", "yaml"}

// parseOutputFormat 출력 형식 플래그 값을 파싱한다. 빈 값은 text 기본값.
func parseOutputFormat(value string) (outputFormat, error) {
	trimmed := strings.TrimSpace(strings.ToLower(value))
	if trimmed == "" {
		return formatText, nil
	}
	for _, format := range supportedOutputFormats {
		if trimmed == format {
			return outputFormat(trimmed), nil
		}
	}
	return "", fmt.Errorf("지원하지 않는 출력 형식입니다: %s (가능한 값: %s)", value, strings.Join(supportedOutputFormats, ", "))
}

// renderNotes 노트 목록을 지정한 형식으로 렌더링한다.
// text는 기존 표 출력, json/yaml은 구조화 출력을 반환한다.
func renderNotes(notes []api.BoardRead, format outputFormat) (string, error) {
	switch format {
	case formatText:
		return buildBoardTable(notes), nil
	case formatJSON:
		data, err := json.MarshalIndent(notes, "", "  ")
		if err != nil {
			return "", fmt.Errorf("노트 목록을 JSON으로 변환하지 못했습니다: %w", err)
		}
		return string(data), nil
	case formatYAML:
		data, err := yaml.Marshal(notes)
		if err != nil {
			return "", fmt.Errorf("노트 목록을 YAML로 변환하지 못했습니다: %w", err)
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("지원하지 않는 출력 형식입니다: %s", format)
	}
}

// renderNote 단일 노트를 지정한 형식으로 렌더링한다.
// text는 빈 문자열을 반환하므로 호출부에서 기존 렌더링 경로를 사용해야 한다.
func renderNote(note *api.BoardRead, format outputFormat) (string, error) {
	switch format {
	case formatText:
		return "", nil
	case formatJSON:
		data, err := json.MarshalIndent(note, "", "  ")
		if err != nil {
			return "", fmt.Errorf("노트를 JSON으로 변환하지 못했습니다: %w", err)
		}
		return string(data), nil
	case formatYAML:
		data, err := yaml.Marshal(note)
		if err != nil {
			return "", fmt.Errorf("노트를 YAML로 변환하지 못했습니다: %w", err)
		}
		return string(data), nil
	default:
		return "", fmt.Errorf("지원하지 않는 출력 형식입니다: %s", format)
	}
}

// printRenderedNotes 렌더링 결과를 표준 출력에 기록한다
func printRenderedNotes(rendered string) {
	_, _ = fmt.Fprint(os.Stdout, rendered)
}
