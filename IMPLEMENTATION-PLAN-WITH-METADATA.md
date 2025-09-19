<!-- ⚠️ EFFORT INFRASTRUCTURE METADATA (R213 - ORCHESTRATOR DEFINED) ⚠️ -->
**METADATA_SOURCE**: ORCHESTRATOR (Single Source of Truth)
**METADATA_VERSION**: 1.0
**GENERATED_AT**: $(date -Iseconds)
**GENERATED_BY**: orchestrator

## 🔧 EFFORT INFRASTRUCTURE METADATA
**WORKING_DIRECTORY**: efforts/phase2/wave1/image-builder
**BRANCH**: idpbuilder-oci-build-push/phase2/wave1/image-builder
**EFFORT_NAME**: E2.1.1-image-builder
**EFFORT_NUMBER**: E2.1.1
**PHASE**: 2
**WAVE**: 1
<!-- END EFFORT METADATA -->

# Implementation Plan for go-containerregistry-image-builder (E2.1.1)

Created: 2025-09-08T00:18:19Z
Location: .software-factory/phase2/wave1/image-builder
Phase: 2
Wave: 1

## 🚨 CRITICAL EFFORT METADATA (FROM WAVE PLAN)

**Branch**: `idpbuilder-oci-build-push/phase2/wave1/image-builder`
**Can Parallelize**: Yes
**Parallel With**: [E2.1.2 gitea-client]
**Size Estimate**: 600 lines (MUST stay under 800)
**Dependencies**: None (can start immediately)
**Base Branch**: `idpbuilder-oci-build-push/phase1/integration`

## Overview

- **Effort**: E2.1.1 - go-containerregistry-image-builder
- **Phase**: 2 (Build & Push Implementation)
- **Wave**: 1 (Core Build & Push)
- **Estimated Size**: 600 lines
- **Implementation Time**: 6-8 hours
- **Purpose**: Implement OCI image assembly using go-containerregistry library

## Technical Architecture

### Core Components

1. **Builder Engine**: Central orchestrator for image building operations
2. **Build Context Handler**: Processes directories into tar archives
3. **Layer Manager**: Creates and manages OCI layers with compression
4. **Manifest Generator**: Produces OCI manifests with proper configuration
5. **Storage Manager**: Handles local OCI tarball storage

### Library Integration

This effort uses `github.com/google/go-containerregistry` v0.19.0 for:
- OCI image manipulation (`v1.Image` interface)
- Layer creation (`v1.Layer` from tar)
- Manifest generation (`v1.Manifest`)
- Local storage (`tarball.Write`)

### Data Flow

```
Build Context Directory → Tar Archive → Compressed Layer → OCI Image → Local Tarball
                      ↓                                  ↓
                .dockerignore                       Manifest + Config
```

## 🔴🔴🔴 EXPLICIT SCOPE CONTROL (R311 MANDATORY) 🔴🔴🔴

### IMPLEMENT EXACTLY:

1. **Type: Builder struct** (~30 lines)
   - Fields: storageDir, images map
   - NO additional fields

2. **Function: NewBuilder(storageDir string) *Builder** (~15 lines)
   - Initialize builder with storage directory
   - Create storage directory if not exists

3. **Function: BuildImage(ctx, opts BuildOptions) (*BuildResult, error)** (~80 lines)
   - Create tar from context directory
   - Apply .dockerignore exclusions
   - Build single layer from tar
   - Generate OCI manifest
   - Save to local storage
   - Return build result

4. **Function: createTarFromContext(path string, exclusions []string) (io.ReadCloser, error)** (~60 lines)
   - Walk directory tree
   - Apply exclusion patterns
   - Create tar archive
   - Return tar reader

5. **Function: createLayer(tarReader io.Reader) (v1.Layer, error)** (~25 lines)
   - Create layer from tar using go-containerregistry
   - Apply gzip compression
   - Return OCI layer

6. **Function: saveImageLocally(img v1.Image, tag, storageDir string) (string, error)** (~40 lines)
   - Write image to OCI tarball format
   - Store in designated directory
   - Return storage path

7. **Type: BuildOptions struct** (~15 lines)
   - ContextPath, Tag, Exclusions, Labels fields

8. **Type: BuildResult struct** (~10 lines)
   - ImageID, Digest, Size, StoragePath fields

9. **Tests: builder_test.go** (~150 lines)
   - TestNewBuilder
   - TestBuildImageBasic
   - TestBuildImageWithExclusions
   - TestBuildImageInvalidContext
   - Mock filesystem for testing

10. **Tests: context_test.go** (~100 lines)
    - TestCreateTarFromContext
    - TestExclusionPatterns
    - TestEmptyDirectory

**TOTAL EXACTLY**: ~525 lines (well under 800 limit)

### ❌❌❌ DO NOT IMPLEMENT (CRITICAL - R311):

