package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// spinnerFrames 스피너 애니메이션 프레임.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// WithSpinner 메시지와 함께 스피너를 표시하면서 fn을 실행한다.
// 표준 출력이 터미널이 아닌 경우 스피너 없이 fn만 실행한다.
// 스피너는 표준 오류(stderr)에 출력해 표준 출력 결과를 오염시키지 않는다.
func WithSpinner(message string, fn func() error) error {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return fn()
	}
	done := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
			}
			fmt.Fprintf(os.Stderr, "\r%s %s", spinnerFrames[i%len(spinnerFrames)], message)
			i++
			time.Sleep(100 * time.Millisecond)
		}
	}()
	err := fn()
	close(done)
	clearSpinnerLine(message)
	return err
}

// clearSpinnerLine 스피너가 차지한 줄을 지운다.
func clearSpinnerLine(message string) {
	width := runewidth.StringWidth(message) + 2
	fmt.Fprintf(os.Stderr, "\r%s\r", strings.Repeat(" ", width))
}
