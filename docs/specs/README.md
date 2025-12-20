# Technical Specifications

This directory contains technical specifications and architectural design documents for IDP Builder.

## Documents

### [Controller-Based Architecture Specification](./controller-architecture-spec.md)

**Status:** Proposal  
**Version:** 1.0 Draft  
**Date:** December 19, 2025

A comprehensive specification for evolving idpbuilder from a CLI-driven installation model to a controller-based architecture. This architectural change enables:

- Kubernetes-native management through Custom Resources
- Separation between CLI (infrastructure provisioning) and controllers (platform reconciliation)
- Support for both development (CLI-driven) and production (GitOps-driven) deployment modes
- Duck-typed provider interfaces for Git, Gateway, and GitOps components
- Pluggable provider architecture (Gitea/GitHub/GitLab, Nginx/Envoy/Istio, ArgoCD/Flux)

**Key Topics:**
- Platform CR and provider CRs design
- Duck-typing pattern for provider independence
- Phased implementation plan
- Migration strategy from v1alpha1 to v1alpha2

### [Pluggable and Configurable Packaging Proposal](./pluggable-packages.md)

**Status:** Implemented  
**Date:** Original proposal

A design document outlining the approach for making packages installed by idpbuilder configurable and pluggable.

**Key Topics:**
- ArgoCD-based package management
- In-cluster Git server (Gitea) for GitOps workflows
- Runtime content generation
- Support for Helm charts, Kustomize, and raw manifests
- Local file handling for development workflows

**Goals:**
- Make packages configurable without recompiling
- Minimize runtime dependencies
- Enable fast local development feedback loops
- Support imperative pipelines via ArgoCD resource hooks

## Purpose

These specifications serve as:

1. **Design Documentation** - Detailed explanation of architectural decisions and rationale
2. **Implementation Guides** - Reference for developers implementing features
3. **Historical Record** - Documentation of why certain design choices were made
4. **Community Input** - Proposals for feedback and discussion

## Related Documentation

- [Implementation Documentation](../implementation/) - Developer docs and testing information
- [User Documentation](../user/) - User-facing guides and how-tos
