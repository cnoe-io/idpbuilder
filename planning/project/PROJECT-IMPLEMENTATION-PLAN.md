# IDPBuilder OCI Support - Master Implementation Plan

## Executive Summary

This implementation plan adds comprehensive OCI (Open Container Initiative) registry support to IDPBuilder, enabling package storage and distribution via OCI registries while maintaining backward compatibility with existing git-based workflows. The implementation follows a parallel OCI provider approach, creating a clean abstraction layer that allows both Docker/kind and OCI workflows to coexist.

### Key Architectural Decisions

1. **Parallel Implementation Strategy**: Create OCI provider alongside existing Docker implementation rather than refactoring, minimizing risk to existing functionality
2. **Provider Abstraction Layer**: Define clean interfaces in Phase 1 to enable maximum parallelization
3. **Registry-Agnostic Design**: Support multiple registry types (Docker Hub, Harbor, Gitea Registry, etc.)
4. **Incremental Integration**: Each phase adds value independently while building toward complete OCI support
5. **Interface-First Development**: All contracts and interfaces defined upfront in Phase 1

### Implementation Approach

- **Total Effort Count**: 38 efforts across 4 phases
- **Maximum Parallelization**: 8-10 parallel efforts possible after Phase 1 Wave 1
- **Effort Size Constraint**: Each effort under 800 lines of code
- **Testing Strategy**: Unit tests within efforts, integration tests as separate efforts
- **Documentation**: Progressive documentation with each phase

---

## Phase 1: Foundation & Interfaces (12 efforts)

**Objective**: Establish all interfaces, contracts, and foundational OCI registry client implementation

### Wave 1: Core Interfaces & Abstractions (4 efforts, all parallel)

#### P1W1-E1: Provider Interface Definition
- **Size**: 150 lines
- **Dependencies**: None
- **Description**: Define provider abstraction interfaces for registry operations
- **Files**:
  - `pkg/providers/interface.go` - Core provider interface
  - `pkg/providers/types.go` - Common types and structures
  - `pkg/providers/errors.go` - Error definitions

#### P1W1-E2: OCI Package Format Specification
- **Size**: 200 lines
- **Dependencies**: None
- **Description**: Define OCI artifact format for IDPBuilder packages
- **Files**:
  - `pkg/oci/format/spec.go` - Package format specification
  - `pkg/oci/format/manifest.go` - Manifest structure
  - `pkg/oci/format/metadata.go` - Metadata definitions

#### P1W1-E3: Registry Configuration Schema
- **Size**: 180 lines
- **Dependencies**: None
- **Description**: Define configuration schema for registry connections
- **Files**:
  - `pkg/config/registry.go` - Registry configuration
  - `pkg/config/auth.go` - Authentication configuration
  - `pkg/config/validation.go` - Configuration validation

#### P1W1-E4: CLI Interface Contracts
- **Size**: 120 lines
- **Dependencies**: None
- **Description**: Define interfaces for new CLI commands
- **Files**:
  - `pkg/cmd/interfaces/oci.go` - OCI command interfaces
  - `pkg/cmd/interfaces/registry.go` - Registry command interfaces

### Wave 2: OCI Client Implementation (4 efforts, 2 can be parallel)

#### P1W2-E1: Base OCI Registry Client
- **Size**: 600 lines
- **Dependencies**: P1W1-E1, P1W1-E3
- **Description**: Implement core OCI registry client using go-containerregistry
- **Files**:
  - `pkg/oci/client/client.go` - Main client implementation
  - `pkg/oci/client/transport.go` - HTTP transport configuration
  - `pkg/oci/client/retry.go` - Retry logic

#### P1W2-E2: Authentication Handler
- **Size**: 400 lines
- **Dependencies**: P1W1-E3
- **Description**: Implement multi-registry authentication support
- **Files**:
  - `pkg/oci/auth/handler.go` - Authentication handler
  - `pkg/oci/auth/docker.go` - Docker config support
  - `pkg/oci/auth/token.go` - Token-based auth
  - `pkg/oci/auth/basic.go` - Basic auth support

