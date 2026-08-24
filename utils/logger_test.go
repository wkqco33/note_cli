package utils

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// captureLogger 테스트용 버퍼에 쓰는 로거로 전환하고, 이전 DebugMode를 반환.
// 원래 핸들러로 복원하는 것을 deferred 콜백으로 반환한다.
func captureLogger(t *testing.T) (buf *bytes.Buffer, restore func()) {
	t.Helper()
	orig := DebugMode
	buf = &bytes.Buffer{}
	SetLoggerHandler(slog.NewTextHandler(buf, &slog.HandlerOptions{
		Level: logLevel,
	}))
	return buf, func() {
		DebugMode = orig
		SetupLogger()
	}
}

func TestSetupLoggerDebugModeOn(t *testing.T) {
	buf, restore := captureLogger(t)
	defer restore()

	DebugMode = true
	SetupLogger()
	Debugln("hello debug")

	if !strings.Contains(buf.String(), "hello debug") {
		t.Fatalf("expected debug log to be emitted, got: %q", buf.String())
	}
}

func TestSetupLoggerDebugModeOff(t *testing.T) {
	buf, restore := captureLogger(t)
	defer restore()

	DebugMode = false
	SetupLogger()
	Debugf("should be hidden %d", 42)

	if buf.Len() != 0 {
		t.Fatalf("expected no log output when debug off, got: %q", buf.String())
	}
}

func TestSetupLoggerTogglesBackToOff(t *testing.T) {
	buf, restore := captureLogger(t)
	defer restore()

	DebugMode = true
	SetupLogger()
	DebugMode = false
	SetupLogger()
	Debugln("again hidden")

	if buf.Len() != 0 {
		t.Fatalf("expected no output after toggling off, got: %q", buf.String())
	}
}
