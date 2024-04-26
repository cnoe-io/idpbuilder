package runtime

import (
	"errors"
	"os"
	"strconv"
)

func DetectRuntime() (rt IRuntime, err error) {
	switch p := os.Getenv("KIND_EXPERIMENTAL_PROVIDER"); p {
	case "", "docker":
		return NewDockerRuntime("docker")
	case "podman":
		return NewDockerRuntime("podman")
	case "finch":
		return NewFinchRuntime()
	default:
		return nil, errors.New("runtime unknown or not supported")
	}
}

func toUint16(portString string) (uint16, error) {
	// Convert port string to uint16
	port, err := strconv.ParseUint(portString, 10, 16)
	if err != nil {
		return 0, err
	}

	// Port validation
	if port > 65535 {
		return 0, errors.New("invalid port number")
	}

	return uint16(port), nil
}
