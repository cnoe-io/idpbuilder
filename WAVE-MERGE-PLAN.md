# Wave 2 Merge Plan - Re-integration After Fixes
**Generated:** 2025-09-19T15:41:00Z
**Integration Agent:** P1W2 Re-run
**Context:** Re-integration after bug fixes applied to all branches

## Target Integration Branch
- **Branch Name:** idpbuilder-oci-build-push/phase1-wave2-integration
- **Base:** Phase 1 Wave 1 Integration
- **Purpose:** Clean integration of P1W2 efforts after fixes

## Branches to Merge (NO SPLITS - Per Updated Plan)

### 1. cert-validation (712 lines)
- **Branch:** cert-validation/idpbuilder-oci-build-push/phase1/wave2/cert-validation
- **Status:** Bug fixes applied
- **No splits needed** - Under 800 line limit

### 2. fallback-core (663 lines)
- **Branch:** fallback-core/idpbuilder-oci-build-push/phase1/wave2/fallback-core
- **Status:** Bug fixes applied

### 3. fallback-recommendations (775 lines)
- **Branch:** fallback-rec/idpbuilder-oci-build-push/phase1/wave2/fallback-recommendations
- **Status:** Bug fixes applied

### 4. fallback-security (833 lines)
- **Branch:** fallback-sec/idpbuilder-oci-build-push/phase1/wave2/fallback-security
- **Status:** Bug fixes applied - Slightly over soft limit but under hard limit

## Total Expected Lines: ~2983

## Post-Merge Validation
- Build: go build ./...
- Tests: go test ./...
- Final push to origin