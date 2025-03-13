//go:build e2e

package container

import (
	"context"
	"time"
)

type Engine interface {
	RunCommand(ctx context.Context, cmd string, timeout time.Duration) ([]byte, error)
	GetClient() string
}
