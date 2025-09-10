# Wave 1: Certificate Management Core - Implementation Plan

## 📌 Wave Overview

**Phase**: 1 - Certificate Infrastructure  
**Wave**: 1 - Certificate Management Core  
**Total Efforts**: 2  
**Total Estimated Lines**: 1,100  
**Parallelization**: YES - Both efforts can be implemented in parallel  
**Created**: 2025-09-06  
**Code Reviewer**: code-reviewer  

## 🎯 Wave Mission

Establish the foundational certificate extraction and trust management infrastructure, enabling secure communication with Gitea's self-signed certificate registry through go-containerregistry. This wave focuses on the core mechanics of certificate retrieval from Kind clusters and integration with the container registry client library.

## 🔄 Atomic PR Requirements (R220 Compliance)

### Wave-Level Atomic Design
```yaml
wave_implementation_atomic_design:
  parallel_pr_efforts:
    - effort_1_1_1: "Kind Certificate Extraction - can merge anytime"
    - effort_1_1_2: "Registry TLS Trust Integration - can merge anytime"
  sequential_pr_efforts: []  # No sequential dependencies
  feature_flags:
    wave_flag: "CERT_INFRASTRUCTURE_ENABLED"
    effort_flags:
      - "KIND_CERT_EXTRACTION_ENABLED"
      - "REGISTRY_TLS_TRUST_ENABLED"
  stub_plan:
    - stub: "MockCertExtractor"
      replaced_by: "E1.1.1"
      used_by: ["E1.1.2 tests"]
    - stub: "MockTrustManager"
      replaced_by: "E1.1.2"
      used_by: ["Phase 2 efforts"]
  merge_strategy:
    order_independent: true
    conflict_resolution: "Rebase on main"
  test_independence:
    each_pr_isolated: true
    flag_permutations: ["all off", "extraction only", "trust only", "all on"]
```

## 📦 Effort 1.1.1: Kind Certificate Extraction

### Metadata
**Branch**: `phase1/wave1/kind-cert-extraction`  
**Can Parallelize**: Yes  
**Parallel With**: [E1.1.2]  
**Size Estimate**: 500 lines  
**Dependencies**: None  
**Feature Flag**: `KIND_CERT_EXTRACTION_ENABLED`  

### Technical Requirements

#### Core Functionality
1. **Cluster Detection**: Verify Kind cluster existence and accessibility
2. **Pod Location**: Find Gitea pod in the Kind cluster
3. **Certificate Extraction**: Copy certificate from pod filesystem
4. **Local Storage**: Save certificate to well-known location
5. **Error Handling**: Graceful degradation when cluster unavailable

### Detailed File Structure

#### `pkg/certs/extractor.go` (150 lines)
```go
package certs

import (
    "context"
    "crypto/x509"
    "fmt"
    "os"
    "path/filepath"
)

// KindCertExtractor handles certificate extraction from Kind clusters
type KindCertExtractor struct {
    client     KindClient
    storage    CertificateStorage
    validator  CertValidator
    config     ExtractorConfig
}

// ExtractorConfig holds configuration for the extractor
type ExtractorConfig struct {
    ClusterName     string
    Namespace       string
    PodLabelSelector string
    CertPath        string // Path inside the pod
    Timeout         time.Duration
    RetryAttempts   int
}

// NewKindCertExtractor creates a new certificate extractor
func NewKindCertExtractor(config ExtractorConfig) (*KindCertExtractor, error) {
    // Initialize kubectl client
    // Validate configuration
    // Setup storage directory
    // Return configured extractor
}

// ExtractGiteaCert extracts the Gitea certificate from Kind cluster
func (e *KindCertExtractor) ExtractGiteaCert(ctx context.Context) (*x509.Certificate, error) {
    // 1. Check feature flag
    if !isFeatureEnabled("KIND_CERT_EXTRACTION_ENABLED") {
        return nil, ErrFeatureDisabled
    }
    
    // 2. Get cluster information
    clusterName, err := e.getClusterName()
    if err != nil {
        return nil, fmt.Errorf("failed to get cluster name: %w", err)
    }
    
    // 3. Find Gitea pod
    podName, err := e.findGiteaPod(ctx, clusterName)
    if err != nil {
        return nil, fmt.Errorf("failed to find Gitea pod: %w", err)
    }
    
    // 4. Extract certificate data
    certData, err := e.client.CopyFromPod(ctx, podName, e.config.CertPath)
    if err != nil {
        return nil, fmt.Errorf("failed to copy certificate: %w", err)
    }
    
    // 5. Parse certificate
    cert, err := parseCertificate(certData)
    if err != nil {
        return nil, fmt.Errorf("failed to parse certificate: %w", err)
    }
    
    // 6. Validate certificate
    if err := e.validator.ValidateCertificate(cert); err != nil {
        return nil, fmt.Errorf("certificate validation failed: %w", err)
    }
    
    // 7. Store locally
    if err := e.storage.Store("gitea", cert); err != nil {
        return nil, fmt.Errorf("failed to store certificate: %w", err)
    }
    
    return cert, nil
}

// GetClusterName returns the current Kind cluster name
func (e *KindCertExtractor) GetClusterName() (string, error) {
    if e.config.ClusterName != "" {
        return e.config.ClusterName, nil
    }
    return e.client.GetCurrentCluster()
}

// ValidateCertificate performs basic certificate validation
func (e *KindCertExtractor) ValidateCertificate(cert *x509.Certificate) error {
    // Check certificate is not nil
    // Check certificate is not expired
    // Check certificate has proper key usage
    // Return validation result
}

// StoreCertificate saves the certificate to local storage
func (e *KindCertExtractor) StoreCertificate(cert *x509.Certificate, path string) error {
    return e.storage.StoreAt(cert, path)
}
```

