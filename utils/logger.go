package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// DebugMode determines if debug logs should be printed
var DebugMode bool

// logLevel 로거 재생성 없이 출력 레벨을 전환하기 위한 동적 레벨 (기본 Info)
var logLevel = new(slog.LevelVar)

var logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
	Level: logLevel,
}))

// SetLoggerHandler 테스트 등에서 로거의 출력 핸들러를 교체한다.
func SetLoggerHandler(h slog.Handler) {
	logger = slog.New(h)
}

// SetupLogger는 DebugMode 값에 따라 로거의 출력 레벨을 재설정합니다.
func SetupLogger() {
	if DebugMode {
		logLevel.Set(slog.LevelDebug)
		return
	}
	logLevel.Set(slog.LevelInfo)
}

// Debugf prints formatted debug logs if DebugMode is true
func Debugf(format string, v ...interface{}) {
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	logger.Debug(fmt.Sprintf(format, v...))
}

// Debugln prints debug logs if DebugMode is true
func Debugln(v ...interface{}) {
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	logger.Debug(fmt.Sprint(v...))
}
