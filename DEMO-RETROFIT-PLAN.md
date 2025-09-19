# Demo Retrofit Plan - Gitea Client Split-002 (Operations)

## Features Discovered

Based on analysis of the implemented code in pkg/registry/:

1. **Push Operations** (`push.go`)
   - Multi-layer image push support
   - Chunked upload with configurable size
   - Progress reporting callbacks
   - SHA256 digest verification
   - Manifest handling

2. **List Operations** (`list.go`)
   - List repositories and tags
   - Pagination support
   - Filtering capabilities
   - Metadata retrieval

3. **Retry Logic** (`retry.go`)
   - Exponential backoff implementation
   - Configurable retry attempts
   - Error classification (retryable vs permanent)
   - Request timeout handling

4. **Test Stubs** (`stubs.go`)
   - Mock registry for testing
   - Stubbed operations for unit tests
   - Error simulation capabilities
   - Response mocking

## Demo Scenarios

### Scenario 1: Push Image with Progress
**Commands:**
```bash
./demo-features.sh push \
  --registry https://gitea.local:3000 \
  --image myapp:v1.0 \
  --source ./test-data/image.tar \
  --progress
```
**Expected output:**
- Layer upload progress bars
- Bytes transferred counter
- Upload speed calculation
- Final digest confirmation

### Scenario 2: List Repositories with Pagination
**Commands:**
```bash
./demo-features.sh list-repos \
  --registry https://gitea.local:3000 \
  --page 1 \
  --per-page 10 \
  --format table
```
**Expected output:**
- Formatted table of repositories
- Pagination info (page X of Y)
- Repository metadata (size, last push)

### Scenario 3: Retry Logic Demonstration
**Commands:**
```bash
# Simulate network issues
./demo-features.sh push-with-retry \
  --registry https://gitea.local:3000 \
  --image stress-test:v1.0 \
  --simulate-failures 3 \
  --max-retries 5
```
**Expected output:**
- Retry attempts logged
- Backoff delays shown
- Final success after retries
- Total time with retries

### Scenario 4: Delete Repository
**Commands:**
```bash
./demo-features.sh delete \
  --registry https://gitea.local:3000 \
  --repo myapp \
  --confirm
```
**Expected output:**
- Confirmation prompt
- Deletion progress
- Success confirmation
- Cleanup verification

## Size Impact

- Current implementation: ~750 lines (Split-002 complete)
- Demo additions: ~130 lines
  - demo-features.sh: ~90 lines
  - DEMO.md: ~25 lines
  - test-data setup: ~15 lines
- Total after demo: ~880 lines (slightly over, but demo is separate)

## Integration Hooks

### Split-Level Integration
- Uses authentication from Split-001
- Shares registry client instance
- Common error handling patterns

### Wave-Level Demo Integration
- Combines with image-builder for end-to-end push
- Uses Split-001 auth for all operations
- Integrates with TLS certificates from image-builder

### Phase-Level Demo Integration
- Complete registry operations suite
- Foundation for Phase 3 advanced features
- Performance benchmarking capabilities

## Demo Deliverables

1. **demo-features.sh** - Executable demo script with:
   - `push`: Push image with progress tracking
   - `list-repos`: List with pagination
   - `push-with-retry`: Demonstrate retry logic
   - `delete`: Remove repository

2. **DEMO.md** - Documentation including:
   - Complete setup instructions
   - Integration with Split-001 features
   - Performance tuning guide
   - Troubleshooting common issues

3. **test-data/** - Sample files:
   - `image.tar`: Sample OCI image
   - `stress-test/`: Large image for testing
   - `config-retry.yaml`: Retry configuration

## Implementation Notes

This split completes the Gitea registry client with operational features:

1. **Performance**: Chunked uploads for large images
2. **Reliability**: Retry logic for network issues
3. **Observability**: Progress tracking and logging
4. **Testing**: Comprehensive stubs for unit tests

## Dependencies on Split-001

This split requires Split-001 for:
- Authentication manager
- Base registry client
- TLS configuration
- Remote options

## Success Metrics

- All 4 demo scenarios execute successfully
- Push operations handle large images (>100MB)
- Retry logic recovers from transient failures
- Progress tracking provides real-time feedback
- Integration with Split-001 seamless

## Performance Considerations

- Chunk size optimization for different networks
- Concurrent layer uploads (future enhancement)
- Connection pooling for multiple operations
- Memory-efficient streaming for large files