#### `pkg/certs/kind_client.go` (120 lines)
```go
package certs

import (
    "bytes"
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// KindClient interface for interacting with Kind clusters
type KindClient interface {
    GetCurrentCluster() (string, error)
    ListClusters() ([]string, error)
    GetPods(ctx context.Context, namespace, labelSelector string) ([]string, error)
    CopyFromPod(ctx context.Context, podName, path string) ([]byte, error)
    ExecInPod(ctx context.Context, podName string, command []string) (string, error)
}

// KubectlKindClient implements KindClient using kubectl
type KubectlKindClient struct {
    kubectlPath string
    kubeconfig  string
}

// NewKubectlKindClient creates a new kubectl-based Kind client
func NewKubectlKindClient() (*KubectlKindClient, error) {
    // Find kubectl binary
    // Verify kubectl is available
    // Setup kubeconfig path
    // Return configured client
}

// GetCurrentCluster returns the current Kind cluster name
func (c *KubectlKindClient) GetCurrentCluster() (string, error) {
    cmd := exec.Command("kind", "get", "clusters")
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("failed to list Kind clusters: %w", err)
    }
    
    clusters := strings.Split(strings.TrimSpace(string(output)), "\n")
    if len(clusters) == 0 {
        return "", ErrNoKindCluster
    }
    
    // Return first cluster (or use context to determine active)
    return clusters[0], nil
}

// ListClusters returns all available Kind clusters
func (c *KubectlKindClient) ListClusters() ([]string, error) {
    cmd := exec.Command("kind", "get", "clusters")
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to list clusters: %w", err)
    }
    
    return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}

// GetPods returns pods matching the label selector
func (c *KubectlKindClient) GetPods(ctx context.Context, namespace, labelSelector string) ([]string, error) {
    args := []string{"get", "pods", "-n", namespace}
    if labelSelector != "" {
        args = append(args, "-l", labelSelector)
    }
    args = append(args, "-o", "jsonpath={.items[*].metadata.name}")
    
    cmd := exec.CommandContext(ctx, c.kubectlPath, args...)
    output, err := cmd.Output()
    if err != nil {
        return nil, fmt.Errorf("failed to get pods: %w", err)
    }
    
    pods := strings.Fields(string(output))
    return pods, nil
}

// CopyFromPod copies a file from a pod
func (c *KubectlKindClient) CopyFromPod(ctx context.Context, podName, path string) ([]byte, error) {
    // Use kubectl cp to extract file
    cmd := exec.CommandContext(ctx, c.kubectlPath, "cp", 
        fmt.Sprintf("%s:%s", podName, path), "-")
    
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    
    if err := cmd.Run(); err != nil {
        return nil, fmt.Errorf("kubectl cp failed: %w, stderr: %s", err, stderr.String())
    }
    
    return stdout.Bytes(), nil
}

// ExecInPod executes a command in a pod
func (c *KubectlKindClient) ExecInPod(ctx context.Context, podName string, command []string) (string, error) {
    args := append([]string{"exec", podName, "--"}, command...)
    cmd := exec.CommandContext(ctx, c.kubectlPath, args...)
    
    output, err := cmd.Output()
    if err != nil {
        return "", fmt.Errorf("exec in pod failed: %w", err)
    }
    
    return string(output), nil
}
```

#### `pkg/certs/storage.go` (100 lines)
```go
package certs

import (
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "os"
    "path/filepath"
)

// CertificateStorage interface for certificate persistence
type CertificateStorage interface {
    Store(name string, cert *x509.Certificate) error
    StoreAt(cert *x509.Certificate, path string) error
    Load(name string) (*x509.Certificate, error)
    Exists(name string) bool
    Remove(name string) error
    ListCertificates() ([]string, error)
}

// LocalCertStorage implements file-based certificate storage
type LocalCertStorage struct {
    baseDir string
}

// NewLocalCertStorage creates a new local certificate storage
func NewLocalCertStorage(baseDir string) (*LocalCertStorage, error) {
    // Expand home directory if needed
    expandedDir := expandHomeDir(baseDir)
    
    // Create directory if it doesn't exist
    if err := os.MkdirAll(expandedDir, 0700); err != nil {
        return nil, fmt.Errorf("failed to create cert directory: %w", err)
    }
    
    return &LocalCertStorage{baseDir: expandedDir}, nil
}

// Store saves a certificate with the given name
func (s *LocalCertStorage) Store(name string, cert *x509.Certificate) error {
    path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))
    return s.StoreAt(cert, path)
}

// StoreAt saves a certificate at a specific path
func (s *LocalCertStorage) StoreAt(cert *x509.Certificate, path string) error {
    // Encode certificate to PEM
    pemBlock := &pem.Block{
        Type:  "CERTIFICATE",
        Bytes: cert.Raw,
    }
    
    pemData := pem.EncodeToMemory(pemBlock)
    if pemData == nil {
        return fmt.Errorf("failed to encode certificate")
    }
    
    // Write to file with secure permissions
    if err := ioutil.WriteFile(path, pemData, 0600); err != nil {
        return fmt.Errorf("failed to write certificate: %w", err)
    }
    
    return nil
}

// Load retrieves a certificate by name
func (s *LocalCertStorage) Load(name string) (*x509.Certificate, error) {
    path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))
    
    // Read certificate file
    pemData, err := ioutil.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, ErrCertNotFound
        }
        return nil, fmt.Errorf("failed to read certificate: %w", err)
    }
    
    // Parse PEM block
    block, _ := pem.Decode(pemData)
    if block == nil {
        return nil, fmt.Errorf("failed to decode PEM block")
    }
    
    // Parse certificate
    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse certificate: %w", err)
    }
    
    return cert, nil
}

// Exists checks if a certificate exists
func (s *LocalCertStorage) Exists(name string) bool {
    path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))
    _, err := os.Stat(path)
    return err == nil
}

// Remove deletes a certificate
func (s *LocalCertStorage) Remove(name string) error {
    path := filepath.Join(s.baseDir, fmt.Sprintf("%s.pem", name))
    return os.Remove(path)
}

// ListCertificates returns all stored certificate names
func (s *LocalCertStorage) ListCertificates() ([]string, error) {
    // Read directory
    // Filter .pem files
    // Return names without extension
}

// expandHomeDir expands ~ to user home directory
func expandHomeDir(path string) string {
    if strings.HasPrefix(path, "~/") {
        home, _ := os.UserHomeDir()
        return filepath.Join(home, path[2:])
    }
    return path
}
```

