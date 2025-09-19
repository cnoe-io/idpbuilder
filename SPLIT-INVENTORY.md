# Complete Split Plan for gitea-client (E2.1.2)
**Sole Planner**: Code Reviewer
**Full Path**: phase2/wave1/gitea-client
**Parent Branch**: software-factory-2.0
**Total Size**: 1268 lines (non-generated)
**Splits Required**: 2
**Created**: 2025-01-09 10:45:00

⚠️ **SPLIT INTEGRITY NOTICE** ⚠️
ALL splits below belong to THIS effort ONLY: phase2/wave1/gitea-client
NO splits should reference efforts outside this path!

## Split Overview
This effort implements a Gitea registry client for managing container images in OCI registries. The implementation includes authentication, push/pull operations, retry logic, and comprehensive testing support.

## Split Boundaries (NO OVERLAPS)
| Split | Lines | Size | Files | Focus | Status |
|-------|-------|------|-------|-------|--------|
| 001   | ~635  | 635  | interface.go, auth.go, gitea.go, remote_options.go | Core interfaces, authentication, and main registry | Planned |
| 002   | ~633  | 633  | push.go, list.go, retry.go, stubs.go | Push/list operations, retry logic, and test stubs | Planned |

## File Distribution Matrix
| File | Lines | Split 001 | Split 002 | Purpose |
|------|-------|-----------|-----------|---------|
| pkg/registry/interface.go | 24 | ✅ | ❌ | Core Registry interface definition |
| pkg/registry/auth.go | 138 | ✅ | ❌ | Authentication and token management |
| pkg/registry/gitea.go | 204 | ✅ | ❌ | Main Gitea registry implementation |
| pkg/registry/remote_options.go | 269 | ✅ | ❌ | Remote registry configuration |
| pkg/registry/push.go | 302 | ❌ | ✅ | Image push operations |
| pkg/registry/list.go | 90 | ❌ | ✅ | Image listing functionality |
| pkg/registry/retry.go | 52 | ❌ | ✅ | Retry logic with exponential backoff |
| pkg/registry/stubs.go | 189 | ❌ | ✅ | Test stubs and mocks |

## Deduplication Matrix
| Module/Feature | Split 001 | Split 002 |
|----------------|-----------|-----------|
| Interface definitions | ✅ | ❌ |
| Authentication | ✅ | ❌ |
| Core registry client | ✅ | ❌ |
| Remote configuration | ✅ | ❌ |
| Push operations | ❌ | ✅ |
| List operations | ❌ | ✅ |
| Retry logic | ❌ | ✅ |
| Test stubs | ❌ | ✅ |

## Dependencies
- **Split 001**: Foundation - no dependencies on Split 002
- **Split 002**: Depends on Split 001 for interfaces and authentication

## Implementation Order
1. **Split 001** - Must be implemented first (core interfaces and authentication)
2. **Split 002** - Can be implemented after Split 001 (operations and utilities)

## Integration Points
- Split 002 will import types and interfaces from Split 001
- Both splits will compile independently
- Full functionality requires both splits merged

## Verification Checklist
- [x] No file appears in multiple splits
- [x] All files from original effort covered
- [x] Each split can compile independently (with proper imports)
- [x] Dependencies properly ordered
- [x] Each split <700 lines (both ~635 lines)
- [x] Logical separation of concerns maintained
- [x] Test stubs isolated in Split 002