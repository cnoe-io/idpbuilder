package runtime

import (
	"errors"
	"strconv"
	"strings"

	"sigs.k8s.io/kind/pkg/exec"
)

func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "-v")
	lines, err := exec.OutputLines(cmd)
	if err != nil || len(lines) != 1 {
		return false
	}
	return strings.HasPrefix(lines[0], "Docker version")
}

func IsFinchAvailable() bool {
	cmd := exec.Command("nerdctl", "-v")
	lines, err := exec.OutputLines(cmd)
	if err != nil || len(lines) != 1 {
		// check finch
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