#### `pkg/certs/errors.go` (50 lines)
```go
package certs

import "errors"

// Certificate operation errors
var (
    // Extraction errors
    ErrNoKindCluster = errors.New("no Kind cluster found")
    ErrGiteaPodNotFound = errors.New("Gitea pod not found in cluster")
    ErrCertNotInPod = errors.New("certificate not found in pod")
    ErrInvalidCertData = errors.New("invalid certificate data")
    
    // Storage errors
    ErrCertNotFound = errors.New("certificate not found in storage")
    ErrStoragePermission = errors.New("insufficient permissions for certificate storage")
    ErrStorageFull = errors.New("certificate storage is full")
    
    // Validation errors
    ErrCertExpired = errors.New("certificate has expired")
    ErrCertNotYetValid = errors.New("certificate is not yet valid")
    ErrCertInvalidKeyUsage = errors.New("certificate has invalid key usage")
    ErrCertSelfSigned = errors.New("certificate is self-signed")
    
    // Feature flag errors
    ErrFeatureDisabled = errors.New("certificate extraction feature is disabled")
)

// CertError wraps certificate errors with context
type CertError struct {
    Op      string // Operation that failed
    Kind    string // Kind of error
    Err     error  // Underlying error
    Context map[string]string // Additional context
}

// Error implements the error interface
func (e *CertError) Error() string {
    if e.Context != nil && len(e.Context) > 0 {
        return fmt.Sprintf("%s: %s: %v (context: %v)", e.Op, e.Kind, e.Err, e.Context)
    }
    return fmt.Sprintf("%s: %s: %v", e.Op, e.Kind, e.Err)
}

// Unwrap returns the underlying error
func (e *CertError) Unwrap() error {
    return e.Err
}

// NewCertError creates a new certificate error
func NewCertError(op, kind string, err error) *CertError {
    return &CertError{
        Op:   op,
        Kind: kind,
        Err:  err,
    }
}
```

#### `pkg/certs/helpers.go` (80 lines)
```go
package certs

import (
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "os"
    "time"
)

// parseCertificate parses certificate from PEM data
func parseCertificate(pemData []byte) (*x509.Certificate, error) {
    block, _ := pem.Decode(pemData)
    if block == nil {
        return nil, fmt.Errorf("failed to parse PEM block")
    }
    
    if block.Type != "CERTIFICATE" {
        return nil, fmt.Errorf("PEM block is not a certificate: %s", block.Type)
    }
    
    cert, err := x509.ParseCertificate(block.Bytes)
    if err != nil {
        return nil, fmt.Errorf("failed to parse certificate: %w", err)
    }
    
    return cert, nil
}

// isFeatureEnabled checks if a feature flag is enabled
func isFeatureEnabled(flag string) bool {
    // Check environment variable
    envVar := fmt.Sprintf("IDPBUILDER_%s", flag)
    value := os.Getenv(envVar)
    
    // Parse boolean value
    return value == "true" || value == "1" || value == "enabled"
}

// findGiteaPod locates the Gitea pod in the cluster
func (e *KindCertExtractor) findGiteaPod(ctx context.Context, clusterName string) (string, error) {
    // Default namespace and labels for Gitea
    namespace := "gitea"
    labelSelector := "app=gitea"
    
    // Override from config if provided
    if e.config.Namespace != "" {
        namespace = e.config.Namespace
    }
    if e.config.PodLabelSelector != "" {
        labelSelector = e.config.PodLabelSelector
    }
    
    // Get pods matching selector
    pods, err := e.client.GetPods(ctx, namespace, labelSelector)
    if err != nil {
        return "", fmt.Errorf("failed to get pods: %w", err)
    }
    
    if len(pods) == 0 {
        return "", ErrGiteaPodNotFound
    }
    
    // Return first matching pod
    return pods[0], nil
}

// getClusterName retrieves the cluster name with fallback
func (e *KindCertExtractor) getClusterName() (string, error) {
    // Try configured name first
    if e.config.ClusterName != "" {
        return e.config.ClusterName, nil
    }
    
    // Fall back to current cluster
    return e.client.GetCurrentCluster()
}

// validateCertificateExpiry checks if certificate is valid time-wise
func validateCertificateExpiry(cert *x509.Certificate) error {
    now := time.Now()
    
    if now.Before(cert.NotBefore) {
        return ErrCertNotYetValid
    }
    
    if now.After(cert.NotAfter) {
        return ErrCertExpired
    }
    
    return nil
}
```

