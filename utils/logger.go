package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// DebugMode 디버그 로그 출력 여부
var DebugMode bool

// logLevel 출력 레벨을 동적으로 전환하기 위한 변수
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

// Debugf DebugMode가 true일 때 포맷된 디버그 로그를 출력한다.
func Debugf(format string, v ...interface{}) {
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	logger.Debug(fmt.Sprintf(format, v...))
}

// Debugln DebugMode가 true일 때 디버그 로그를 출력한다.
func Debugln(v ...interface{}) {
	if !logger.Enabled(context.Background(), slog.LevelDebug) {
		return
	}
	logger.Debug(fmt.Sprint(v...))
}
