package registry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PushProgress struct {
	Repository    string        `json:"repository"`
	Tag           string        `json:"tag"`
	LayersTotal   int           `json:"layers_total"`
	LayersCurrent int           `json:"layers_current"`
	BytesCurrent  int64         `json:"bytes_current"`
	Status        string        `json:"status"`
	StartTime     time.Time     `json:"start_time"`
	Duration      time.Duration `json:"duration"`
}

type PushConfig struct {
	ChunkSize        int64
	ProgressCallback func(*PushProgress)
}

func DefaultPushConfig() *PushConfig {
	return &PushConfig{ChunkSize: 5 * 1024 * 1024} // 5MB chunks
}

type Layer struct {
	Digest    string    `json:"digest"`
	Size      int64     `json:"size"`
	MediaType string    `json:"mediaType"`
	Data      io.Reader `json:"-"`
}

type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Config        Layer   `json:"config"`
	Layers        []Layer `json:"layers"`
}

type ImagePusher struct {
	registry *GiteaRegistry
	config   *PushConfig
	progress *PushProgress
}

func NewImagePusher(registry *GiteaRegistry, config *PushConfig) *ImagePusher {
	if config == nil {
		config = DefaultPushConfig()
	}
	return &ImagePusher{
		registry: registry,
		config:   config,
		progress: &PushProgress{StartTime: time.Now(), Status: "initializing"},
	}
}

func (p *ImagePusher) updateProgress(update func(*PushProgress)) {
	update(p.progress)
	p.progress.Duration = time.Since(p.progress.StartTime)
	if p.config.ProgressCallback != nil { p.config.ProgressCallback(p.progress) }
}

func (p *ImagePusher) initiateBlobUpload(ctx context.Context, repository string) (string, error) {
	uploadURL := fmt.Sprintf("%s/v2/%s/blobs/uploads/", p.registry.baseURL, repository)
	
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, nil)
	if err != nil {
		return "", err
	}
	
	if authHeader, err := p.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	
	resp, err := p.registry.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	
	return resp.Header.Get("Location"), nil
}

func (p *ImagePusher) uploadBlobChunk(ctx context.Context, uploadURL string, data []byte, start, end int64) error {
	req, err := http.NewRequestWithContext(ctx, "PATCH", uploadURL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Range", fmt.Sprintf("%d-%d", start, end-1))
	req.Header.Set("Content-Length", strconv.FormatInt(int64(len(data)), 10))
	req.Header.Set("Content-Type", "application/octet-stream")
	
	if authHeader, err := p.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	
	resp, err := p.registry.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("unexpected chunk status: %d", resp.StatusCode)
	}
	
	return nil
}

func (p *ImagePusher) finalizeBlobUpload(ctx context.Context, uploadURL, expectedDigest string) error {
	req, err := http.NewRequestWithContext(ctx, "PUT", uploadURL+"&digest="+expectedDigest, nil)
	if err != nil {
		return err
	}
	if authHeader, err := p.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	resp, err := p.registry.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (p *ImagePusher) pushLayer(ctx context.Context, repository string, layer Layer) error {
	// Check if layer exists
	if exists, _ := p.layerExists(ctx, repository, layer.Digest); exists {
		return nil
	}
	
	// Initiate upload
	uploadURL, err := p.initiateBlobUpload(ctx, repository)
	if err != nil {
		return err
	}
	
	// Parse upload URL
	baseURL, _ := url.Parse(p.registry.baseURL)
	fullUploadURL, _ := baseURL.Parse(uploadURL)
	uploadURL = fullUploadURL.String()
	
	// Upload in chunks
	buffer := make([]byte, p.config.ChunkSize)
	var uploaded int64
	
	for {
		n, err := layer.Data.Read(buffer)
		if n > 0 {
			if err := p.uploadBlobChunk(ctx, uploadURL, buffer[:n], uploaded, uploaded+int64(n)); err != nil {
				return err
			}
			uploaded += int64(n)
			p.updateProgress(func(prog *PushProgress) { prog.BytesCurrent += int64(n); prog.Status = "uploading" })
		}
		if err == io.EOF { break }
		if err != nil { return err }
	}
	
	// Finalize upload
	if err := p.finalizeBlobUpload(ctx, uploadURL, layer.Digest); err != nil {
		return err
	}
	
	p.updateProgress(func(prog *PushProgress) { prog.LayersCurrent++ })
	return nil
}

func (p *ImagePusher) layerExists(ctx context.Context, repository, digest string) (bool, error) {
	checkURL := fmt.Sprintf("%s/v2/%s/blobs/%s", p.registry.baseURL, repository, digest)
	
	req, err := http.NewRequestWithContext(ctx, "HEAD", checkURL, nil)
	if err != nil {
		return false, err
	}
	
	if authHeader, err := p.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	
	resp, err := p.registry.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	
	return resp.StatusCode == http.StatusOK, nil
}

func (p *ImagePusher) pushManifest(ctx context.Context, repository, tag string, manifest *Manifest) error {
	manifestURL := fmt.Sprintf("%s/v2/%s/manifests/%s", p.registry.baseURL, repository, tag)
	
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	
	req, err := http.NewRequestWithContext(ctx, "PUT", manifestURL, bytes.NewReader(manifestBytes))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", manifest.MediaType)
	
	if authHeader, err := p.registry.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	
	resp, err := p.registry.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected manifest status: %d", resp.StatusCode)
	}
	
	return nil
}

func (p *ImagePusher) PushImage(ctx context.Context, repository, tag string, manifest *Manifest) error {
	p.updateProgress(func(prog *PushProgress) { prog.Repository = repository; prog.Tag = tag; prog.LayersTotal = len(manifest.Layers) + 1; prog.Status = "starting" })
	
	if err := p.pushLayer(ctx, repository, manifest.Config); err != nil { return err }
	
	for _, layer := range manifest.Layers {
		if err := p.pushLayer(ctx, repository, layer); err != nil { return err }
	}
	
	if err := p.pushManifest(ctx, repository, tag, manifest); err != nil { return err }
	
	p.updateProgress(func(prog *PushProgress) { prog.Status = "complete" })
	return nil
}

func ParseImageRef(image string) (repository, tag string, err error) {
	parts := strings.SplitN(image, ":", 2)
	if len(parts) == 1 {
		return parts[0], "latest", nil
	}
	return parts[0], parts[1], nil
}

func calculateDigest(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash)
}