### Test Requirements for E1.1.1

#### `pkg/certs/extractor_test.go` (150 lines)
- Test successful certificate extraction from Kind cluster
- Test handling of missing Kind cluster
- Test handling of missing Gitea pod
- Test certificate parsing errors
- Test storage permission errors
- Test feature flag disabled scenario
- Test timeout and retry logic
- Mock kubectl commands for unit testing

#### `pkg/certs/kind_client_test.go` (100 lines)
- Test cluster detection with multiple clusters
- Test pod listing with various selectors
- Test file copy from pod
- Test command execution in pod
- Test error handling for kubectl failures
- Mock exec.Command for testing

#### `pkg/certs/storage_test.go` (80 lines)
- Test certificate storage and retrieval
- Test handling of non-existent certificates
- Test permission errors (using temp directories)
- Test certificate listing
- Test certificate removal
- Test home directory expansion

## 📦 Effort 1.1.2: Registry TLS Trust Integration

### Metadata
**Branch**: `phase1/wave1/registry-tls-trust`  
**Can Parallelize**: Yes  
**Parallel With**: [E1.1.1]  
**Size Estimate**: 600 lines  
**Dependencies**: None (can use mock certificates for testing)  
**Feature Flag**: `REGISTRY_TLS_TRUST_ENABLED`  

### Technical Requirements

#### Core Functionality
1. **CA Pool Management**: Load custom CAs into x509.CertPool
2. **Transport Configuration**: Configure go-containerregistry with custom TLS
3. **Certificate Rotation**: Support runtime certificate updates
4. **Insecure Mode**: Provide explicit --insecure flag override
5. **Error Reporting**: Clear messages for TLS configuration issues

### Detailed File Structure

#### `pkg/certs/trust.go` (180 lines)
```go
package certs

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "sync"
    "github.com/google/go-containerregistry/pkg/v1/remote"
)

// TrustStoreManager manages trusted certificates for registries
type TrustStoreManager struct {
    mu              sync.RWMutex
    trustedCerts    map[string][]*x509.Certificate
    insecureRegistries map[string]bool
    systemPool      *x509.CertPool
    storage         CertificateStorage
}

// NewTrustStoreManager creates a new trust store manager
func NewTrustStoreManager(storage CertificateStorage) (*TrustStoreManager, error) {
    // Load system certificate pool
    systemPool, err := x509.SystemCertPool()
    if err != nil {
        // Fall back to empty pool if system pool unavailable
        systemPool = x509.NewCertPool()
    }
    
    return &TrustStoreManager{
        trustedCerts:       make(map[string][]*x509.Certificate),
        insecureRegistries: make(map[string]bool),
        systemPool:        systemPool,
        storage:           storage,
    }, nil
}

// AddCertificate adds a trusted certificate for a registry
func (m *TrustStoreManager) AddCertificate(registry string, cert *x509.Certificate) error {
    // Check feature flag
    if !isFeatureEnabled("REGISTRY_TLS_TRUST_ENABLED") {
        return ErrFeatureDisabled
    }
    
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Validate certificate
    if err := validateCertificateExpiry(cert); err != nil {
        return fmt.Errorf("certificate validation failed: %w", err)
    }
    
    // Add to trusted certs
    if m.trustedCerts[registry] == nil {
        m.trustedCerts[registry] = make([]*x509.Certificate, 0)
    }
    m.trustedCerts[registry] = append(m.trustedCerts[registry], cert)
    
    // Persist to storage
    storageKey := fmt.Sprintf("registry_%s", sanitizeRegistryName(registry))
    if err := m.storage.Store(storageKey, cert); err != nil {
        return fmt.Errorf("failed to persist certificate: %w", err)
    }
    
    return nil
}

// RemoveCertificate removes a trusted certificate for a registry
func (m *TrustStoreManager) RemoveCertificate(registry string) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    delete(m.trustedCerts, registry)
    
    // Remove from storage
    storageKey := fmt.Sprintf("registry_%s", sanitizeRegistryName(registry))
    return m.storage.Remove(storageKey)
}

// SetInsecureRegistry marks a registry as insecure (skip TLS verification)
func (m *TrustStoreManager) SetInsecureRegistry(registry string, insecure bool) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    if insecure {
        // Log security decision
        logSecurityDecision("INSECURE_REGISTRY", registry, "User explicitly set --insecure flag")
        m.insecureRegistries[registry] = true
    } else {
        delete(m.insecureRegistries, registry)
    }
    
    return nil
}

// GetTrustedCerts returns all trusted certificates for a registry
func (m *TrustStoreManager) GetTrustedCerts(registry string) ([]*x509.Certificate, error) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    certs := m.trustedCerts[registry]
    if certs == nil {
        // Try to load from storage
        storageKey := fmt.Sprintf("registry_%s", sanitizeRegistryName(registry))
        if cert, err := m.storage.Load(storageKey); err == nil {
            certs = []*x509.Certificate{cert}
            // Cache for future use
            m.trustedCerts[registry] = certs
        }
    }
    
    return certs, nil
}

// GetTransportOptions returns go-containerregistry options for a registry
func (m *TrustStoreManager) GetTransportOptions(registry string) ([]remote.Option, error) {
    // Check if registry is marked as insecure
    if m.isInsecure(registry) {
        transport := &http.Transport{
            TLSClientConfig: &tls.Config{
                InsecureSkipVerify: true,
            },
        }
        return []remote.Option{remote.WithTransport(transport)}, nil
    }
    
    // Build custom CA pool
    pool := x509.NewCertPool()
    
    // Add system certificates
    pool.AppendCertsFromPEM(m.systemPool.Subjects())
    
    // Add custom certificates for this registry
    certs, err := m.GetTrustedCerts(registry)
    if err != nil {
        return nil, fmt.Errorf("failed to get trusted certs: %w", err)
    }
    
    for _, cert := range certs {
        pool.AddCert(cert)
    }
    
    // Create transport with custom CA pool
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            RootCAs: pool,
        },
    }
    
    return []remote.Option{remote.WithTransport(transport)}, nil
}

// ReloadCertificates reloads certificates from storage
func (m *TrustStoreManager) ReloadCertificates() error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    // Clear current cache
    m.trustedCerts = make(map[string][]*x509.Certificate)
    
    // Reload from storage
    certNames, err := m.storage.ListCertificates()
    if err != nil {
        return fmt.Errorf("failed to list certificates: %w", err)
    }
    
    for _, name := range certNames {
        if strings.HasPrefix(name, "registry_") {
            cert, err := m.storage.Load(name)
            if err != nil {
                // Log but continue
                fmt.Printf("Warning: failed to load certificate %s: %v\n", name, err)
                continue
            }
            
            // Extract registry name
            registry := strings.TrimPrefix(name, "registry_")
            if m.trustedCerts[registry] == nil {
                m.trustedCerts[registry] = make([]*x509.Certificate, 0)
            }
            m.trustedCerts[registry] = append(m.trustedCerts[registry], cert)
        }
    }
    
    return nil
}

// isInsecure checks if a registry is marked as insecure
func (m *TrustStoreManager) isInsecure(registry string) bool {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.insecureRegistries[registry]
}
```

