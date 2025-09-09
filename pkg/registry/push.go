package registry

import (
	"context"
	"fmt"
	"log"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/name"
)

// Push uploads an image to the registry using the provided reference.
// Integrates with Phase 1 certificate infrastructure for secure TLS connections
// and provides comprehensive error handling with retry logic.
func (r *giteaRegistryImpl) Push(ctx context.Context, image v1.Image, reference string) error {
	if err := r.validateRegistry(); err != nil {
		return fmt.Errorf("registry validation failed: %v", err)
	}
	
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
	
	// Ensure authentication before push
	if !r.authn.IsAuthenticated() {
		if err := r.Authenticate(ctx); err != nil {
			return fmt.Errorf("authentication required for push: %v", err)
		}
	}
	
	// Get configured remote options with certificate handling
	options := r.GetRemoteOptions()
	
	log.Printf("Starting push of image to %s", reference)
	
	// Perform the push with retry logic
	return r.performPushWithRetry(ctx, ref, image, options)
}

// parseImageReference parses and validates the image reference format
func (r *giteaRegistryImpl) parseImageReference(reference string) (name.Reference, error) {
	// Ensure reference contains the registry host
	if !strings.Contains(reference, r.baseURL.Host) {
		// If not, prepend the registry URL
		reference = fmt.Sprintf("%s/%s", r.baseURL.Host, strings.TrimPrefix(reference, "/"))
	}
	
	ref, err := name.ParseReference(reference)
	if err != nil {
		return nil, fmt.Errorf("failed to parse reference: %v", err)
	}
	
	// Validate that the registry matches our configuration
	if ref.Context().RegistryStr() != r.baseURL.Host {
		return nil, fmt.Errorf("reference registry %q does not match configured registry %q", 
			ref.Context().RegistryStr(), r.baseURL.Host)
	}
	
	return ref, nil
}

// performPushWithRetry executes the push operation with exponential backoff retry
func (r *giteaRegistryImpl) performPushWithRetry(ctx context.Context, ref name.Reference, image v1.Image, options []remote.Option) error {
	operation := func() error {
		return r.executePush(ctx, ref, image, options)
	}
	
	return retryWithExponentialBackoff(operation, "push", ref.Name())
}

// executePush performs the actual push operation
func (r *giteaRegistryImpl) executePush(ctx context.Context, ref name.Reference, image v1.Image, options []remote.Option) error {
	// Create progress tracking channel
	progressChan := make(chan v1.Update, 100)
	
	// Start a goroutine to handle progress updates
	go func() {
		for update := range progressChan {
			r.logProgress(ref.Name(), update)
		}
	}()
	defer close(progressChan)
	
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
func (r *giteaRegistryImpl) handlePushError(err error, reference string) error {
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
func (r *giteaRegistryImpl) logProgress(reference string, update v1.Update) {
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