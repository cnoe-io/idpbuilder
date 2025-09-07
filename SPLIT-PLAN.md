# Split Plan for Certificate Validation Pipeline Effort

## Current Situation
**Problem**: Effort size exceeds hard limit (808 lines > 800 lines)
**Solution**: Split into 3 manageable sub-efforts with clear boundaries
**Created**: 2025-09-07 15:23:29 UTC
**Sole Planner**: Code Reviewer Agent

## Complete Split Inventory
**Parent Effort**: certificate-validation-pipeline (E1.2.1)
**Full Path**: phase1/wave2/cert-validation
**Parent Branch**: idpbuilder-oci-build-push/phase1/wave2/cert-validation
**Total Size**: 808 lines (measured via tools/line-counter.sh)
**Splits Required**: 3

⚠️ **SPLIT INTEGRITY NOTICE** ⚠️
ALL splits below belong to THIS effort ONLY: phase1/wave2/cert-validation
NO splits should reference efforts outside this path!

## Split Boundaries (NO OVERLAPS)

| Split | Description | Target Size | Actual Files | Status |
|-------|------------|-------------|--------------|--------|
| 001 | Core Types & Errors | 200 lines | validation_errors.go, diagnostics.go | Planned |
| 002 | Certificate Validator | 350 lines | validator.go, interfaces | Planned |
| 003 | Chain Validator & Tests | 350 lines | chain_validator.go, tests | Planned |

## Deduplication Matrix

| File/Module | Split 001 | Split 002 | Split 003 |
|-------------|-----------|-----------|-----------|
| validation_errors.go | ✅ | ❌ | ❌ |
| diagnostics.go | ✅ | ❌ | ❌ |
| validator.go | ❌ | ✅ | ❌ |
| TrustStoreProvider interface | ❌ | ✅ | ❌ |
| chain_validator.go | ❌ | ❌ | ✅ |
| validator_test.go | ❌ | ❌ | ✅ |
| chain_validator_test.go | ❌ | ❌ | ✅ |

## Verification Checklist
- [ ] No file appears in multiple splits
- [ ] All files from original effort covered
- [ ] Each split compiles independently
- [ ] Dependencies properly ordered
- [ ] Each split <400 lines (well under 800 limit)
- [ ] Tests included in appropriate split

---

# SPLIT-PLAN-001.md
## Split 001 of 3: Core Types and Error Definitions
**Planner**: Code Reviewer Agent
**Parent Effort**: certificate-validation-pipeline
**Branch**: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001

<!-- ⚠️ ORCHESTRATOR METADATA PLACEHOLDER - DO NOT REMOVE ⚠️ -->
<!-- The orchestrator will add infrastructure metadata below: -->
<!-- WORKING_DIRECTORY, BRANCH, REMOTE, BASE_BRANCH, etc. -->
<!-- SW Engineers MUST read this metadata to navigate to the correct directory -->
<!-- END PLACEHOLDER -->

### Boundaries
- **Previous Split**: None (first split of THIS effort)
  - Path: N/A (this is Split 001)
  - Branch: N/A
- **This Split**: Split 001 of phase1/wave2/cert-validation
  - Path: efforts/phase1/wave2/cert-validation/split-001/
  - Branch: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001
- **Next Split**: Split 002 of phase1/wave2/cert-validation
  - Path: efforts/phase1/wave2/cert-validation/split-002/
  - Branch: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002

### Files in This Split (EXCLUSIVE - no overlap with other splits)
- pkg/certs/validation_errors.go (117 lines) - All validation error types and error handling
- pkg/certs/diagnostics.go (28 lines) - Certificate diagnostics structure

### Functionality
- Define all validation error types (InvalidCertificate, Expired, NotYetValid, etc.)
- Implement ValidationError structure with proper error formatting
- Define CertDiagnostics structure for diagnostic information
- Provide error categorization and string representations

### Dependencies
- None (foundational split - no dependencies on other splits)
- Standard library only (time, fmt, strings)

### Implementation Instructions
1. Create pkg/certs directory structure in split-001 workspace
2. Implement validation_errors.go with all error types:
   - ValidationErrorType enum
   - ValidationError struct
   - NewValidationError constructor
   - Error() method implementation
   - String representations for all error types
3. Implement diagnostics.go with CertDiagnostics struct
4. Ensure all types are properly exported
5. Add comprehensive godoc comments
6. Measure with ${PROJECT_ROOT}/tools/line-counter.sh

### Acceptance Criteria
- All error types defined and documented
- Diagnostics structure complete
- No external dependencies beyond standard library
- Compiles independently
- Under 200 lines total

---

# SPLIT-PLAN-002.md
## Split 002 of 3: Certificate Validator Implementation
**Planner**: Code Reviewer Agent
**Parent Effort**: certificate-validation-pipeline
**Branch**: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002

### Boundaries
- **Previous Split**: Split 001 of phase1/wave2/cert-validation
  - Path: efforts/phase1/wave2/cert-validation/split-001/
  - Branch: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001
  - Summary: Core types, error definitions, and diagnostics
- **This Split**: Split 002 of phase1/wave2/cert-validation
  - Path: efforts/phase1/wave2/cert-validation/split-002/
  - Branch: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002