#### `pkg/certs/transport.go` (150 lines)
```go
package certs

import (
    "crypto/tls"
    "crypto/x509"
    "fmt"
    "net/http"
    "time"
)

// TransportConfigurer configures HTTP transports with custom TLS
type TransportConfigurer interface {
    ConfigureTransport(baseTransport http.RoundTripper, tlsConfig *tls.Config) http.RoundTripper
    GetTLSConfig(registry string) (*tls.Config, error)
    SetInsecureSkipVerify(skip bool)
}

// DefaultTransportConfigurer implements TransportConfigurer
type DefaultTransportConfigurer struct {
    trustManager *TrustStoreManager
    insecure     bool
    timeout      time.Duration
}

// NewDefaultTransportConfigurer creates a new transport configurer
func NewDefaultTransportConfigurer(trustManager *TrustStoreManager) *DefaultTransportConfigurer {
    return &DefaultTransportConfigurer{
        trustManager: trustManager,
        timeout:      30 * time.Second,
    }
}

// ConfigureTransport wraps a base transport with custom TLS configuration
func (c *DefaultTransportConfigurer) ConfigureTransport(baseTransport http.RoundTripper, tlsConfig *tls.Config) http.RoundTripper {
    if baseTransport == nil {
        baseTransport = http.DefaultTransport
    }
    
    // Type assert to get underlying transport
    if transport, ok := baseTransport.(*http.Transport); ok {
        // Clone to avoid modifying shared transport
        cloned := transport.Clone()
        cloned.TLSClientConfig = tlsConfig
        
        // Set reasonable timeouts
        cloned.TLSHandshakeTimeout = 10 * time.Second
        cloned.ResponseHeaderTimeout = c.timeout
        
        return cloned
    }
    
    // If not http.Transport, wrap with custom round tripper
    return &tlsRoundTripper{
        base:      baseTransport,
        tlsConfig: tlsConfig,
    }
}

// GetTLSConfig builds a TLS configuration for a registry
func (c *DefaultTransportConfigurer) GetTLSConfig(registry string) (*tls.Config, error) {
    // Check for insecure mode
    if c.insecure || c.trustManager.isInsecure(registry) {
        logSecurityDecision("TLS_SKIP_VERIFY", registry, "Insecure mode enabled")
        return &tls.Config{
            InsecureSkipVerify: true,
        }, nil
    }
    
    // Build CA pool
    pool, err := c.buildCAPool(registry)
    if err != nil {
        return nil, fmt.Errorf("failed to build CA pool: %w", err)
    }
    
    config := &tls.Config{
        RootCAs:            pool,
        MinVersion:         tls.VersionTLS12,
        PreferServerCipherSuites: true,
    }
    
    // Add SNI if registry looks like a hostname
    if !strings.Contains(registry, ":") {
        config.ServerName = registry
    } else {
        // Extract hostname from registry:port
        host, _, err := net.SplitHostPort(registry)
        if err == nil {
            config.ServerName = host
        }
    }
    
    return config, nil
}

// SetInsecureSkipVerify sets global insecure mode
func (c *DefaultTransportConfigurer) SetInsecureSkipVerify(skip bool) {
    c.insecure = skip
    if skip {
        logSecurityDecision("GLOBAL_INSECURE", "all", "Global insecure mode enabled")
    }
}

// buildCAPool creates a certificate pool for a registry
func (c *DefaultTransportConfigurer) buildCAPool(registry string) (*x509.CertPool, error) {
    // Start with system pool
    pool, err := x509.SystemCertPool()
    if err != nil {
        // Fall back to empty pool
        pool = x509.NewCertPool()
    }
    
    // Add custom certificates
    certs, err := c.trustManager.GetTrustedCerts(registry)
    if err != nil {
        return nil, err
    }
    
    for _, cert := range certs {
        pool.AddCert(cert)
    }
    
    return pool, nil
}

// tlsRoundTripper wraps a RoundTripper with TLS configuration
type tlsRoundTripper struct {
    base      http.RoundTripper
    tlsConfig *tls.Config
}

// RoundTrip implements http.RoundTripper
func (t *tlsRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
    // Only modify HTTPS requests
    if req.URL.Scheme != "https" {
        return t.base.RoundTrip(req)
    }
    
    // Clone request to avoid modifying original
    reqClone := req.Clone(req.Context())
    
    // Apply TLS config through context or transport modification
    // This is a simplified version - actual implementation would
    // need to handle transport configuration properly
    
    return t.base.RoundTrip(reqClone)
}
```

