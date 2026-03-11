package utils

import (
	"fmt"
	"log/slog"
	"os"
)

// DebugMode determines if debug logs should be printed
var DebugMode bool

var logger *slog.Logger

func init() {
	// 기본 로거는 Info 레벨로 설정하여 Debug 로그를 무시하도록 합니다.
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

// SetupLogger는 DebugMode 값에 따라 로거의 출력 레벨을 재설정합니다.
func SetupLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	if DebugMode {
		opts.Level = slog.LevelDebug
	}
	logger = slog.New(slog.NewTextHandler(os.Stderr, opts))
}

// Debugf prints formatted debug logs if DebugMode is true
func Debugf(format string, v ...interface{}) {
	if DebugMode {
		logger.Debug(fmt.Sprintf(format, v...))
	}
}

// Debugln prints debug logs if DebugMode is true
func Debugln(v ...interface{}) {
	if DebugMode {
		logger.Debug(fmt.Sprint(v...))
	}
}
