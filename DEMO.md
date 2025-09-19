# Gitea Client Split-002 Demo Guide

This demo showcases the advanced registry operations implemented in Split-002, including push operations with progress tracking, repository listing with pagination, retry logic with exponential backoff, and repository deletion capabilities.

## Overview

Split-002 completes the Gitea registry client with operational features:

1. **Performance**: Chunked uploads for large images
2. **Reliability**: Retry logic for network issues  
3. **Observability**: Progress tracking and logging
4. **Testing**: Comprehensive stubs for unit tests

## Prerequisites

- Access to a Gitea registry (or use simulation mode)
- Split-001 authentication components available
- `demo-features.sh` script in the current directory

## Demo Scenarios

### 1. Push Image with Progress Tracking

Demonstrates multi-layer image push with real-time progress reporting.

```bash
./demo-features.sh push \
  --registry https://gitea.local:3000 \
  --image myapp:v1.0 \
  --source ./test-data/image.tar \
  --progress
```

**Features Demonstrated:**
- Chunked upload with configurable size (5MB default)
- Progress callbacks with layer and byte tracking
- SHA256 digest verification per layer
- Upload speed calculation
- Final manifest push

**Expected Output:**
- Layer upload progress bars
- Bytes transferred counter and upload speed
- SHA256 digest confirmations
- Total upload time and size metrics

### 2. Repository Listing with Pagination

Shows repository discovery with pagination support and metadata retrieval.

```bash
./demo-features.sh list-repos \
  --registry https://gitea.local:3000 \
  --page 1 \
  --per-page 10 \
  --format table
```

**Features Demonstrated:**
- Registry catalog API integration
- Pagination controls (page/per-page)
- Tag information retrieval
- Formatted table output
- Repository metadata display

**Expected Output:**
- Formatted table of repositories with tags and timestamps
- Pagination information (page X of Y)
- Total repository count
- Repository metadata (size, last push time)

### 3. Retry Logic with Exponential Backoff

Illustrates network resilience through configurable retry mechanisms.

```bash
./demo-features.sh push-with-retry \
  --registry https://gitea.local:3000 \
  --image stress-test:v1.0 \
  --simulate-failures 3 \
  --max-retries 5
```

**Features Demonstrated:**
- Exponential backoff algorithm
- Configurable retry attempts and delays
- Error classification (retryable vs permanent)
- Request timeout handling
- Retry attempt logging

**Expected Output:**
- Individual retry attempts with failure reasons
- Backoff delay calculations (1s, 2s, 4s, etc.)
- Final success after configured retries
- Total time including retry delays

### 4. Repository Deletion

Shows repository cleanup operations with confirmation safeguards.

```bash
./demo-features.sh delete \
  --registry https://gitea.local:3000 \
  --repo myapp \
  --confirm
```

**Features Demonstrated:**
- Repository existence verification
- Confirmation requirement for safety
- Manifest and blob cleanup
- Catalog update procedures
- Verification of successful deletion

**Expected Output:**
- Confirmation prompt handling
- Step-by-step deletion progress
- Cleanup verification results
- Success confirmation message

## Integration with Split-001

Split-002 seamlessly integrates with Split-001 authentication:

- **Shared Authentication**: Uses auth manager from Split-001
- **Common Registry Client**: Extends base client with operational features
- **TLS Configuration**: Leverages Split-001 certificate handling
- **Remote Options**: Compatible with Split-001 transport settings

### Integration Example

```bash
# Authenticate using Split-001 components
export GITEA_USERNAME="demo-user"
export GITEA_PASSWORD="demo-pass"

# Use Split-002 operations with Split-001 auth
./demo-features.sh push --registry https://gitea.local:3000 --image demo:latest
```

## Performance Tuning

### Chunk Size Optimization

Configure upload chunk size based on network conditions:

```bash
# For fast networks
export CHUNK_SIZE="10MB"

# For slower networks
export CHUNK_SIZE="1MB"

# For very slow networks
export CHUNK_SIZE="512KB"
```

### Retry Policy Tuning

Adjust retry behavior for different environments:

```bash
# Aggressive retries for unstable networks
export MAX_RETRIES=10
export INITIAL_DELAY="500ms"
export BACKOFF_FACTOR=1.5

# Conservative retries for stable networks
export MAX_RETRIES=3
export INITIAL_DELAY="1s"
export BACKOFF_FACTOR=2.0
```

## Troubleshooting

### Common Issues

#### 1. Authentication Failures

```bash
# Verify Split-001 auth setup
echo "Checking authentication..."
curl -H "Authorization: Bearer $GITEA_TOKEN" https://gitea.local:3000/v2/

# Solution: Ensure Split-001 auth components are configured
export GITEA_TOKEN=$(get-auth-token.sh)
```

#### 2. Network Timeouts

```bash
# Increase timeout values
export REQUEST_TIMEOUT="60s"
export TLS_HANDSHAKE_TIMEOUT="30s"

# Enable retry logic
./demo-features.sh push-with-retry --max-retries 10
```

#### 3. Large Image Upload Failures

```bash
# Use smaller chunk sizes
export CHUNK_SIZE="1MB"

# Enable progress tracking
./demo-features.sh push --progress
```

#### 4. TLS Certificate Issues

```bash
# Use insecure mode for testing (NOT for production)
export TLS_INSECURE=true

# Or add certificate to trust store (recommended)
cp gitea-cert.pem /usr/local/share/ca-certificates/
update-ca-certificates
```

### Debug Mode

Enable detailed logging for troubleshooting:

```bash
export DEBUG_MODE=true
export LOG_LEVEL=debug

./demo-features.sh push --registry https://gitea.local:3000 --image debug:latest
```

## Testing Configuration

### Mock Registry Setup

For testing without a real Gitea instance:

```bash
# Use test stubs from stubs.go
export DEMO_MODE="simulation"
export MOCK_REGISTRY=true

./demo-features.sh push --registry mock://localhost
```

### Performance Testing

Benchmark upload performance:

```bash
# Create test image
dd if=/dev/zero of=./test-data/large-image.tar bs=1M count=100

# Test with different chunk sizes
for chunk in 1MB 5MB 10MB; do
    export CHUNK_SIZE=$chunk
    time ./demo-features.sh push --source ./test-data/large-image.tar
done
```

## Security Considerations

### Authentication Security

- Always use HTTPS in production
- Store credentials securely (environment variables, not files)
- Rotate authentication tokens regularly
- Use least-privilege access controls

### TLS Security

- Verify server certificates in production
- Use `--insecure` only for development/testing
- Keep CA certificates updated
- Monitor for certificate expiration

### Network Security

- Use private networks when possible
- Implement network policies for registry access
- Monitor for suspicious upload patterns
- Log all registry operations for audit

## Advanced Usage

### Batch Operations

Process multiple images:

```bash
#!/bin/bash
for image in $(cat image-list.txt); do
    ./demo-features.sh push --image "$image" --progress
done
```

### Integration with CI/CD

```yaml
# Example GitHub Actions step
- name: Push to Gitea Registry
  run: |
    ./demo-features.sh push \
      --registry ${{ secrets.GITEA_REGISTRY_URL }} \
      --image ${{ github.repository }}:${{ github.sha }} \
      --progress
```

### Monitoring Integration

```bash
# Export metrics for monitoring
./demo-features.sh push --image monitor:latest --progress | \
  grep -E "(uploaded|complete)" | \
  curl -X POST http://metrics-collector/api/registry-ops
```

## API Reference

The demo script supports these environment variables:

- `REGISTRY_URL`: Default registry URL
- `DEMO_MODE`: Set to "simulation" for testing
- `CHUNK_SIZE`: Upload chunk size
- `MAX_RETRIES`: Maximum retry attempts
- `INITIAL_DELAY`: Initial retry delay
- `BACKOFF_FACTOR`: Exponential backoff multiplier
- `DEBUG_MODE`: Enable debug logging
- `TLS_INSECURE`: Skip TLS verification (testing only)

## Support

For issues or questions:

1. Check the troubleshooting section above
2. Review logs with debug mode enabled
3. Verify Split-001 integration is working
4. Test with simulation mode to isolate issues

## Version Compatibility

- Requires Split-001 (authentication) to be implemented
- Compatible with go-containerregistry v0.19.0+
- Tested with Gitea 1.20+
- Supports Docker Registry API v2