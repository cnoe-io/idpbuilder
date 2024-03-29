package runtime

import (
	"context"
)

const (
	Running = "running"
)

type IRuntime interface {
	// checks whether the container has the followin
	ContainerWithPort(ctx context.Context, name, port string) (bool, error)
}
