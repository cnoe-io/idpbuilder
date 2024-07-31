package helpers

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	LogLevel    string
	LogLevelMsg = "Set the log verbosity. Supported values are: debug, info, warn, and error."
	CmdLogger   logr.Logger
)

func SetLogger() error {
	l, err := getSlogLevel(LogLevel)
	if err != nil {
		return err
	}

	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l}))
	kslogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: getKlogLevel(l)}))
	logger := logr.FromSlogHandler(slogger.Handler())
	klogger := logr.FromSlogHandler(kslogger.Handler())

	klog.SetLogger(klogger)
	ctrl.SetLogger(logger)
	CmdLogger = logger
	return nil
}

func getSlogLevel(s string) (slog.Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelDebug, fmt.Errorf("%s is not a valid log level", s)
	}
}

// For end users, klog messages are mostly useless. We set it to error level unless debug logging is enabled.
func getKlogLevel(l slog.Level) slog.Level {
	if l < slog.LevelInfo {
		return l
	}
	return slog.LevelError
}
