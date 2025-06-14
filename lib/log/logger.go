package log

import (
	"fmt"
	"log/slog"
	"os"
	"sync"
)

var initSync sync.Once
var stdErrLogger *slog.Logger
var logLevel slog.LevelVar

func defaultStdErrLogger() *slog.Logger {
	initSync.Do(func() {
		h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: &logLevel})
		l := slog.New(h)
		stdErrLogger = l
	})

	return stdErrLogger
}

func SetDefaultLogLevel(level slog.Level) { logLevel.Set(level) }
func Default() *slog.Logger               { return defaultStdErrLogger() }

func DebugErr(msg string, err error) {
	defaultStdErrLogger().
		Debug(msg, slog.String("error", err.Error()))
}

func Infof(format string, args ...any) {
	defaultStdErrLogger().
		Info(fmt.Sprintf(format, args...))
}
