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
)

func SetLogger() error {
	l, err := getSlogLevel(LogLevel)
	if err != nil {
		return err
	}
	slogger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l}))
	logger := logr.FromSlogHandler(slogger.Handler())
	klog.SetLogger(logger)
	ctrl.SetLogger(logger)
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