#### P1W2-E3: Push Operations Implementation
- **Size**: 500 lines
- **Dependencies**: P1W2-E1, P1W1-E2
- **Description**: Implement package push functionality
- **Files**:
  - `pkg/oci/operations/push.go` - Push implementation
  - `pkg/oci/operations/manifest.go` - Manifest handling
  - `pkg/oci/operations/layer.go` - Layer management

#### P1W2-E4: Pull Operations Implementation
- **Size**: 450 lines
- **Dependencies**: P1W2-E1, P1W1-E2
- **Description**: Implement package pull functionality
- **Files**:
  - `pkg/oci/operations/pull.go` - Pull implementation
  - `pkg/oci/operations/extract.go` - Artifact extraction
  - `pkg/oci/operations/verify.go` - Verification logic

### Wave 3: Registry Provider Implementations (4 efforts, all parallel after Wave 2)

#### P1W3-E1: Docker Hub Provider
- **Size**: 350 lines
- **Dependencies**: P1W2-E1, P1W2-E2
- **Description**: Docker Hub specific provider implementation
- **Files**:
  - `pkg/providers/dockerhub/provider.go` - Provider implementation
  - `pkg/providers/dockerhub/auth.go` - Docker Hub auth

#### P1W3-E2: Harbor Provider
- **Size**: 380 lines
- **Dependencies**: P1W2-E1, P1W2-E2
- **Description**: Harbor registry provider implementation
- **Files**:
  - `pkg/providers/harbor/provider.go` - Provider implementation
  - `pkg/providers/harbor/project.go` - Project management

#### P1W3-E3: Gitea Registry Provider
- **Size**: 400 lines
- **Dependencies**: P1W2-E1, P1W2-E2
- **Description**: Gitea built-in registry provider
- **Files**:
  - `pkg/providers/gitea/provider.go` - Provider implementation
  - `pkg/providers/gitea/integration.go` - Gitea integration

#### P1W3-E4: Generic OCI Provider
- **Size**: 300 lines
- **Dependencies**: P1W2-E1, P1W2-E2
- **Description**: Generic OCI registry provider for standard registries
- **Files**:
  - `pkg/providers/generic/provider.go` - Generic provider
  - `pkg/providers/generic/discovery.go` - Registry discovery

---

## Phase 2: Package Management & Core Features (10 efforts)

**Objective**: Implement package creation, validation, and management capabilities

### Wave 1: Package Operations (4 efforts, 3 can be parallel)

#### P2W1-E1: Package Builder
- **Size**: 550 lines
- **Dependencies**: P1W1-E2
- **Description**: Build IDPBuilder packages as OCI artifacts
- **Files**:
  - `pkg/package/builder.go` - Package builder
  - `pkg/package/bundler.go` - Resource bundling
  - `pkg/package/compress.go` - Compression utilities

#### P2W1-E2: Package Validator
- **Size**: 400 lines
- **Dependencies**: P1W1-E2
- **Description**: Validate package structure and contents
- **Files**:
  - `pkg/package/validator.go` - Validation logic
  - `pkg/package/schema.go` - Schema validation
  - `pkg/package/rules.go` - Validation rules

#### P2W1-E3: Version Manager
- **Size**: 350 lines
- **Dependencies**: P1W1-E2
- **Description**: Handle package versioning and tagging
- **Files**:
  - `pkg/package/version.go` - Version management
  - `pkg/package/semver.go` - Semantic versioning
  - `pkg/package/tags.go` - Tag management

#### P2W1-E4: Dependency Resolver
- **Size**: 450 lines
- **Dependencies**: P2W1-E3
- **Description**: Resolve package dependencies
- **Files**:
  - `pkg/package/resolver.go` - Dependency resolver
  - `pkg/package/graph.go` - Dependency graph
  - `pkg/package/conflicts.go` - Conflict resolution

### Wave 2: Registry Operations (3 efforts, all parallel after Wave 1)

#### P2W2-E1: Registry Discovery Service
- **Size**: 400 lines
- **Dependencies**: Phase 1 complete
- **Description**: Discover and catalog available packages
- **Files**:
  - `pkg/registry/discovery.go` - Discovery service
  - `pkg/registry/catalog.go` - Package catalog
  - `pkg/registry/search.go` - Search functionality

