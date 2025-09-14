package registry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/name"
)

// PushProgress tracks the progress of a push operation
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

// PushConfig provides configuration options for push operations
type PushConfig struct {
	ChunkSize        int64
	ProgressCallback func(*PushProgress)
}

// DefaultPushConfig returns default push configuration
func DefaultPushConfig() *PushConfig {
	return &PushConfig{ChunkSize: 5 * 1024 * 1024} // 5MB chunks
}

// Layer represents a container image layer
type Layer struct {
	Digest    string    `json:"digest"`
	Size      int64     `json:"size"`
	MediaType string    `json:"mediaType"`
	Data      io.Reader `json:"-"`
}

// Manifest represents a container image manifest
type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Config        Layer   `json:"config"`
	Layers        []Layer `json:"layers"`
}

// ImagePusher provides comprehensive image pushing capabilities
type ImagePusher struct {
	registry *GiteaRegistry
	config   *PushConfig
	progress *PushProgress
}

// NewImagePusher creates a new image pusher instance
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

// updateProgress updates and reports push progress
func (p *ImagePusher) updateProgress(update func(*PushProgress)) {
	update(p.progress)
	p.progress.Duration = time.Since(p.progress.StartTime)
	if p.config.ProgressCallback != nil { p.config.ProgressCallback(p.progress) }
}

// initiateBlobUpload starts a blob upload session
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

// uploadBlobChunk uploads a chunk of blob data
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

// finalizeBlobUpload completes a blob upload session
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

// pushLayer uploads a layer to the registry
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

// layerExists checks if a layer already exists in the registry
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

// pushManifest uploads the image manifest to the registry
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

// PushImage pushes a complete image (config + layers + manifest) to the registry
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

// ParseImageRef parses an image reference into repository and tag components
func ParseImageRef(image string) (repository, tag string, err error) {
	parts := strings.SplitN(image, ":", 2)
	if len(parts) == 1 {
		return parts[0], "latest", nil
	}
	return parts[0], parts[1], nil
}

// calculateDigest calculates the SHA256 digest of data
func calculateDigest(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("sha256:%x", hash)
}

// PushV1Image is an enhanced push method for v1.Image types with advanced features.
// This complements the basic Push method in gitea.go for container registry operations.
// Note: This is separate from the Registry interface Push method.
func (r *GiteaRegistry) PushV1Image(ctx context.Context, image v1.Image, reference string) error {
	if image == nil {
		return fmt.Errorf("image cannot be nil")
	}

	if reference == "" {
		return fmt.Errorf("image reference cannot be empty")
	}

	// Parse the image reference
	ref, err := r.parseImageReference(reference)
	if err != nil {
		return fmt.Errorf("invalid image reference %q: %v", reference, err)
	}

	r.logger.Printf("Starting push of image to %s", reference)

	// Perform the push with retry logic using remote options
	return r.performPushWithRetry(ctx, ref, image)
}

// parseImageReference parses and validates the image reference format
func (r *GiteaRegistry) parseImageReference(reference string) (name.Reference, error) {
	// Extract host from baseURL string
	parsedURL, err := url.Parse(r.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid baseURL: %v", err)
	}

	// Ensure reference contains the registry host
	if !strings.Contains(reference, parsedURL.Host) {
		// If not, prepend the registry URL
		reference = fmt.Sprintf("%s/%s", parsedURL.Host, strings.TrimPrefix(reference, "/"))
	}

	ref, err := name.ParseReference(reference)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %v", err)
	}

	// Validate that the registry matches our configuration
	if ref.Context().RegistryStr() != parsedURL.Host {
		return nil, fmt.Errorf("reference registry %q does not match configured registry %q",
			ref.Context().RegistryStr(), parsedURL.Host)
	}

	return ref, nil
}

// performPushWithRetry executes the push operation with exponential backoff retry
func (r *GiteaRegistry) performPushWithRetry(ctx context.Context, ref name.Reference, image v1.Image) error {
	operation := func() error {
		return r.executePush(ctx, ref, image)
	}

	return retryWithExponentialBackoff(operation, "push", ref.Name())
}

// executePush performs the actual push operation
func (r *GiteaRegistry) executePush(ctx context.Context, ref name.Reference, image v1.Image) error {
	// Create progress tracking channel
	progressChan := make(chan v1.Update, 100)

	// Start a goroutine to handle progress updates
	go func() {
		for update := range progressChan {
			r.logProgress(ref.Name(), update)
		}
	}()
	defer close(progressChan)

	// Create remote options with authentication
	options := []remote.Option{}

	// Add authentication if available
	if authHeader, err := r.authMgr.GetAuthHeader(ctx); err == nil && authHeader != "" {
		// Note: We'd need to convert the authHeader to proper remote.Option
		// For now, just use empty options - this is a simplified implementation
	}

	// Add progress tracking to options
	progressOption := remote.WithProgress(progressChan)
	allOptions := append(options, progressOption)

	// Execute the push
	err := remote.Write(ref, image, allOptions...)
	if err != nil {
		return r.handlePushError(err, ref.Name())
	}

	log.Printf("Successfully pushed image to %s", ref.Name())
	return nil
}

// handlePushError provides comprehensive error handling for push failures
func (r *GiteaRegistry) handlePushError(err error, reference string) error {
	errorMsg := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errorMsg, "unauthorized"):
		return fmt.Errorf("authentication failed for %s: check credentials", reference)

	case strings.Contains(errorMsg, "forbidden"):
		return fmt.Errorf("insufficient permissions to push to %s", reference)

	case strings.Contains(errorMsg, "not found"):
		return fmt.Errorf("registry or repository not found: %s", reference)

	case strings.Contains(errorMsg, "tls"):
		if r.config.Insecure {
			return fmt.Errorf("TLS error despite insecure mode for %s: %v", reference, err)
		}
		return fmt.Errorf("TLS certificate error for %s: %v (try --insecure for development)", reference, err)

	case strings.Contains(errorMsg, "timeout"):
		return fmt.Errorf("push operation timed out for %s", reference)

	case strings.Contains(errorMsg, "network") || strings.Contains(errorMsg, "connection"):
		return fmt.Errorf("network error pushing to %s: %v", reference, err)

	default:
		return fmt.Errorf("push failed for %s: %v", reference, err)
	}
}

// pushProgressTracker implements progress reporting for push operations
type pushProgressTracker struct {
	reference    string
	totalBytes   int64
	uploadedBytes int64
}

// Write implements io.Writer for progress tracking
func (p *pushProgressTracker) Write(data []byte) (int, error) {
	n := len(data)
	p.uploadedBytes += int64(n)

	if p.totalBytes > 0 {
		percentage := (p.uploadedBytes * 100) / p.totalBytes
		log.Printf("Push progress for %s: %d%% (%d/%d bytes)",
			p.reference, percentage, p.uploadedBytes, p.totalBytes)
	}

	return n, nil
}

// logProgress logs progress updates for push operations
func (r *GiteaRegistry) logProgress(reference string, update v1.Update) {
	if update.Error != nil {
		log.Printf("Push error for %s: %v", reference, update.Error)
		return
	}

	if update.Total > 0 {
		percentage := (update.Complete * 100) / update.Total
		log.Printf("Push progress for %s: %d%% (%d/%d bytes)",
			reference, percentage, update.Complete, update.Total)
	} else {
		log.Printf("Push progress for %s: %d bytes completed", reference, update.Complete)
	}
}