#### `pkg/certs/pool.go` (120 lines)
```go
package certs

import (
    "crypto/x509"
    "encoding/pem"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "sync"
)

// CertPoolManager manages certificate pools with caching
type CertPoolManager struct {
    mu       sync.RWMutex
    pools    map[string]*x509.CertPool
    storage  CertificateStorage
}

// NewCertPoolManager creates a new certificate pool manager
func NewCertPoolManager(storage CertificateStorage) *CertPoolManager {
    return &CertPoolManager{
        pools:   make(map[string]*x509.CertPool),
        storage: storage,
    }
}

// GetPool returns a certificate pool for a registry
func (m *CertPoolManager) GetPool(registry string) (*x509.CertPool, error) {
    m.mu.RLock()
    pool, exists := m.pools[registry]
    m.mu.RUnlock()
    
    if exists {
        return pool, nil
    }
    
    // Build new pool
    pool, err := m.buildPool(registry)
    if err != nil {
        return nil, err
    }
    
    // Cache for future use
    m.mu.Lock()
    m.pools[registry] = pool
    m.mu.Unlock()
    
    return pool, nil
}

// buildPool creates a new certificate pool for a registry
func (m *CertPoolManager) buildPool(registry string) (*x509.CertPool, error) {
    // Start with system pool
    pool, err := x509.SystemCertPool()
    if err != nil {
        pool = x509.NewCertPool()
    }
    
    // Load registry-specific certificates
    certKey := fmt.Sprintf("registry_%s", sanitizeRegistryName(registry))
    if cert, err := m.storage.Load(certKey); err == nil {
        pool.AddCert(cert)
    }
    
    // Load any additional CA certificates
    caDir := filepath.Join(getConfigDir(), "ca-certificates")
    if certs, err := loadCertificatesFromDir(caDir); err == nil {
        for _, cert := range certs {
            pool.AddCert(cert)
        }
    }
    
    return pool, nil
}

// AddCertificateToPool adds a certificate to a registry's pool
func (m *CertPoolManager) AddCertificateToPool(registry string, cert *x509.Certificate) error {
    pool, err := m.GetPool(registry)
    if err != nil {
        return err
    }
    
    pool.AddCert(cert)
    
    // Invalidate cache to force rebuild
    m.mu.Lock()
    delete(m.pools, registry)
    m.mu.Unlock()
    
    return nil
}

// loadCertificatesFromDir loads all certificates from a directory
func loadCertificatesFromDir(dir string) ([]*x509.Certificate, error) {
    files, err := ioutil.ReadDir(dir)
    if err != nil {
        return nil, err
    }
    
    var certs []*x509.Certificate
    for _, file := range files {
        if filepath.Ext(file.Name()) != ".pem" {
            continue
        }
        
        path := filepath.Join(dir, file.Name())
        pemData, err := ioutil.ReadFile(path)
        if err != nil {
            continue // Skip files we can't read
        }
        
        // Parse all certificates in the file
        for len(pemData) > 0 {
            var block *pem.Block
            block, pemData = pem.Decode(pemData)
            if block == nil {
                break
            }
            
            if block.Type != "CERTIFICATE" {
                continue
            }
            
            cert, err := x509.ParseCertificate(block.Bytes)
            if err != nil {
                continue
            }
            
            certs = append(certs, cert)
        }
    }
    
    return certs, nil
}

// ClearCache clears all cached certificate pools
func (m *CertPoolManager) ClearCache() {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.pools = make(map[string]*x509.CertPool)
}
```