#### P2W2-E2: Package Cache Manager
- **Size**: 450 lines
- **Dependencies**: Phase 1 complete
- **Description**: Local caching of pulled packages
- **Files**:
  - `pkg/cache/manager.go` - Cache manager
  - `pkg/cache/storage.go` - Storage backend
  - `pkg/cache/eviction.go` - Cache eviction

#### P2W2-E3: Registry Sync Service
- **Size**: 500 lines
- **Dependencies**: P2W2-E1
- **Description**: Synchronize packages between registries
- **Files**:
  - `pkg/sync/service.go` - Sync service
  - `pkg/sync/replication.go` - Replication logic
  - `pkg/sync/scheduler.go` - Sync scheduling

### Wave 3: Package Controllers (3 efforts, sequential)

#### P2W3-E1: OCI Package Controller
- **Size**: 600 lines
- **Dependencies**: P2W1-E1, P2W2-E1
- **Description**: Kubernetes controller for OCI packages
- **Files**:
  - `pkg/controllers/ocipackage/controller.go` - Controller
  - `pkg/controllers/ocipackage/reconciler.go` - Reconciliation
  - `pkg/controllers/ocipackage/status.go` - Status management

#### P2W3-E2: Package CRD Definition
- **Size**: 250 lines
- **Dependencies**: P2W3-E1
- **Description**: Custom Resource Definitions for OCI packages
- **Files**:
  - `api/v1alpha1/ocipackage_types.go` - CRD types
  - `api/v1alpha1/ocipackage_webhook.go` - Webhooks

#### P2W3-E3: Package Lifecycle Manager
- **Size**: 450 lines
- **Dependencies**: P2W3-E1, P2W3-E2
- **Description**: Manage package lifecycle events
- **Files**:
  - `pkg/lifecycle/manager.go` - Lifecycle manager
  - `pkg/lifecycle/events.go` - Event handling
  - `pkg/lifecycle/hooks.go` - Lifecycle hooks

---

## Phase 3: Integration & CLI (10 efforts)

**Objective**: Integrate OCI support with ArgoCD and implement comprehensive CLI

### Wave 1: ArgoCD Integration (4 efforts)

#### P3W1-E1: ArgoCD OCI Source Plugin
- **Size**: 600 lines
- **Dependencies**: Phase 2 complete
- **Description**: ArgoCD plugin for OCI sources
- **Files**:
  - `pkg/argocd/plugin/main.go` - Plugin entry point
  - `pkg/argocd/plugin/generate.go` - Manifest generation
  - `pkg/argocd/plugin/download.go` - OCI download

#### P3W1-E2: Application Converter
- **Size**: 450 lines
- **Dependencies**: P3W1-E1
- **Description**: Convert OCI packages to ArgoCD Applications
- **Files**:
  - `pkg/argocd/converter.go` - Conversion logic
  - `pkg/argocd/templates.go` - Application templates
  - `pkg/argocd/mapping.go` - Resource mapping

#### P3W1-E3: Hybrid Workflow Manager
- **Size**: 500 lines
- **Dependencies**: P3W1-E2
- **Description**: Support mixed git and OCI sources
- **Files**:
  - `pkg/workflow/hybrid.go` - Hybrid workflow
  - `pkg/workflow/selector.go` - Source selection
  - `pkg/workflow/merger.go` - Source merging

#### P3W1-E4: ArgoCD Configuration Manager
- **Size**: 350 lines
- **Dependencies**: P3W1-E1
- **Description**: Configure ArgoCD for OCI support
- **Files**:
  - `pkg/argocd/config/manager.go` - Config manager
  - `pkg/argocd/config/patches.go` - Configuration patches

### Wave 2: CLI Implementation (6 efforts, 4 can be parallel)

#### P3W2-E1: Push Command Implementation
- **Size**: 400 lines
- **Dependencies**: Phase 2 complete
- **Description**: CLI command for pushing packages
- **Files**:
  - `pkg/cmd/push/root.go` - Push command
  - `pkg/cmd/push/validate.go` - Input validation
  - `pkg/cmd/push/progress.go` - Progress reporting

