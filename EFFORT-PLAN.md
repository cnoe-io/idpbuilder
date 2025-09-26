# Effort 1.1.3: TLS Configuration Implementation Plan

## 🚨 CRITICAL EFFORT METADATA
**Branch**: `phase1-wave1-effort-1.1.3-tls-config`
**Can Parallelize**: Yes
**Parallel With**: [Effort 1.1.1, Effort 1.1.2]
**Size Estimate**: 180 lines (well under 800 limit)
**Dependencies**: None (foundational effort)

## Overview
- **Effort**: Add TLS configuration support for push command
- **Phase**: 1, Wave: 1
- **Estimated Size**: ~180 lines
- **Implementation Time**: 2-3 hours

## Purpose
Implement TLS configuration handling to support secure connections to the Gitea registry, including the ability to bypass certificate validation for self-signed certificates using the `--insecure` flag.

## File Structure
```
efforts/phase1/wave1/effort-1.1.3-tls-config/
├── cmd/
│   └── push.go                    # Add --insecure flag (~30 lines)
├── pkg/
│   └── tls/
│       ├── config.go              # TLS config factory (~80 lines)
│       └── config_test.go         # Unit tests (~70 lines)
└── work-log.md                    # Progress tracking
```

## Implementation Steps

### Step 1: Add --insecure Flag to Push Command (30 lines)
**File**: `cmd/push.go`
```go
// Add to existing push command initialization
func init() {
    // Add insecure flag
    pushCmd.Flags().Bool("insecure", false,
        "Skip TLS certificate verification (use for self-signed certificates)")
}
```
- Add flag definition
- Add flag description for help text
- Ensure flag is properly bound to viper if using configuration

### Step 2: Create TLS Configuration Factory (80 lines)
**File**: `pkg/tls/config.go`
```go
package tls

import (
    "crypto/tls"
    "net/http"
)

// Config holds TLS configuration options
type Config struct {
    InsecureSkipVerify bool
}

// NewConfig creates a new TLS configuration
func NewConfig(insecure bool) *Config {
    return &Config{
        InsecureSkipVerify: insecure,
    }
}

// ToTLSConfig converts to standard crypto/tls.Config
func (c *Config) ToTLSConfig() *tls.Config {
    return &tls.Config{
        InsecureSkipVerify: c.InsecureSkipVerify,
    }
}

// ApplyToHTTPClient applies TLS config to an HTTP client
func (c *Config) ApplyToHTTPClient(client *http.Client) {
    if client.Transport == nil {
        client.Transport = &http.Transport{}
    }

    if transport, ok := client.Transport.(*http.Transport); ok {
        transport.TLSClientConfig = c.ToTLSConfig()
    }
}

// ApplyToTransport applies TLS config to an HTTP transport
func (c *Config) ApplyToTransport(transport *http.Transport) {
    transport.TLSClientConfig = c.ToTLSConfig()
}
```

### Step 3: Write Unit Tests (70 lines)
**File**: `pkg/tls/config_test.go`
```go
package tls

import (
    "testing"
    "net/http"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
    tests := []struct {
        name     string
        insecure bool
        expected bool
    }{
        {
            name:     "secure mode",
            insecure: false,
            expected: false,
        },
        {
            name:     "insecure mode",
            insecure: true,
            expected: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cfg := NewConfig(tt.insecure)
            assert.Equal(t, tt.expected, cfg.InsecureSkipVerify)
        })
    }
}

func TestToTLSConfig(t *testing.T) {
    cfg := NewConfig(true)
    tlsConfig := cfg.ToTLSConfig()

    require.NotNil(t, tlsConfig)
    assert.True(t, tlsConfig.InsecureSkipVerify)
}

func TestApplyToHTTPClient(t *testing.T) {
    client := &http.Client{}
    cfg := NewConfig(true)

    cfg.ApplyToHTTPClient(client)

    transport, ok := client.Transport.(*http.Transport)
    require.True(t, ok)
    require.NotNil(t, transport.TLSClientConfig)
    assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
}
```

## Size Management
- **Estimated Lines**: 180 total
  - cmd/push.go: ~30 lines (flag addition)
  - pkg/tls/config.go: ~80 lines (TLS config implementation)
  - pkg/tls/config_test.go: ~70 lines (unit tests)
- **Measurement Tool**: `${PROJECT_ROOT}/tools/line-counter.sh`
- **Check Frequency**: After each file completion
- **Split Threshold**: 700 lines (warning), 800 lines (stop)

## Test Requirements
- **Unit Tests**: 90% coverage for pkg/tls/
- **Integration Tests**: Verify flag parsing in command tests
- **Security Tests**: Validate certificate handling in both modes
- **Test Files**:
  - `pkg/tls/config_test.go`: Unit tests for TLS configuration
  - `cmd/push_test.go`: Integration test for flag (separate effort)

## Pattern Compliance
- **Go Patterns**:
  - Factory pattern for configuration creation
  - Option pattern for extensibility
  - Interface compliance with standard library
- **Security Requirements**:
  - Default to secure (verify certificates)
  - Explicit opt-in for insecure mode
  - Clear warning messages when insecure mode is used
- **CLI Patterns**:
  - Standard cobra flag definition
  - Consistent with other IDPBuilder commands

## Integration Points
- **With Effort 1.1.1**: Push command will use this TLS config
- **With Effort 1.1.2**: Authentication will work over configured TLS
- **With Wave 2.1**: Registry client will apply this TLS configuration
- **With go-containerregistry**: TLS config will be passed to registry transport

## Security Considerations
1. **Default Behavior**: Always verify certificates by default
2. **Warning Messages**: Display clear warning when --insecure is used
3. **Documentation**: Clear documentation on when to use --insecure
4. **Audit Logging**: Log when insecure mode is enabled

## Success Criteria
- ✅ --insecure flag properly added to push command
- ✅ TLS configuration factory creates proper tls.Config
- ✅ Configuration can be applied to HTTP clients/transports
- ✅ Unit tests achieve >90% coverage
- ✅ Default behavior is secure (certificates verified)
- ✅ Clear documentation and warning messages

## Out of Scope
- Advanced certificate management (cert pinning, custom CA bundles)
- Certificate rotation handling
- mTLS (mutual TLS) support
- Certificate generation or signing

## Dependencies for Future Efforts
This effort provides the TLS configuration foundation that will be used by:
- Registry client connection handling (Wave 2.1)
- Authentication over TLS (Wave 2.2)
- OCI image push operations (Phase 4)

## Notes
- Keep implementation simple and focused
- Use standard library crypto/tls where possible
- Ensure compatibility with go-containerregistry library
- Follow Go best practices for configuration handling