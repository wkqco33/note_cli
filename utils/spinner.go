package utils

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// spinnerFrames 스피너 애니메이션 프레임.
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// Quiet true이면 스피너 같은 비필수 출력을 억제한다 (-q/--quiet).
var Quiet bool

// spinnerDelay 스피너를 표시하기 전 대기 시간.
// 빠른 작업(로컬 조회 등)에서는 스피너가 깜빡이지 않게 하고,
// 느린 작업(네트워크 요청)에서만 피드백을 제공한다.
var spinnerDelay = 150 * time.Millisecond

// spinnerInterval 스피너 프레임 갱신 주기.
var spinnerInterval = 100 * time.Millisecond

// WithSpinner 메시지와 함께 스피너를 표시하면서 fn을 실행한다.
// 표준 출력이 터미널이 아니거나 Quiet이면 스피너 없이 fn만 실행한다.
// 스피너는 표준 오류(stderr)에 출력해 표준 출력 결과를 오염시키지 않는다.
func WithSpinner(message string, fn func() error) error {
	if Quiet || !term.IsTerminal(int(os.Stdout.Fd())) {
		return fn()
	}

	return withSpinner(message, os.Stderr, spinnerDelay, fn)
}

// withSpinner 스피너 코어. delay가 지나도 fn이 끝나지 않을 때만 out에 프레임을 쓴다.
// 테스트에서 TTY 없이 지연/표시 동작을 검증할 수 있도록 분리되어 있다.
func withSpinner(message string, out io.Writer, delay time.Duration, fn func() error) error {
	done := make(chan struct{})
	var started atomic.Bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// 빠른 작업이면 스피너를 시작하지 않는다.
		select {
		case <-done:
			return
		case <-time.After(delay):
		}

		started.Store(true)
		i := 0
		for {
			select {
			case <-done:
				return
			default:
			}
			_, _ = fmt.Fprintf(out, "\r%s %s", spinnerFrames[i%len(spinnerFrames)], message)
			i++
			time.Sleep(spinnerInterval)
		}
	}()

	err := fn()
	close(done)
	wg.Wait()
	if started.Load() {
		clearSpinnerLine(out, message)
	}

	return err
}

// clearSpinnerLine 스피너가 차지한 줄을 지운다.
func clearSpinnerLine(w io.Writer, message string) {
	width := runewidth.StringWidth(message) + 2
	_, _ = fmt.Fprintf(w, "\r%s\r", strings.Repeat(" ", width))
}