#### P3W2-E2: Pull Command Implementation
- **Size**: 380 lines
- **Dependencies**: Phase 2 complete
- **Description**: CLI command for pulling packages
- **Files**:
  - `pkg/cmd/pull/root.go` - Pull command
  - `pkg/cmd/pull/extract.go` - Extraction options

#### P3W2-E3: List Command Implementation
- **Size**: 350 lines
- **Dependencies**: Phase 2 complete
- **Description**: CLI command for listing packages
- **Files**:
  - `pkg/cmd/list/root.go` - List command
  - `pkg/cmd/list/format.go` - Output formatting
  - `pkg/cmd/list/filter.go` - Filtering options

#### P3W2-E4: Registry Command Implementation
- **Size**: 400 lines
- **Dependencies**: Phase 2 complete
- **Description**: CLI commands for registry management
- **Files**:
  - `pkg/cmd/registry/root.go` - Registry command
  - `pkg/cmd/registry/add.go` - Add registry
  - `pkg/cmd/registry/config.go` - Configure registry

#### P3W2-E5: Package Command Enhancement
- **Size**: 450 lines
- **Dependencies**: P3W2-E1, P3W2-E2
- **Description**: Enhance existing package command with OCI support
- **Files**:
  - `pkg/cmd/package/oci.go` - OCI operations
  - `pkg/cmd/package/build.go` - Package building
  - `pkg/cmd/package/publish.go` - Publishing

#### P3W2-E6: CLI Integration & Help
- **Size**: 300 lines
- **Dependencies**: P3W2-E1, P3W2-E2, P3W2-E3, P3W2-E4
- **Description**: Integrate new commands and update help system
- **Files**:
  - `pkg/cmd/root.go` - Root command updates
  - `pkg/cmd/help/oci.go` - OCI help topics

---

## Phase 4: Testing, Documentation & Finalization (6 efforts)

**Objective**: Comprehensive testing, documentation, and production readiness

### Wave 1: Testing Infrastructure (3 efforts, all parallel)

#### P4W1-E1: Unit Test Suite
- **Size**: 700 lines
- **Dependencies**: Phase 3 complete
- **Description**: Comprehensive unit tests for OCI functionality
- **Files**:
  - `pkg/oci/client/client_test.go`
  - `pkg/oci/operations/push_test.go`
  - `pkg/oci/operations/pull_test.go`
  - `pkg/package/builder_test.go`

#### P4W1-E2: Integration Test Suite
- **Size**: 600 lines
- **Dependencies**: Phase 3 complete
- **Description**: Integration tests with real registries
- **Files**:
  - `tests/integration/oci/push_pull_test.go`
  - `tests/integration/oci/registry_test.go`
  - `tests/integration/oci/argocd_test.go`

#### P4W1-E3: E2E Test Suite
- **Size**: 500 lines
- **Dependencies**: Phase 3 complete
- **Description**: End-to-end workflow tests
- **Files**:
  - `tests/e2e/oci/workflow_test.go`
  - `tests/e2e/oci/hybrid_test.go`
  - `tests/e2e/oci/migration_test.go`

### Wave 2: Documentation & Examples (3 efforts, all parallel after Wave 1)

#### P4W2-E1: User Documentation
- **Size**: 400 lines
- **Dependencies**: P4W1-E3
- **Description**: User guides and documentation
- **Files**:
  - `docs/oci/getting-started.md`
  - `docs/oci/registry-setup.md`
  - `docs/oci/package-management.md`
  - `docs/oci/cli-reference.md`

#### P4W2-E2: Migration Guide
- **Size**: 300 lines
- **Dependencies**: P4W1-E3
- **Description**: Guide for migrating from git to OCI
- **Files**:
  - `docs/migration/git-to-oci.md`
  - `docs/migration/hybrid-setup.md`
  - `scripts/migration/migrate.sh`

#### P4W2-E3: Example Packages & Configurations
- **Size**: 350 lines
- **Dependencies**: P4W1-E3
- **Description**: Example packages and configurations
- **Files**:
  - `examples/oci/packages/sample-app/`
  - `examples/oci/registry-configs/`
  - `examples/oci/argocd-apps/`

