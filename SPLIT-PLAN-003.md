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

## 🚨 SPLIT INFRASTRUCTURE METADATA (Added by Orchestrator)
**WORKING_DIRECTORY**: /home/vscode/workspaces/idpbuilder-oci-build-push/efforts/phase1/wave2/cert-validation-SPLIT-003
**BRANCH**: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003
**REMOTE**: origin/idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-003
**BASE_BRANCH**: idpbuilder-oci-build-push/phase1/wave2/cert-validation-split-002
**SPLIT_NUMBER**: 003
**CREATED_AT**: 2025-09-07 19:56:00

### SW Engineer Instructions
1. READ this metadata FIRST
2. cd to WORKING_DIRECTORY above
3. Verify branch matches BRANCH above
4. ONLY THEN proceed with implementation
