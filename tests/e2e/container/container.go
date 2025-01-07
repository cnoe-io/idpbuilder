package container

import (
	"context"
	"time"
)

type Engine interface {
	RunIdpCommand(ctx context.Context, cmd string, timeout time.Duration) ([]byte, error)
	RunCommand(ctx context.Context, cmd string, timeout time.Duration) ([]byte, error)
	GetClient() string
}