---

## Dependencies and Parallelization Matrix

### Phase 1 Parallelization
- Wave 1: 4 efforts fully parallel (P1W1-E1 to E4)
- Wave 2: 2 parallel groups (E1+E2, then E3+E4)
- Wave 3: 4 efforts fully parallel

### Phase 2 Parallelization
- Wave 1: 3 parallel efforts after E1
- Wave 2: 3 efforts fully parallel
- Wave 3: Sequential execution required

### Phase 3 Parallelization
- Wave 1: Sequential with some parallelization
- Wave 2: 4 efforts can run in parallel

### Phase 4 Parallelization
- Wave 1: 3 efforts fully parallel
- Wave 2: 3 efforts fully parallel

### Maximum Concurrent Efforts
- Phase 1: Up to 4 parallel efforts
- Phase 2: Up to 3 parallel efforts
- Phase 3: Up to 4 parallel efforts
- Phase 4: Up to 3 parallel efforts

---

## Risk Mitigation

### Technical Risks
1. **Registry Compatibility**: Mitigated by provider abstraction layer
2. **Breaking Changes**: Parallel implementation preserves existing functionality
3. **Performance**: Caching layer and efficient client implementation
4. **Security**: Comprehensive authentication handling from Phase 1

### Implementation Risks
1. **Scope Creep**: Strictly defined effort boundaries
2. **Integration Complexity**: Interface-first design reduces integration issues
3. **Testing Coverage**: Dedicated testing phase ensures quality

---

## Success Metrics

### Phase 1 Success Criteria
- All interfaces defined and documented
- Basic OCI push/pull operations working
- Authentication working with major registries

### Phase 2 Success Criteria
- Package building and validation functional
- Registry discovery and cataloging operational
- Controllers managing OCI packages

### Phase 3 Success Criteria
- ArgoCD successfully deploying from OCI sources
- All CLI commands functional and documented
- Hybrid workflows operational

### Phase 4 Success Criteria
- All tests passing (unit, integration, E2E)
- Complete documentation available
- Migration path validated

---

## Implementation Timeline Estimate

### Phase Duration (with parallel execution)
- **Phase 1**: 2-3 weeks (12 efforts with high parallelization)
- **Phase 2**: 2-3 weeks (10 efforts with moderate parallelization)
- **Phase 3**: 2-3 weeks (10 efforts with mixed parallelization)
- **Phase 4**: 1-2 weeks (6 efforts with high parallelization)

### Total Timeline: 7-11 weeks

---

## Configuration Management

### Environment Variables
```
IDPBUILDER_OCI_CACHE_DIR=/var/cache/idpbuilder/oci
IDPBUILDER_OCI_TIMEOUT=300
IDPBUILDER_OCI_RETRY_COUNT=3
```

### Configuration Files
```
~/.idpbuilder/registries.yaml  # Registry configurations
~/.idpbuilder/oci-cache/        # Local package cache
```

---

## Backward Compatibility

### Preserved Functionality
- All existing git-based workflows remain unchanged
- Current CLI commands continue to work
- Existing ArgoCD applications unaffected

### Gradual Migration Path
1. Enable OCI support alongside git
2. Test OCI workflows in parallel
3. Migrate packages incrementally
4. Maintain hybrid setup as needed

---

## Appendix: Technology Stack

### Core Libraries
- **go-containerregistry**: Google's OCI registry client
- **oras-go**: OCI Registry As Storage library
- **go-digest**: Content addressable digests

### Standards Compliance
- OCI Image Specification v1.0
- OCI Distribution Specification v1.0
- Docker Registry HTTP API V2

### Registry Support Matrix
| Registry | Auth Method | Tested | Priority |
|----------|-------------|---------|----------|
| Docker Hub | Token | Yes | High |
| Harbor | Basic/Token | Yes | High |
| Gitea | Basic | Yes | High |
| GitHub Container Registry | Token | Yes | Medium |
| AWS ECR | IAM | Planned | Medium |
| Azure ACR | Token | Planned | Low |