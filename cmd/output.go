package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/glamour/styles"
	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// quietMode 비필수 진행/상태 메시지를 억제한다 (-q/--quiet).
// 출력 계층이 직접 참조하므로 output.go에 정의한다.
var quietMode bool

// statusWriter 진행/상태 안내 출력 대상.
// 실제 결과(표/JSON/YAML/노트 본문)는 stdout에 두고, 사람용 진행 메시지는 stderr로 보내
// 파이프/스크립트에서 stdout이 오염되지 않게 한다.
var statusWriter io.Writer = os.Stderr

// statusf 진행/상태 안내를 stderr로 출력한다. -q/--quiet면 생략한다.
func statusf(format string, args ...any) {
	if quietMode {
		return
	}
	_, _ = fmt.Fprintf(statusWriter, format+"\n", args...)
}

// statusln 진행/상태 안내를 stderr에 개행과 함께 출력한다. -q/--quiet면 생략한다.
func statusln(args ...any) {
	if quietMode {
		return
	}
	_, _ = fmt.Fprintln(statusWriter, args...)
}

// progressBarVisible 진행률 표시줄 노출 여부.
// quiet이거나 stderr가 터미널이 아니면(CI/리다이렉트) 애니메이션을 출력하지 않는다.
func progressBarVisible(quiet, stderrTTY bool) bool {
	return !quiet && stderrTTY
}

// newProgressBar 진행률 표시줄을 만든다. 노출하지 않는 경우 출력을 버리는 바를 반환한다.
func newProgressBar(total int64, description string) *progressbar.ProgressBar {
	if !progressBarVisible(quietMode, term.IsTerminal(int(os.Stderr.Fd()))) {
		return progressbar.DefaultBytesSilent(total, description)
	}

	return progressbar.DefaultBytes(total, description)
}

// resolveMarkdownStyle 색상 비활성 여부에 따라 glamour 스타일을 고른다.
// --no-color/NO_COLOR이면 색상 없는 notty 스타일을 사용한다.
func resolveMarkdownStyle(noColor bool) string {
	if noColor {
		return styles.NoTTYStyle
	}

	return styles.AutoStyle
}