#### `pkg/certs/config.go` (80 lines)
```go
package certs

import (
    "os"
    "path/filepath"
    "time"
)

// TLSConfig holds TLS configuration for registry connections
type TLSConfig struct {
    // Registry-specific settings
    Registry           string
    InsecureSkipVerify bool
    CAFile             string
    CertFile           string
    KeyFile            string
    
    // Global settings
    MinVersion         uint16
    PreferServerCiphers bool
    SessionCache       bool
    
    // Timeouts
    HandshakeTimeout   time.Duration
    
    // Certificate validation
    ValidateHostname   bool
    AllowExpiredCerts  bool
}

// DefaultTLSConfig returns a secure default TLS configuration
func DefaultTLSConfig() *TLSConfig {
    return &TLSConfig{
        MinVersion:          tls.VersionTLS12,
        PreferServerCiphers: true,
        SessionCache:        true,
        HandshakeTimeout:    10 * time.Second,
        ValidateHostname:    true,
        AllowExpiredCerts:   false,
    }
}

// LoadFromEnv loads TLS configuration from environment variables
func (c *TLSConfig) LoadFromEnv() {
    // Check for insecure mode
    if os.Getenv("IDPBUILDER_TLS_INSECURE") == "true" {
        c.InsecureSkipVerify = true
    }
    
    // Load CA file path
    if caFile := os.Getenv("IDPBUILDER_CA_FILE"); caFile != "" {
        c.CAFile = caFile
    }
    
    // Load client certificate paths
    if certFile := os.Getenv("IDPBUILDER_CERT_FILE"); certFile != "" {
        c.CertFile = certFile
    }
    if keyFile := os.Getenv("IDPBUILDER_KEY_FILE"); keyFile != "" {
        c.KeyFile = keyFile
    }
}

// getConfigDir returns the configuration directory path
func getConfigDir() string {
    // Check environment variable first
    if dir := os.Getenv("IDPBUILDER_CONFIG_DIR"); dir != "" {
        return dir
    }
    
    // Default to ~/.idpbuilder
    home, err := os.UserHomeDir()
    if err != nil {
        return ".idpbuilder"
    }
    
    return filepath.Join(home, ".idpbuilder")
}

// sanitizeRegistryName converts registry URL to safe filename
func sanitizeRegistryName(registry string) string {
    // Replace problematic characters
    safe := strings.ReplaceAll(registry, ":", "_")
    safe = strings.ReplaceAll(safe, "/", "_")
    safe = strings.ReplaceAll(safe, ".", "_")
    return safe
}
```

#### `pkg/certs/logging.go` (70 lines)
```go
package certs

import (
    "fmt"
    "log"
    "os"
    "time"
)

// SecurityLogger logs security-relevant decisions
type SecurityLogger struct {
    logger *log.Logger
    file   *os.File
}

var securityLogger *SecurityLogger

// InitSecurityLogger initializes the security logger
func InitSecurityLogger() error {
    logPath := filepath.Join(getConfigDir(), "security.log")
    
    file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        return fmt.Errorf("failed to open security log: %w", err)
    }
    
    securityLogger = &SecurityLogger{
        logger: log.New(file, "[SECURITY] ", log.LstdFlags|log.LUTC),
        file:   file,
    }
    
    return nil
}

// logSecurityDecision logs a security-relevant decision
func logSecurityDecision(decision, target, reason string) {
    if securityLogger == nil {
        // Fall back to stderr if logger not initialized
        fmt.Fprintf(os.Stderr, "[SECURITY] %s: %s - %s (reason: %s)\n",
            time.Now().UTC().Format(time.RFC3339),
            decision, target, reason)
        return
    }
    
    securityLogger.logger.Printf("%s: %s - %s", decision, target, reason)
}

// CloseSecurityLogger closes the security log file
func CloseSecurityLogger() {
    if securityLogger != nil && securityLogger.file != nil {
        securityLogger.file.Close()
    }
}

// SecurityAuditEntry represents a security audit log entry
type SecurityAuditEntry struct {
    Timestamp   time.Time
    Decision    string
    Target      string
    Reason      string
    User        string
    Success     bool
}

// LogAuditEntry logs a structured audit entry
func LogAuditEntry(entry SecurityAuditEntry) {
    if entry.Timestamp.IsZero() {
        entry.Timestamp = time.Now().UTC()
    }
    
    if entry.User == "" {
        entry.User = os.Getenv("USER")
    }
    
    msg := fmt.Sprintf("AUDIT: user=%s decision=%s target=%s success=%v reason=%s",
        entry.User, entry.Decision, entry.Target, entry.Success, entry.Reason)
    
    logSecurityDecision("AUDIT", entry.Target, msg)
}
```

### Test Requirements for E1.1.2

#### `pkg/certs/trust_test.go` (150 lines)
- Test adding/removing certificates for registries
- Test insecure registry configuration
- Test certificate reloading
- Test transport options generation
- Test concurrent access safety
- Test feature flag disabled scenario

#### `pkg/certs/transport_test.go` (120 lines)
- Test TLS configuration building
- Test insecure mode behavior
- Test custom CA pool inclusion
- Test hostname verification
- Test transport wrapping
- Mock HTTP transports for testing

#### `pkg/certs/pool_test.go` (80 lines)
- Test certificate pool creation
- Test pool caching behavior
- Test loading certificates from directory
- Test system pool fallback
- Test certificate addition to pools

## 🔄 Integration Points

### Between E1.1.1 and E1.1.2
While these efforts can be implemented in parallel, they share some common types:
- Both use `CertificateStorage` interface
- Both reference certificate validation functions
- Both check feature flags

### Integration Strategy
1. **Mock Interfaces**: E1.1.2 can use mock `CertificateStorage` for testing
2. **Shared Types**: Define interfaces in a common `types.go` file
3. **Feature Flags**: Both efforts check their respective flags independently
4. **Storage Path**: Agree on `~/.idpbuilder/certs/` as standard location

## 📊 Size Tracking

### Measurement Command
```bash
# From project root, use the line counter tool
PROJECT_ROOT=$(pwd)
while [ "$PROJECT_ROOT" != "/" ]; do 
    [ -f "$PROJECT_ROOT/orchestrator-state.yaml" ] && break
    PROJECT_ROOT=$(dirname "$PROJECT_ROOT")
done

# For each effort branch
cd efforts/phase1/wave1/kind-cert-extraction
$PROJECT_ROOT/tools/line-counter.sh

cd efforts/phase1/wave1/registry-tls-trust  
$PROJECT_ROOT/tools/line-counter.sh
```

