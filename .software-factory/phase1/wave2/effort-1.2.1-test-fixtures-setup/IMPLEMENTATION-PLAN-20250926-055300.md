# effort-1.2.1-test-fixtures-setup Implementation Plan

## Effort Metadata (R211 Compliance)
**Branch**: `igp/phase1/wave2/effort-1.2.1-test-fixtures-setup`
**Can Parallelize**: No (First effort in Wave 2, sequential implementation required)
**Parallel With**: None
**Size Estimate**: ~250 lines
**Dependencies**: Wave 1 integration (phase1-wave1-integration branch)

## Overview
- **Effort**: Test fixtures and helper functions setup
- **Phase**: 1, Wave: 2
- **Theme**: Establish comprehensive test infrastructure for the idpbuilder push command testing
- **Estimated Size**: 250 lines
- **Implementation Time**: 2-3 hours

## Context from Wave 1 Dependencies (R219)
### Analyzed Dependencies
1. **Wave 1 Integration (phase1-wave1-integration)**:
   - Push command skeleton established
   - Authentication flags implemented (--username, --password)
   - TLS configuration (--insecure-skip-verify, --cacert)
   - Total: 537 lines across 3 efforts

### Integration Strategy
This effort will:
- Build test helpers that can test Wave 1 components
- Create fixtures compatible with Cobra command testing
- Establish patterns for authentication and TLS testing
- Prepare foundation for effort-1.2.2 (command-testing-framework)

## File Structure
```
test/
├── fixtures/                     # Test data directory
│   ├── auth/                     # Authentication test fixtures
│   │   ├── credentials.yaml      # Sample credential configs (~20 lines)
│   │   └── tokens.json           # Sample token data (~15 lines)
│   ├── certs/                    # Certificate fixtures for TLS testing
│   │   ├── ca.crt                # Sample CA certificate (data file)
│   │   ├── client.crt            # Sample client certificate (data file)
│   │   └── client.key            # Sample client key (data file)
│   ├── images/                   # OCI image fixtures
│   │   ├── manifest.json         # Sample OCI manifest (~30 lines)
│   │   ├── config.json           # Sample OCI config (~25 lines)
│   │   └── layer.tar             # Small test layer (binary file)
│   └── repos/                    # Repository configurations
│       ├── gitea-config.yaml    # Gitea registry config (~20 lines)
│       └── registry-urls.txt    # Test registry URLs (~10 lines)
└── helpers.go                    # Helper functions (~130 lines)
```

## Implementation Steps

### Step 1: Create Test Directory Structure
```bash
mkdir -p test/fixtures/auth
mkdir -p test/fixtures/certs
mkdir -p test/fixtures/images
mkdir -p test/fixtures/repos
```

### Step 2: Implement Test Helper Functions (test/helpers.go)
```go
package test

import (
    "testing"
    "os"
    "path/filepath"
    "io/ioutil"
    "encoding/json"
    "github.com/stretchr/testify/require"
)

// Core helper functions to implement:

// GetFixturePath returns absolute path to fixture file
func GetFixturePath(t *testing.T, relativePath string) string

// LoadFixture loads fixture content as bytes
func LoadFixture(t *testing.T, relativePath string) []byte

// LoadJSONFixture loads and unmarshals JSON fixture
func LoadJSONFixture(t *testing.T, relativePath string, target interface{})

// CreateTempRegistry creates a temporary directory simulating a registry
func CreateTempRegistry(t *testing.T) (string, func())

// SetupTestCredentials sets up test authentication environment
func SetupTestCredentials(t *testing.T) (username, password string, cleanup func())

// SetupTestTLS creates test certificates and returns paths
func SetupTestTLS(t *testing.T) (caCert, clientCert, clientKey string, cleanup func())

// MockGiteaRegistry creates a mock Gitea registry configuration
func MockGiteaRegistry(t *testing.T) *GiteaRegistryConfig

// CreateTestImageFixture creates a minimal OCI image for testing
func CreateTestImageFixture(t *testing.T, name string) string

// CompareManifests compares two OCI manifests for testing
func CompareManifests(t *testing.T, expected, actual []byte)

// SetupTestEnvironment prepares complete test environment
func SetupTestEnvironment(t *testing.T) *TestEnvironment

// TestEnvironment struct containing all test resources
type TestEnvironment struct {
    TempDir string
    Registry string
    Credentials *TestCredentials
    TLS *TestTLSConfig
    Cleanup func()
}

// TestCredentials for authentication testing
type TestCredentials struct {
    Username string
    Password string
    Token string
}

// TestTLSConfig for TLS testing
type TestTLSConfig struct {
    CACert string
    ClientCert string
    ClientKey string
    InsecureSkipVerify bool
}

// GiteaRegistryConfig for Gitea-specific testing
type GiteaRegistryConfig struct {
    URL string
    Namespace string
    Repository string
}
```

