# Split Plan for E2.1.2 gitea-client

## Overview
This document outlines the split strategy for the gitea-client effort (E2.1.2), which implements a Gitea registry client for managing container images in OCI registries.

## Split Requirement Reason
- **Current Size**: 1268 lines (measured with line-counter.sh)
- **Limit**: 800 lines per effort
- **Required Splits**: 2

## Split Strategy
The effort has been divided into two logical splits that maintain clean separation of concerns:

### Split 001: Core Interfaces and Authentication (635 lines)
**Focus**: Foundation components including interfaces, authentication, and core registry implementation
**Files**:
- `pkg/registry/interface.go` (24 lines) - Core Registry interface
- `pkg/registry/auth.go` (138 lines) - Authentication logic
- `pkg/registry/gitea.go` (204 lines) - Main Gitea registry client
- `pkg/registry/remote_options.go` (269 lines) - Remote configuration

**Why this grouping**: These files form the foundation that all other operations depend on. They must be implemented first to establish the core contracts and authentication mechanisms.

### Split 002: Operations and Utilities (633 lines)
**Focus**: Image operations (push/list) and supporting utilities
**Files**:
- `pkg/registry/push.go` (302 lines) - Push operations
- `pkg/registry/list.go` (90 lines) - List operations
- `pkg/registry/retry.go` (52 lines) - Retry logic
- `pkg/registry/stubs.go` (189 lines) - Test stubs

**Why this grouping**: These files implement the actual registry operations and testing utilities. They depend on the interfaces and authentication from Split 001.

## Implementation Order
1. **Split 001** must be implemented first (foundation)
2. **Split 002** can only start after Split 001 is complete (depends on interfaces)

## Branch Strategy
- Base branch: `software-factory-2.0`
- Split 001 branch: `phase2/wave1/gitea-client-split-001`
- Split 002 branch: `phase2/wave1/gitea-client-split-002` (branches from split-001)

## Integration Plan
1. Complete Split 001 implementation and review
2. Merge Split 001 to base
3. Complete Split 002 implementation and review
4. Merge Split 002 to base
5. Final integration testing

## Files Created
- `SPLIT-INVENTORY.md` - Complete split matrix and deduplication tracking
- `SPLIT-PLAN-001.md` - Detailed plan for Split 001
- `SPLIT-PLAN-002.md` - Detailed plan for Split 002

## Verification
- No file appears in multiple splits ✅
- Each split is under 700 lines ✅
- Logical separation maintained ✅
- Dependencies properly ordered ✅
- Complete functionality preserved ✅