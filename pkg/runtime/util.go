package runtime

import (
	"errors"
	"strconv"
	"strings"

	"sigs.k8s.io/kind/pkg/exec"
)

func DetectRuntime() (rt IRuntime, err error) {
	if isDockerAvailable() {
		if rt, err = NewDockerRuntime(); err != nil {
			return nil, err
		}
	} else if isFinchAvailable() {
		rt, _ = NewFinchRuntime()
	} else {
		return nil, errors.New("no runtime found")
	}
	return rt, nil
}

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "-v")
	lines, err := exec.OutputLines(cmd)
	if err != nil || len(lines) != 1 {
		return false
	}
	return strings.HasPrefix(lines[0], "Docker version")
}

func isFinchAvailable() bool {
	cmd := exec.Command("nerdctl", "-v")
	lines, err := exec.OutputLines(cmd)
	if err != nil || len(lines) != 1 {
		cmd = exec.Command("finch", "-v")
		lines, err = exec.OutputLines(cmd)
		if err != nil || len(lines) != 1 {
			return false
		}
		return strings.HasPrefix(lines[0], "finch version")
	}
	return strings.HasPrefix(lines[0], "nerdctl version")
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
