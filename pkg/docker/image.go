package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	dockerClient "github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func GetDockerClient() (*dockerClient.Client, error) {
	return dockerClient.NewClientWithOpts()
}

type AuxBody struct {
	ID string
}

type AuxMessage struct {
	Aux AuxBody `json:"aux"`
}

type ErrorMessage struct {
	ErrorStr string
}

func (e ErrorMessage) Error() string {
	return e.ErrorStr
}

func BuildDir(ctx context.Context, client *dockerClient.Client, path string, buildOpts types.ImageBuildOptions) (*AuxMessage, error) {
	log := log.FromContext(ctx)

	tar, err := archive.TarWithOptions(path, &archive.TarOptions{})
	if err != nil {
		return nil, err
	}

	if buildOpts.Dockerfile == "" {
		buildOpts.Dockerfile = "Dockerfile"
	}

	br, err := client.ImageBuild(ctx, tar, buildOpts)
	if err != nil {
		return nil, err
	}
	defer br.Body.Close()

	var buildOut []string

	var buildErrorMessage ErrorMessage
	var auxMessage *AuxMessage
	buildBuffIOReader := bufio.NewReader(br.Body)

	for {
		var tryAuxMessage AuxMessage
		streamBytes, err := buildBuffIOReader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		if err = json.Unmarshal(streamBytes, &buildErrorMessage); err != nil {
			return nil, err
		}
		if buildErrorMessage.ErrorStr != "" {
			return nil, buildErrorMessage
		}
		if err = json.Unmarshal(streamBytes, &tryAuxMessage); err != nil {
			return nil, err
		}
		if tryAuxMessage.Aux.ID != "" {
			auxMessage = &tryAuxMessage
		}
		buildOut = append(buildOut, string(streamBytes))
	}
	log.Info("Docker build output", "output", strings.Join(buildOut, "\n"))

	return auxMessage, nil
}
