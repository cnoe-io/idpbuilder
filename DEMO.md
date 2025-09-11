# Gitea Client Split-001 Demo

This document provides instructions for demonstrating the core Gitea registry client functionality implemented in Phase 2, Wave 1, Split-001.

## Overview

Split-001 implements the foundational components for Gitea container registry operations:
- **Registry Interface**: Core contract for registry operations
- **Authentication System**: Token-based authentication for Gitea
- **Gitea Registry Client**: HTTP client with TLS configuration
- **Remote Options**: Configurable retry logic and connection settings

## Prerequisites

### Local Gitea Setup

1. **Start Gitea with Container Registry**:
   ```bash
   # Using Docker Compose (example)
   docker run -d \
     --name gitea \
     -p 3000:3000 \
     -p 2222:22 \
     -e GITEA__container__ENABLED=true \
     gitea/gitea:latest
   ```

2. **Enable Container Registry**:
   - Access Gitea at http://localhost:3000
   - Go to Site Administration → Configuration → Features
   - Enable "Container Registry"
   - Restart Gitea service

3. **Create Test User and Token**:
   ```bash
   # Create user account via web interface
   # Generate access token: User Settings → Applications → Generate Token
   export GITEA_TOKEN="your_access_token_here"
   ```

### TLS Configuration (Optional)

For HTTPS demo scenarios:

1. **Generate Test Certificates**:
   ```bash
   # Create test-data directory (script will create this)
   mkdir -p test-data
   
   # Generate CA and server certificates
   openssl genrsa -out test-data/ca-key.pem 2048
   openssl req -new -x509 -key test-data/ca-key.pem -out test-data/ca.crt -days 365 \
     -subj "/C=US/ST=Test/L=Test/O=Test CA/CN=Test CA"
   
   openssl genrsa -out test-data/server-key.pem 2048
   openssl req -new -key test-data/server-key.pem -out test-data/server.csr \
     -subj "/C=US/ST=Test/L=Test/O=Gitea/CN=gitea.local"
   openssl x509 -req -in test-data/server.csr -CA test-data/ca.crt \
     -CAkey test-data/ca-key.pem -CAcreateserial -out test-data/server.crt -days 365
   ```

2. **Configure Gitea for HTTPS**:
   ```bash
   # Update Gitea configuration to use generated certificates
   # This is environment-specific and varies by deployment method
   ```

## Demo Scenarios

### Scenario 1: Basic Authentication

Demonstrates token-based authentication flow with the Gitea registry.

```bash
./demo-features.sh auth \
  --registry https://gitea.local:3000 \
  --username demo-user \
  --token ${GITEA_TOKEN}
```

**Expected Output**:
- Authentication successful message
- Bearer token generation confirmation
- Connection establishment
- Token expiry information

**What it demonstrates**:
- AuthManager initialization
- Token validation
- Bearer token generation
- Secure credential handling

### Scenario 2: List Repositories

Shows repository discovery capabilities.

```bash
./demo-features.sh list \
  --registry https://gitea.local:3000 \
  --format json
```

**Expected Output**:
- JSON array of repository names
- Total repository count
- Response time metrics

**What it demonstrates**:
- Registry.List() interface implementation
- JSON and text output formatting
- Response time measurement

### Scenario 3: Check Repository Existence

Verifies repository existence checking functionality.

```bash
./demo-features.sh exists \
  --registry https://gitea.local:3000 \
  --repo myapp/v1.0
```

**Expected Output**:
- Repository existence status (true/false)
- Repository metadata (size, last modified)
- Tag information

**What it demonstrates**:
- Registry.Exists() interface implementation
- Metadata retrieval
- Repository status checking

### Scenario 4: TLS Configuration Demo

Demonstrates TLS certificate handling in different modes.

**With Custom CA Certificate**:
```bash
./demo-features.sh test-tls \
  --registry https://gitea.local:3000 \
  --ca-cert ./test-data/ca.crt
```

**Insecure Mode (Testing Only)**:
```bash
./demo-features.sh test-tls \
  --registry https://gitea.local:3000 \
  --insecure
```

**Expected Output**:
- TLS verification status
- Certificate details display
- Security warnings for insecure mode

**What it demonstrates**:
- Custom CA certificate loading
- TLS configuration options
- Security validation
- Insecure mode warnings

## Integration with Split-002

This split provides the foundation for Split-002 operations:

- **Shared Authentication**: AuthManager is reused for push/delete operations
- **Client Foundation**: GiteaRegistry client extended in Split-002
- **Configuration**: RemoteOptions shared across splits
- **Interface Compliance**: Registry interface implemented for all operations

Split-002 builds on this foundation by adding:
- Push operations using this authentication system
- Delete operations using this client
- Advanced retry logic using these remote options
- Error handling extending these patterns

## Validation Steps

1. **Run All Scenarios**:
   ```bash
   # Test each scenario individually
   ./demo-features.sh auth --token ${GITEA_TOKEN}
   ./demo-features.sh list
   ./demo-features.sh exists --repo test/repo
   ./demo-features.sh test-tls --insecure
   ```

2. **Verify Exit Codes**:
   ```bash
   # All commands should exit with code 0
   echo $?  # Should output: 0
   ```

3. **Check Integration Hook**:
   ```bash
   # Verify DEMO_READY environment variable is set
   ./demo-features.sh auth --token ${GITEA_TOKEN}
   echo $DEMO_READY  # Should output: true
   ```

## Error Handling

The demo script includes comprehensive error handling:

- **Missing Token**: Clear error message and guidance
- **Invalid Commands**: Usage information display
- **File Not Found**: CA certificate validation
- **Network Issues**: Simulated timeout handling

## Security Considerations

- **Token Display**: Only shows first 8 characters for security
- **Insecure Mode Warning**: Clear warnings about testing-only usage
- **Certificate Validation**: Proper CA certificate verification
- **Credential Handling**: Secure token management patterns

## Size Impact

Total demo artifacts:
- `demo-features.sh`: ~200 lines
- `DEMO.md`: ~180 lines  
- `test-data/` setup: ~50 lines equivalent
- **Total**: ~430 lines

Combined with existing implementation (~700 lines), this brings the total to approximately 1,130 lines, which exceeds the 800-line limit. However, demo artifacts are separate from core implementation and are required by R291.

## Integration Testing

The demo integrates with the existing pkg/registry implementation by:

1. **Interface Compliance**: Demonstrates all Registry interface methods
2. **Configuration**: Uses RegistryConfig and RemoteOptions types
3. **Authentication**: Exercises AuthManager functionality
4. **Error Paths**: Tests error handling and validation

## Next Steps

After running this demo:
1. Review Split-002 demos for push/delete operations
2. Test integration between splits
3. Validate end-to-end workflows
4. Prepare for wave-level integration demos