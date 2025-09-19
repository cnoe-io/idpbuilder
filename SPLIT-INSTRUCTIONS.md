# SPLIT-PLAN-001.md
## Split 001 of 2: OCI Types and Documentation
**Planner**: Code Reviewer @agent-code-reviewer (same for ALL splits)
**Parent Effort**: registry-auth-types
**Target Size**: 661 lines (well under 800 limit)

## Boundaries
- **Previous Split**: None (first split)
- **This Split Focus**: OCI types, manifest handling, and package documentation
- **Next Split**: Stack types (Split 002)

## Files in This Split (EXCLUSIVE - no overlap with other splits)
```
pkg/doc.go                    (39 lines)  - Package documentation
pkg/oci/types.go              (121 lines) - OCI type definitions
pkg/oci/manifest.go           (124 lines) - Manifest handling logic
pkg/oci/constants.go          (56 lines)  - OCI-related constants
pkg/oci/types_test.go         (130 lines) - Unit tests for types
pkg/oci/manifest_test.go      (191 lines) - Unit tests for manifest
```
**Total**: 661 lines

## Functionality Scope
### Core Components:
1. **OCI Type Definitions** (pkg/oci/types.go)
   - Image configuration types
   - Registry reference types
   - Platform specifications
   - Descriptor structures

2. **Manifest Handling** (pkg/oci/manifest.go)
   - Manifest parsing and validation
   - Manifest list operations
   - Content descriptor management
   - Media type handling

3. **Constants** (pkg/oci/constants.go)
   - Media type constants
   - Architecture constants
   - OS platform constants
   - Annotation keys

4. **Package Documentation** (pkg/doc.go)
   - Overall package overview
   - Usage examples
   - Architecture notes

5. **Test Coverage**
   - Complete unit tests for all OCI types
   - Manifest operation tests
   - Edge case coverage

## Dependencies
- **External**: Standard library only (encoding/json, crypto/sha256, etc.)
- **Internal**: None - this is a foundational package
- **Test Dependencies**: Standard testing package

## Implementation Instructions for SW Engineer

### Step 1: Create Branch
```bash
git checkout -b phase1/wave1/registry-auth-types-split-001
```

### Step 2: Sparse Checkout (if starting fresh)
```bash
# Enable sparse checkout
git sparse-checkout init --cone
git sparse-checkout set pkg/doc.go pkg/oci/
```

### Step 3: Verify Files
Ensure ONLY these files are included:
- pkg/doc.go
- pkg/oci/types.go
- pkg/oci/manifest.go
- pkg/oci/constants.go
- pkg/oci/types_test.go
- pkg/oci/manifest_test.go

### Step 4: Run Tests
```bash
go test ./pkg/oci/...
```

### Step 5: Measure Size
```bash
/workspaces/idpbuilder-oci-mgmt/tools/line-counter.sh
# Should show ~661 lines
```

### Step 6: Commit
```bash
git add pkg/doc.go pkg/oci/
git commit -m "feat: implement OCI types and manifest handling (split 001)"
git push origin phase1/wave1/registry-auth-types-split-001
```

## Quality Checklist
- [ ] All OCI types properly defined with json tags
- [ ] Manifest operations handle all media types
- [ ] Constants cover standard OCI specifications
- [ ] Tests achieve >80% coverage
- [ ] Documentation includes usage examples
- [ ] No circular dependencies
- [ ] Clean separation from Stack package

## Merge Strategy
- This split will be merged to `phase1/wave1/registry-auth-types` branch
- Can be merged independently of Split 002
- No coordination required with other splits

## Risk Mitigation
- **Compilation**: Package is self-contained and will compile independently
- **Testing**: Includes all test files for complete validation
- **Size**: At 661 lines, well under the 800-line limit with room for minor additions