### Step 3: Create Authentication Fixtures (test/fixtures/auth/)
**credentials.yaml** (~20 lines):
```yaml
users:
  - username: testuser
    password: testpass123
    email: test@example.com
  - username: admin
    password: adminpass456
    email: admin@example.com
tokens:
  bearer: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  basic: "dGVzdHVzZXI6dGVzdHBhc3MxMjM="
```

**tokens.json** (~15 lines):
```json
{
  "tokens": [
    {
      "type": "bearer",
      "value": "test-bearer-token",
      "expires": "2025-12-31T23:59:59Z"
    }
  ]
}
```

### Step 4: Create Certificate Fixtures (test/fixtures/certs/)
- Generate self-signed certificates for testing
- Include CA certificate, client certificate, and client key
- Add README.md explaining certificate generation commands

### Step 5: Create OCI Image Fixtures (test/fixtures/images/)
**manifest.json** (~30 lines):
```json
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.oci.image.config.v1+json",
    "digest": "sha256:test",
    "size": 1234
  },
  "layers": [
    {
      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
      "digest": "sha256:layer1",
      "size": 5678
    }
  ]
}
```

**config.json** (~25 lines):
```json
{
  "architecture": "amd64",
  "os": "linux",
  "config": {
    "Env": ["PATH=/usr/bin"],
    "Cmd": ["/bin/sh"],
    "WorkingDir": "/"
  },
  "rootfs": {
    "type": "layers",
    "diff_ids": ["sha256:test"]
  }
}
```

### Step 6: Create Repository Configuration Fixtures (test/fixtures/repos/)
**gitea-config.yaml** (~20 lines):
```yaml
registry:
  url: "localhost:3000"
  namespace: "test-org"
  repository: "test-app"
auth:
  username: "${GITEA_USERNAME}"
  password: "${GITEA_PASSWORD}"
tls:
  insecure: false
  ca_cert: "/path/to/ca.crt"
```

**registry-urls.txt** (~10 lines):
```
localhost:3000/test-org/test-app:latest
localhost:3000/test-org/test-app:v1.0.0
gitea.example.com/org/repo:main
```

## Size Management
- **Estimated Lines**: 250 (130 in helpers.go + 120 in fixtures)
- **Measurement Tool**: ${PROJECT_ROOT}/tools/line-counter.sh
- **Check Frequency**: After completing each major section
- **Split Threshold**: 700 lines (warning), 800 lines (stop)

## Test Requirements
- **Unit Tests**: Not applicable (this IS test infrastructure)
- **Self-Testing**: Helper functions should have minimal self-tests
- **Documentation**: Each helper function must have godoc comments
- **Examples**: Include example usage in comments

## Pattern Compliance
- **Go Testing Patterns**: Follow Go testing best practices
- **testify Integration**: Use testify/require for assertions
- **Fixture Organization**: Logical grouping by functionality
- **Helper Design**: Reusable, composable test helpers
- **Cleanup Functions**: Always return cleanup functions for resources

## Security Considerations
- Test credentials must be clearly marked as test-only
- No real credentials in fixtures
- Self-signed certificates only for testing
- Clear warnings in fixture files about test-only usage

## Dependencies
- **stretchr/testify**: Already in go.mod (v1.9.0)
- **Standard library**: os, path/filepath, io/ioutil, encoding/json
- **No new external dependencies required**

## Library Version Requirements (R381)
### Locked Dependencies (DO NOT UPDATE)
- stretchr/testify: v1.9.0 (LOCKED - existing in project)

### New Dependencies Allowed
- None required for this effort

## Integration Points
- **Wave 1**: Compatible with push command skeleton, auth flags, TLS config
- **Wave 2 Next**: effort-1.2.2 will use these fixtures and helpers
- **Future Waves**: All testing efforts will build on this foundation

## Success Criteria
1. All helper functions implemented and documented
2. Complete fixture directory structure created
3. Fixtures cover authentication, TLS, and OCI image scenarios
4. Helper functions are reusable and well-abstracted
5. Clear separation between test infrastructure and actual tests
6. No hardcoded paths - all paths resolved dynamically
7. Cleanup functions prevent test pollution

## Out of Scope
- Actual test implementations (that's effort-1.2.2)
- Integration with real registries
- Network mocking (future effort)
- Performance testing infrastructure
- E2E test framework

## Notes for SW Engineer
1. Start with helpers.go to establish the foundation
2. Create fixtures incrementally, testing each helper as you go
3. Use relative paths from test/ directory for all fixtures
4. Ensure all resources created during tests are cleaned up
5. Keep fixtures minimal but realistic
6. Add comments explaining the purpose of each fixture
7. This is test INFRASTRUCTURE only - no actual tests yet