### Size Targets
- **E1.1.1**: 500 lines (excluding tests)
- **E1.1.2**: 600 lines (excluding tests)
- **Warning Threshold**: 700 lines
- **Hard Limit**: 800 lines

## 🧪 Testing Strategy

### Unit Test Coverage Requirements
- Minimum: 80% coverage
- Target: 90% coverage
- Must test all error paths
- Must test concurrent access where applicable

### Integration Test Requirements
- Test with real Kind cluster (CI environment)
- Test with mock registry server
- Test certificate rotation scenarios
- Test fallback mechanisms

### Test Execution
```bash
# Unit tests
go test ./pkg/certs/... -cover

# Integration tests (requires Kind cluster)
go test ./tests/integration/certs/... -tags=integration

# Coverage report
go test ./pkg/certs/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## 🚀 Deployment Considerations

### Environment Variables
```bash
# Feature flags
export IDPBUILDER_CERT_INFRASTRUCTURE_ENABLED=true
export IDPBUILDER_KIND_CERT_EXTRACTION_ENABLED=true
export IDPBUILDER_REGISTRY_TLS_TRUST_ENABLED=true

# Configuration
export IDPBUILDER_CONFIG_DIR=~/.idpbuilder
export IDPBUILDER_TLS_INSECURE=false  # Only set true for testing

# Certificate paths (optional)
export IDPBUILDER_CA_FILE=/path/to/ca.pem
export IDPBUILDER_CERT_FILE=/path/to/cert.pem
export IDPBUILDER_KEY_FILE=/path/to/key.pem
```

### File System Layout
```
~/.idpbuilder/
├── certs/
│   ├── gitea.pem                    # Extracted Gitea certificate
│   ├── registry_gitea_10_0_0_1.pem  # Registry-specific cert
│   └── ca-certificates/              # Additional CA certs
│       └── custom-ca.pem
├── security.log                     # Security audit log
└── config.yaml                      # Future: configuration file
```

## 📝 Implementation Guidelines for SW Engineers

### Development Order
1. **Start with interfaces and types** - Define contracts first
2. **Implement storage layer** - Foundation for both efforts
3. **Build extraction (E1.1.1)** or trust management (E1.1.2)
4. **Add error handling and logging**
5. **Write comprehensive tests**
6. **Document public APIs**

### Code Quality Requirements
- **NO TODO COMMENTS** - All code must be complete
- **NO STUB IMPLEMENTATIONS** - No "not implemented" errors
- **PROPER ERROR HANDLING** - Wrap errors with context
- **CLEAR LOGGING** - Especially for security decisions
- **CONCURRENT SAFETY** - Use mutexes where needed

### Pull Request Checklist
- [ ] All tests passing with >80% coverage
- [ ] No linting errors
- [ ] Feature flags properly checked
- [ ] Security logging implemented
- [ ] Public APIs documented
- [ ] Line count under 800 (use line-counter.sh)
- [ ] Can merge independently to main

## 🔒 Security Considerations

### Critical Security Rules
1. **NEVER silently bypass certificate validation**
2. **ALWAYS require explicit --insecure flag**
3. **LOG all security-relevant decisions**
4. **PROTECT certificate files with 0600 permissions**
5. **VALIDATE certificate chains properly**

### Security Audit Points
- Certificate extraction must verify pod identity
- Trust configuration must not leak credentials
- Insecure mode must show clear warnings
- All bypasses must be logged with reason

## 🎬 Demo Requirements (R330)

### Demo Objectives
1. Demonstrate Gitea registry authentication with token management
2. Showcase repository listing and existence checking
3. Illustrate TLS configuration with custom certificates
4. Display flexible remote options configuration
5. Prove foundation for Split-002 operations

### Demo Deliverables
- **demo-features.sh** (executable) - Main demo script with 4 scenarios
- **DEMO.md** (documentation) - Setup guide and validation steps
- **test-data/** (sample files) - CA certificates and configurations

### Demo Scenarios Summary
1. **Authentication** - Token-based Gitea authentication
2. **List Repos** - Discover available repositories
3. **Check Existence** - Verify repository presence
4. **TLS Config** - Custom CA and insecure mode demo

## 📚 References

### External Dependencies
- [go-containerregistry](https://github.com/google/go-containerregistry) v0.19.0
- [Kind](https://kind.sigs.k8s.io/) (latest stable)
- Standard library: crypto/x509, crypto/tls

### Related Documentation
- Phase 1 Architecture Plan: `phase-plans/PHASE-1-PLAN.md`
- Certificate Handling Guide: (to be created)
- Security Best Practices: (to be created)
- Demo Retrofit Plan: `DEMO-RETROFIT-PLAN.md`
- Split-002 Integration: `../gitea-client-split-002/DEMO-RETROFIT-PLAN.md`

---

**Wave Plan Version**: 1.0  
**Created**: 2025-09-06  
**Code Reviewer**: code-reviewer  
**Status**: Ready for Implementation

**Remember**: 
- Both efforts can work in PARALLEL
- Each PR must be INDEPENDENTLY mergeable
- NO stub implementations allowed
- Use line-counter.sh for ALL size measurements
- Feature flags enable gradual rollout