- ❌ ListImages function (future effort)
- ❌ RemoveImage function (future effort)
- ❌ TagImage function (future effort)
- ❌ Multi-stage builds (out of scope)
- ❌ Dockerfile parsing (not needed)
- ❌ Build cache management (future optimization)
- ❌ Registry push operations (E2.1.2's responsibility)
- ❌ Authentication handling (not our concern)
- ❌ Progress reporting callbacks (Wave 2 CLI concern)
- ❌ Concurrent build support (future enhancement)
- ❌ Image history tracking (not required)
- ❌ Layer deduplication (advanced feature)
- ❌ Config validation beyond basics (minimal MVP)
- ❌ Comprehensive error types (use standard errors)
- ❌ Metrics/telemetry (out of scope)

**FAILURE TO RESPECT THESE BOUNDARIES = PROJECT FAILURE**

## File Structure

```
pkg/build/
├── builder.go          # Main Builder implementation (~150 lines)
├── context.go          # Build context and tar handling (~100 lines)
├── layer.go            # Layer creation utilities (~50 lines)
├── manifest.go         # Manifest generation (~50 lines)
├── storage.go          # Local storage management (~75 lines)
├── types.go            # BuildOptions, BuildResult types (~25 lines)
├── builder_test.go     # Builder unit tests (~150 lines)
└── context_test.go     # Context handling tests (~100 lines)
```

**Note**: Some files from the original wave plan are being consolidated to stay within scope:
- `layer.go` and `manifest.go` functionality folded into `builder.go`
- `storage.go` simplified to just saving functionality
- No integration tests in this effort (can be added if room)

## Implementation Steps

### Step 1: Create Types and Interfaces (50 lines)
```go
// In pkg/build/types.go
type BuildOptions struct {
    ContextPath string
    Tag         string
    Exclusions  []string
    Labels      map[string]string
}

type BuildResult struct {
    ImageID     string
    Digest      v1.Hash
    Size        int64
    StoragePath string
}

// In pkg/build/builder.go
type Builder struct {
    storageDir string
    images     map[string]string // tag -> path mapping
}
```

### Step 2: Implement Context Processing (100 lines)
```go
// In pkg/build/context.go
func createTarFromContext(contextPath string, exclusions []string) (io.ReadCloser, error) {
    // Implementation details:
    // 1. Validate context directory exists
    // 2. Create pipe for tar writer
    // 3. Walk directory applying exclusions
    // 4. Write files to tar archive
    // 5. Return reader end of pipe
}
```

### Step 3: Implement Core Builder (200 lines)
```go
// In pkg/build/builder.go
func (b *Builder) BuildImage(ctx context.Context, opts BuildOptions) (*BuildResult, error) {
    // 1. Create tar from context
    // 2. Create layer from tar
    // 3. Build image with single layer
    // 4. Add labels to config
    // 5. Save to local storage
    // 6. Return build result
}
```

### Step 4: Add Storage Management (75 lines)
```go
// In pkg/build/storage.go
func saveImageLocally(img v1.Image, tag, storageDir string) (string, error) {
    // 1. Generate filename from tag
    // 2. Create storage directory
    // 3. Write as OCI tarball
    // 4. Return full path
}
```

### Step 5: Write Unit Tests (250 lines)
- Test builder initialization
- Test successful build with mock filesystem
- Test exclusion patterns work correctly
- Test error cases (missing context, write failures)
- Test tar creation independently

## Size Management Strategy

### Continuous Monitoring
```bash
# After each file implementation:
EFFORT_DIR="/home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase2/wave1/image-builder"
cd $EFFORT_DIR && $CLAUDE_PROJECT_DIR/tools/line-counter.sh

# Check at these milestones:
# - After types.go: Should be ~50 lines
# - After context.go: Should be ~150 lines
# - After builder.go: Should be ~350 lines
# - After storage.go: Should be ~425 lines
# - After tests: Should be ~675 lines
```

### Size Control Measures
1. **No feature creep**: Strictly implement ONLY listed functions
2. **Minimal error handling**: Use fmt.Errorf, not custom error types
3. **No optimization**: Write simple, working code first
4. **Defer advanced features**: No caching, no concurrency
5. **Compact tests**: Focus on critical paths only

### Warning Thresholds
- **500 lines**: Check if on track
- **600 lines**: Start being very selective with tests
- **700 lines**: WARNING - approaching limit, wrap up
- **750 lines**: CRITICAL - finish current function only
- **800 lines**: STOP - Must not exceed

## 🔴 Atomic PR Design (R220 MANDATORY)

### PR Completeness Requirements

```yaml
effort_atomic_pr_design:
  pr_summary: "feat: implement OCI image builder using go-containerregistry"
  can_merge_to_main_alone: true  # MUST be true
  
  feature_flags_needed:
    - flag: "ENABLE_IMAGE_BUILDER"
      purpose: "Enable new image builder functionality"
      default: false
      location: "pkg/build/feature_flags.go"
      activation: "When Wave 2 CLI integration complete"
  
  stubs_required:
    # None - this is foundational effort with no dependencies
    # But we provide minimal interface for future efforts
  
  interfaces_to_implement:
    - interface: "Builder (minimal)"
      methods: ["BuildImage"]
      implementation: "Complete in this PR"
      note: "Other methods (List, Remove, Tag) stubbed with 'not implemented'"
  
  pr_verification:
    tests_pass_alone: true
    build_remains_working: true
    flags_tested_both_ways: true
    no_external_dependencies: true
    backward_compatible: true
  
  example_pr_structure:
    files_added:
      - "pkg/build/builder.go"
      - "pkg/build/context.go"
      - "pkg/build/layer.go"
      - "pkg/build/manifest.go"
      - "pkg/build/storage.go"
      - "pkg/build/types.go"
      - "pkg/build/feature_flags.go"
      - "pkg/build/builder_test.go"
      - "pkg/build/context_test.go"
    tests_included:
      - "Unit tests with flag off (returns error)"
      - "Unit tests with flag on (full functionality)"
      - "Context processing tests"
      - "Storage tests with temp directories"
    documentation:
      - "README update with build package description"
      - "Inline godoc comments"
```

### Independent Mergeability Checklist
- ✅ Code compiles when merged to phase1/integration alone
- ✅ No dependency on E2.1.2 (gitea-client)
- ✅ Feature flag prevents activation until ready
- ✅ All tests pass in isolation
- ✅ No breaking changes to existing code

## Test Requirements

### Unit Test Coverage (Target: 80%)
1. **Builder Tests** (builder_test.go)
   - `TestNewBuilder`: Verify initialization
   - `TestBuildImageSuccess`: Happy path with mock fs
   - `TestBuildImageMissingContext`: Error handling
   - `TestBuildImageWithExclusions`: .dockerignore patterns

2. **Context Tests** (context_test.go)
   - `TestCreateTarFromContext`: Tar creation
   - `TestExclusionPatterns`: Pattern matching
   - `TestEmptyDirectory`: Edge case

### Test Implementation Strategy
```go
// Use testify for assertions
// Use afero for mock filesystem
// Keep tests focused and fast
// No integration tests in this effort (size constraint)
```

## Pattern Compliance

### Go Patterns
- ✅ Error wrapping with context: `fmt.Errorf("failed to X: %w", err)`
- ✅ Context propagation: All operations accept `context.Context`
- ✅ Defer cleanup: Use defer for closing resources
- ✅ Interface segregation: Minimal Builder interface

### Code Style
- ✅ gofmt/goimports compliance
- ✅ golangci-lint passing
- ✅ Exported functions have godoc comments
- ✅ No TODO comments in final code

## Integration Points

### Phase 1 Certificate Infrastructure
- Not directly used in this effort
- E2.1.2 (gitea-client) will handle certificate integration
- This effort focuses purely on local image building

### Wave 1 Integration
- E2.1.1 and E2.1.2 can develop in parallel
- Both merge to phase2/wave1-integration
- Wave 2 CLI will consume both interfaces

## Validation Checkpoints

### Before Starting
- [ ] Verify in correct directory: `efforts/phase2/wave1/image-builder`
- [ ] Confirm on correct branch: `idpbuilder-oci-build-push/phase2/wave1/image-builder`
- [ ] Verify base branch exists: `idpbuilder-oci-build-push/phase1/integration`

### During Implementation (Every 2 Hours)
- [ ] Run line counter: Must be < 700 lines
- [ ] Run tests: Must maintain > 80% coverage
- [ ] Check scope: No feature creep
- [ ] Verify branch isolation: Can merge independently

### Before Completion
- [ ] All tests passing
- [ ] Line count < 800 (use line-counter.sh)
- [ ] No TODO comments
- [ ] Feature flag implemented
- [ ] Code review ready

## Risk Mitigation

### Size Overrun Risk
- **Mitigation**: Strict scope control, continuous measurement
- **Contingency**: Stop at 750 lines, defer remaining to future effort

### Dependency Risk
- **Mitigation**: No external dependencies except go-containerregistry
- **Contingency**: Implement minimal stubs if library issues

### Integration Risk
- **Mitigation**: Clear interface boundaries
- **Contingency**: E2.1.2 can mock Builder interface if needed

## Success Metrics

- ✅ Implementation under 800 lines (measured by line-counter.sh)
- ✅ Test coverage ≥ 80%
- ✅ All tests passing
- ✅ Can build simple directory to OCI image
- ✅ Saves image locally as tarball
- ✅ Feature flag controls activation
- ✅ PR can merge independently

## Notes for SW Engineer

1. **Start Small**: Implement types first, then build up
2. **Test Early**: Write tests alongside implementation
3. **Measure Often**: Run line-counter.sh every 100 lines
4. **Stay Focused**: Resist adding "nice to have" features
5. **Use Mocks**: Mock filesystem for tests to save lines
6. **Simple Errors**: Use fmt.Errorf, not custom error types
7. **Minimal Logging**: Only essential error cases

## Summary

This effort implements the core OCI image building functionality using go-containerregistry. It focuses on a minimal, working implementation that can build a directory into an OCI image and save it locally. The strict scope control ensures we stay well within the 800-line limit while delivering a complete, testable component that can be integrated with the Wave 2 CLI.

**Key Principle**: Do exactly what's specified, nothing more, nothing less. Every line counts toward the limit.