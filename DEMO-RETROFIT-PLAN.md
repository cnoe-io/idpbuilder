# Demo Retrofit Plan - Gitea Client Split-001 (Core)

## Features Discovered

Based on analysis of the implemented code in pkg/registry/:

1. **Registry Interface** (`interface.go`)
   - Core registry operations contract
   - Push, List, Exists, Delete operations
   - Configuration management

2. **Authentication System** (`auth.go`)
   - Token-based authentication for Gitea
   - Bearer token generation
   - Credential management
   - Token refresh capabilities

3. **Gitea Registry Client** (`gitea.go`)
   - HTTP client with TLS configuration
   - Proxy support
   - Registry URL handling
   - Connection management

4. **Remote Options** (`remote_options.go`)
   - Configurable retry logic
   - Timeout management
   - TLS/Insecure mode settings
   - Proxy configuration

## Demo Scenarios

### Scenario 1: Basic Authentication
**Commands:**
```bash
./demo-features.sh auth \
  --registry https://gitea.local:3000 \
  --username demo-user \
  --token ${GITEA_TOKEN}
```
**Expected output:**
- Authentication successful
- Bearer token generated
- Connection established

### Scenario 2: List Repositories
**Commands:**
```bash
./demo-features.sh list \
  --registry https://gitea.local:3000 \
  --format json
```
**Expected output:**
- JSON array of repository names
- Total count displayed
- Response time shown

### Scenario 3: Check Repository Existence
**Commands:**
```bash
./demo-features.sh exists \
  --registry https://gitea.local:3000 \
  --repo myapp/v1.0
```
**Expected output:**
- Repository exists: true/false
- Metadata if exists (size, last modified)

### Scenario 4: TLS Configuration Demo
**Commands:**
```bash
# With custom CA
./demo-features.sh test-tls \
  --registry https://gitea.local:3000 \
  --ca-cert ./test-data/ca.crt

# Insecure mode (testing only)
./demo-features.sh test-tls \
  --registry https://gitea.local:3000 \
  --insecure
```
**Expected output:**
- TLS verification status
- Certificate details displayed
- Security warnings for insecure mode

## Size Impact

- Current implementation: ~700 lines (Split-001 complete)
- Demo additions: ~120 lines
  - demo-features.sh: ~80 lines
  - DEMO.md: ~25 lines
  - test-data setup: ~15 lines
- Total after demo: ~820 lines (slightly over limit, but demo is separate)

## Integration Hooks

### Split-Level Integration
- Coordinates with Split-002 for push/delete operations
- Shares authentication manager between splits
- Common configuration objects

### Wave-Level Demo Integration
- Integrates with image-builder TLS certificates
- Provides registry client for end-to-end demos
- Shares test data with other efforts

### Phase-Level Demo Integration
- Foundation for Phase 2 registry operations
- Authentication reused across all registry interactions
- Base client for advanced operations

## Demo Deliverables

1. **demo-features.sh** - Executable demo script with:
   - `auth`: Test authentication flow
   - `list`: List repositories
   - `exists`: Check repository existence
   - `test-tls`: Verify TLS configuration

2. **DEMO.md** - Documentation including:
   - Setup instructions for local Gitea
   - Authentication token generation guide
   - TLS certificate configuration
   - Integration with Split-002 features

3. **test-data/** - Sample files:
   - `ca.crt`: Sample CA certificate
   - `config.yaml`: Registry configuration
   - `.env.example`: Environment variables template

## Implementation Notes

This split provides the core foundation for Gitea registry operations. The demo focuses on:

1. **Authentication**: Demonstrating secure token-based auth
2. **Discovery**: Repository listing and existence checking
3. **Security**: TLS configuration and certificate handling
4. **Configuration**: Flexible options for various environments

## Integration with Split-002

Split-002 builds on this foundation by adding:
- Push operations (using auth from Split-001)
- Delete operations (using client from Split-001)
- Retry logic demonstration
- Advanced error handling

## Success Metrics

- All 4 demo scenarios execute successfully
- Authentication tokens properly managed
- TLS verification works with custom CA
- Clear separation between Split-001 and Split-002 demos
- Total demo code under 150 lines