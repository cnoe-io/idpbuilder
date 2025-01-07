package container

import (
	"context"
	"time"
)

const (
	IdpbuilderBinaryLocation = "../../../idpbuilder"
)

type Engine interface {
	RunIdpCommand(ctx context.Context, cmd string, timeout time.Duration) ([]byte, error)
	RunCommand(ctx context.Context, cmd string, timeout time.Duration) ([]byte, error)
	GetClient() string
}
