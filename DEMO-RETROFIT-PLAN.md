# Demo Retrofit Plan - Image Builder

## Features Discovered

Based on analysis of the implemented code in pkg/build/:

1. **OCI Image Building** (`image_builder.go`)
   - Create OCI images from directory contexts
   - Build and store container images locally
   - Feature flag support for image builder operations

2. **TLS Certificate Management** (`tls.go`)
   - Generate self-signed certificates
   - Create Kubernetes TLS secrets
   - Support for ArgoCD TLS configuration

3. **Build Context Management** (`context.go`)
   - Handle build context preparation
   - Support for various context types

4. **Storage Backend** (`storage.go`)
   - Local storage for built images
   - Image registry operations

5. **Feature Flags** (`feature_flags.go`)
   - Dynamic enable/disable of builder features
   - Environment-based configuration

## Demo Scenarios

### Scenario 1: Build Simple OCI Image
**Commands:**
```bash
./demo-features.sh build-image \
  --context ./test-data/sample-app \
  --tag myapp:v1.0 \
  --storage /tmp/oci-storage
```
**Expected output:** 
- Image built successfully
- SHA256 digest displayed
- Image stored in local storage

### Scenario 2: Generate TLS Certificates
**Commands:**
```bash
./demo-features.sh generate-certs \
  --namespace demo \
  --secret-name demo-tls \
  --output ./test-data/certs
```
**Expected output:**
- Certificate and key generated
- Files written to test-data/certs/
- Ready for Kubernetes secret creation

### Scenario 3: Push to Registry with TLS
**Commands:**
```bash
./demo-features.sh push-with-tls \
  --image myapp:v1.0 \
  --registry localhost:5000 \
  --cert-path ./test-data/certs/ca.crt
```
**Expected output:**
- Image pushed to registry
- TLS verification successful
- Push confirmation with digest

### Scenario 4: Feature Flag Toggle
**Commands:**
```bash
# Enable feature
export IMAGE_BUILDER_ENABLED=true
./demo-features.sh status

# Disable feature
export IMAGE_BUILDER_ENABLED=false
./demo-features.sh status
```
**Expected output:**
- Feature status displayed
- Operations blocked when disabled

## Size Impact

- Current implementation: ~1,200 lines (Phase 1 complete)
- Demo additions: ~150 lines
  - demo-features.sh: ~100 lines
  - DEMO.md: ~30 lines
  - test-data setup: ~20 lines
- Total after demo: ~1,350 lines (well within 800 limit for any single PR)

## Integration Hooks

### Wave-Level Demo Integration
- Integrate with gitea-client demos for end-to-end registry operations
- Share TLS certificates between efforts for consistent security demo

### Phase-Level Demo Integration
- Provide base image building for Phase 2 complete demo
- Export functions for use in integration test suite

## Demo Deliverables

1. **demo-features.sh** - Executable demo script with the following functions:
   - `build-image`: Build OCI image from context
   - `generate-certs`: Create TLS certificates
   - `push-with-tls`: Push image with TLS verification
   - `status`: Show feature flag status

2. **DEMO.md** - Documentation explaining:
   - How to run each demo scenario
   - Expected outputs and validation steps
   - Integration with other Phase 2 efforts

3. **test-data/** - Sample files including:
   - `sample-app/`: Simple app context for building
   - `certs/`: Generated certificates (gitignored)
   - `configs/`: Sample configuration files

## Implementation Notes

The demo will showcase the complete certificate extraction and trust management infrastructure that was implemented in Phase 1. Focus areas:

1. **Security**: Demonstrate proper TLS certificate handling
2. **Flexibility**: Show feature flag configuration
3. **Integration**: Connect with Gitea registry operations
4. **Error Handling**: Display graceful failure modes

## Success Metrics

- All 4 demo scenarios execute without errors
- Demo can run in isolation or as part of wave integration
- Total added code stays under 150 lines
- Clear documentation for operators