package apps

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/cnoe-io/idpbuilder/globals"
	"github.com/cnoe-io/idpbuilder/pkg/docker"
	"github.com/cnoe-io/idpbuilder/pkg/util"

	"github.com/docker/docker/api/types"
	registryTypes "github.com/docker/docker/api/types/registry"
	dockerClient "github.com/docker/docker/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func PushImage(ctx context.Context, client *dockerClient.Client, tag string) (*string, error) {
	log := log.FromContext(ctx)

	// Push docker image
	log.Info("Pushing docker image", "tag", tag)
	authConfig, err := registryTypes.EncodeAuthConfig(registryTypes.AuthConfig{})
	if err != nil {
		return nil, err
	}
	resp, err := client.ImagePush(ctx, tag, types.ImagePushOptions{
		RegistryAuth: authConfig,
	})
	if err != nil {
		return nil, err
	}
	defer resp.Close()

	// Check for error in output
	var pushErrorMessage docker.ErrorMessage
	pushBuffIOReader := bufio.NewReader(resp)

	for {
		streamBytes, err := pushBuffIOReader.ReadBytes('\n')
		if err == io.EOF {
			break
		}
		json.Unmarshal(streamBytes, &pushErrorMessage)
		if pushErrorMessage.ErrorStr != "" {
			return nil, pushErrorMessage
		}
	}

	di, err := client.DistributionInspect(ctx, tag, authConfig)
	if err != nil {
		return nil, err
	}

	regImgId := di.Descriptor.Digest.String()
	log.Info("Image Pushed", "digest", regImgId)

	return &regImgId, nil
}

func BuildAppsImage(ctx context.Context, client *dockerClient.Client, tags []string, labels map[string]string, appsFS fs.FS) (*string, error) {
	log := log.FromContext(ctx)

	log.Info("Building docker image")

	// Create docker image dir and defer cleanup
	workDir, err := os.MkdirTemp("", fmt.Sprintf("%s-dockerimage", globals.ProjectName))
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(workDir)

	// Write base image to work dir
	if err := util.CopyDir(GitServerFS, workDir); err != nil {
		log.Error(err, "Copying base git server contents to work dir")
		return nil, err
	}

	// Write apps fs to workdir
	srvWorkDir := filepath.Join(workDir, "srv")
	if err = os.Mkdir(srvWorkDir, 0700); err != nil {
		log.Error(err, "Creating srv dir for apps fs")
		return nil, err
	}
	if err = util.WriteFS(appsFS, srvWorkDir); err != nil {
		log.Error(err, "Writing apps fs to work dir")
		return nil, err
	}

	// Build docker image
	auxMessage, err := docker.BuildDir(ctx, client, workDir, types.ImageBuildOptions{
		Tags:   tags,
		Labels: labels,
	})
	if err != nil {
		log.Error(err, "Building docker image")
		return nil, err
	}
	if auxMessage == nil {
		return nil, fmt.Errorf("docker image build failed")
	}

	return &auxMessage.Aux.ID, nil
}
