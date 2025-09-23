# IDPBuilder OCI Support - Architect Requirements

## Project Context
IDPBuilder is a tool for setting up Internal Developer Platforms (IDPs) using Kubernetes. It currently uses ArgoCD for application deployment and Gitea as an in-cluster git server.

## Implementation Goal
Add comprehensive OCI registry support to IDPBuilder for:
1. Storing and retrieving package definitions as OCI artifacts
2. Supporting private OCI registries with authentication
3. Enabling package distribution via OCI registries
4. Maintaining backward compatibility with git-based workflows

## Current State
- Basic Docker registry authentication exists (--registry-config flag)
- ArgoCD is used for GitOps deployments
- Gitea provides in-cluster git hosting
- Packages are currently defined as ArgoCD Applications

## Required Features

### 1. OCI Package Storage
- Store IDPBuilder packages as OCI artifacts
- Support versioning and tagging of packages
- Enable package discovery from OCI registries

### 2. Registry Management
- Support multiple OCI registries (Docker Hub, Harbor, Gitea Registry, etc.)
- Handle authentication for private registries
- Registry configuration management

### 3. Package Operations
- Push packages to OCI registries
- Pull packages from OCI registries
- List available packages in registries
- Package validation and verification

### 4. ArgoCD Integration
- Enable ArgoCD to use OCI artifacts as sources
- Maintain compatibility with existing git-based workflows
- Support hybrid deployments (git + OCI)

### 5. CLI Enhancements
- New commands for OCI operations (push, pull, list)
- Registry configuration commands
- Package management commands

## Technical Constraints
- Must maintain single Docker dependency at runtime
- Should not break existing git-based workflows
- Must integrate with current ArgoCD deployment model
- Keep the solution simple and maintainable

## Architecture Considerations
- Use established OCI libraries (go-containerregistry, oras-go)
- Follow OCI Image and Distribution specifications
- Maintain clear separation between git and OCI workflows
- Consider using Gitea's built-in OCI registry support

## Success Criteria
1. Users can push IDPBuilder packages to OCI registries
2. Users can deploy packages from OCI registries via ArgoCD
3. Authentication works with major registry providers
4. Documentation covers all OCI workflows
5. Tests verify OCI functionality

## Deliverables
1. OCI registry client implementation
2. Package format specification for OCI artifacts
3. CLI commands for OCI operations
4. ArgoCD configuration for OCI sources
5. Documentation and examples
6. Comprehensive test suite

## Implementation Phases

### Phase 1: Foundation (Core OCI Support)
- Registry client implementation
- Authentication handling
- Basic push/pull operations

### Phase 2: Package Management
- Package format definition
- Package creation and validation
- Versioning and tagging

### Phase 3: ArgoCD Integration
- Configure ArgoCD for OCI sources
- Implement OCI-to-Application conversion
- Hybrid workflow support

### Phase 4: CLI and UX
- CLI commands for all operations
- User-friendly error messages
- Progress indicators and logging

### Phase 5: Testing and Documentation
- Unit and integration tests
- End-to-end testing
- User documentation
- Migration guides

## Questions for Implementation
1. Should we use Gitea's built-in registry or support external registries primarily?
2. What package format should we use (Helm OCI, custom format)?
3. How should we handle package dependencies?
4. Should packages include ArgoCD Applications or be more generic?
5. What metadata should be included with packages?