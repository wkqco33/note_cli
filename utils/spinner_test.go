package utils

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func withQuiet(t *testing.T, quiet bool) {
	t.Helper()
	prev := Quiet
	Quiet = quiet
	t.Cleanup(func() { Quiet = prev })
}

func TestWithSpinnerRunsFunctionAndPropagatesError(t *testing.T) {
	withQuiet(t, false)

	wantErr := errors.New("boom")
	ran := false
	err := WithSpinner("작업 중...", func() error {
		ran = true
		return wantErr
	})

	if !ran {
		t.Fatal("WithSpinner가 fn을 실행하지 않았습니다")
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("WithSpinner() error = %v, want %v", err, wantErr)
	}
}

func TestWithSpinnerQuietStillRunsFunction(t *testing.T) {
	withQuiet(t, true)

	ran := false
	if err := WithSpinner("작업 중...", func() error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("WithSpinner() error = %v, want nil", err)
	}
	if !ran {
		t.Fatal("quiet 모드에서도 fn은 실행되어야 합니다")
	}
}

func TestWithSpinnerNonTTYRunsFunction(t *testing.T) {
	// 테스트 환경의 stdout은 터미널이 아니므로 스피너 없이 fn만 실행된다.
	withQuiet(t, false)

	ran := false
	if err := WithSpinner("작업 중...", func() error {
		ran = true
		return nil
	}); err != nil {
		t.Fatalf("WithSpinner() error = %v, want nil", err)
	}
	if !ran {
		t.Fatal("비TTY에서도 fn은 실행되어야 합니다")
	}
}

func TestWithSpinnerHidesFastWork(t *testing.T) {
	var buf bytes.Buffer

	if err := withSpinner("작업 중", &buf, 40*time.Millisecond, func() error { return nil }); err != nil {
		t.Fatalf("withSpinner() error = %v, want nil", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("빠른 작업에서 스피너가 출력되었습니다: %q", buf.String())
	}
}

func TestWithSpinnerShowsSlowWork(t *testing.T) {
	var buf bytes.Buffer

	err := withSpinner("작업 중", &buf, time.Millisecond, func() error {
		time.Sleep(30 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("withSpinner() error = %v, want nil", err)
	}
	if !strings.Contains(buf.String(), "작업 중") {
		t.Fatalf("느린 작업에서 스피너가 출력되지 않았습니다: %q", buf.String())
	}
}