- **Next Split**: Split 003 of phase1/wave2/cert-validation
  - Path: efforts/phase1/wave2/cert-validation/split-003/
  - Branch: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003

### Files in This Split (EXCLUSIVE - no overlap with other splits)
- pkg/certs/validator.go (254 lines) - Main validator implementation and interfaces

### Functionality
- ValidationMode enum (StrictMode, LenientMode, PermissiveMode)
- CertificateValidator interface definition
- TrustStoreProvider interface (for Wave 1 integration)
- DefaultCertificateValidator implementation
- ValidationResult structure
- Basic certificate validation methods

### Dependencies
- Requires Split 001 (imports ValidationError types and CertDiagnostics)
- Standard library (crypto/x509, time)

### Implementation Instructions
1. Import types from Split 001 (validation_errors.go, diagnostics.go)
2. Define ValidationMode enum with three modes
3. Create CertificateValidator interface with all required methods:
   - ValidateChain
   - ValidateCertificate
   - VerifyHostname
   - GenerateDiagnostics
   - SetValidationMode
4. Define TrustStoreProvider interface for trust store integration
5. Implement DefaultCertificateValidator struct
6. Implement all interface methods with proper mode handling
7. Add constructor NewDefaultCertificateValidator
8. Measure with ${PROJECT_ROOT}/tools/line-counter.sh

### Acceptance Criteria
- All interfaces properly defined
- DefaultCertificateValidator fully implements CertificateValidator
- Proper integration points for TrustStoreProvider
- Validation modes properly handled
- Under 350 lines total

---

# SPLIT-PLAN-003.md
## Split 003 of 3: Chain Validator and Comprehensive Tests
**Planner**: Code Reviewer Agent
**Parent Effort**: certificate-validation-pipeline
**Branch**: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003

### Boundaries
- **Previous Split**: Split 002 of phase1/wave2/cert-validation
  - Path: efforts/phase1/wave2/cert-validation/split-002/
  - Branch: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002
  - Summary: Certificate validator implementation and interfaces
- **This Split**: Split 003 of phase1/wave2/cert-validation
  - Path: efforts/phase1/wave2/cert-validation/split-003/
  - Branch: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003
- **Next Split**: None (final split)

### Files in This Split (EXCLUSIVE - no overlap with other splits)
- pkg/certs/chain_validator.go (309 lines) - Chain validation logic
- pkg/certs/validator_test.go (new, ~40 lines) - Tests for validator
- pkg/certs/chain_validator_test.go (new, ~40 lines) - Tests for chain validator

### Functionality
- ChainValidator struct implementation
- ChainValidationOptions configuration
- Complete certificate chain validation logic
- Chain ordering and trust verification
- Comprehensive test coverage for all validators

### Dependencies
- Requires Split 001 (imports error types and diagnostics)
- Requires Split 002 (uses TrustStoreProvider interface, ValidationMode)
- Standard library (crypto/x509, testing)

### Implementation Instructions
1. Import types from Split 001 and Split 002
2. Implement ChainValidator struct with:
   - trustStore field (TrustStoreProvider)
   - mode field (ValidationMode)
3. Define ChainValidationOptions struct
4. Implement NewChainValidator constructor
5. Implement ValidateChain method with complete logic:
   - Chain length validation
   - Certificate ordering verification
   - Trust chain validation
   - Signature verification
6. Add helper methods for validation options based on mode
7. Create comprehensive test files:
   - validator_test.go for DefaultCertificateValidator
   - chain_validator_test.go for ChainValidator
8. Ensure test coverage >80%
9. Measure with ${PROJECT_ROOT}/tools/line-counter.sh

### Acceptance Criteria
- Complete chain validation implementation
- Proper error handling using Split 001's error types
- Integration with Split 002's interfaces
- Comprehensive test coverage
- Under 400 lines total (including tests)

---

## Integration Strategy

### Sequential Execution Order
1. **Split 001**: Foundation - types and errors (no dependencies)
2. **Split 002**: Core validator (depends on Split 001)
3. **Split 003**: Chain validator and tests (depends on Split 001 & 002)

### Merge Strategy
After all splits are complete and reviewed:
1. Merge Split 001 to parent branch
2. Merge Split 002 to parent branch
3. Merge Split 003 to parent branch
4. Final integration testing on parent branch
5. Ready for Wave 2 integration

### Risk Mitigation
- Each split compiles independently (with stated dependencies)
- Clear interface boundaries prevent coupling
- Tests in final split validate entire implementation
- No file duplication ensures clean merges
## 🚨 SPLIT INFRASTRUCTURE METADATA (Added by Orchestrator)
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave2/cert-validation-SPLIT-001
**BRANCH**: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001
**REMOTE**: origin/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-001
**BASE_BRANCH**: idpbuilder-oci-build-push/phase1/wave1/integration
**SPLIT_NUMBER**: 001
**CREATED_AT**: $(date '+%Y-%m-%d %H:%M:%S')

### SW Engineer Instructions
1. READ this metadata FIRST
2. cd to WORKING_DIRECTORY above
3. Verify branch matches BRANCH above
4. ONLY THEN proceed with implementation
5. Implement ONLY Split-001 (validation core) from this plan
