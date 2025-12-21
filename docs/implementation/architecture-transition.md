# Architecture Transition: v1alpha1 (Localbuild) → v1alpha2 (Platform-Based)

## Visual Overview

### Current State (v1alpha1 with Localbuild)

```mermaid
graph TB
    CLI["<b>CLI (build.go)</b><br/>- Creates Kind cluster<br/>- Deploys controllers<br/>- Creates Localbuild CR ⚠️ (TO BE REMOVED)"]
    LB["<b>LocalbuildReconciler (DEPRECATED)</b><br/>❌ Installs ArgoCD directly (embedded manifests)<br/>❌ Installs Gitea directly (embedded manifests)<br/>❌ Installs Nginx directly (embedded manifests)<br/>❌ Creates GitRepository CRs<br/>❌ Creates ArgoCD Applications<br/>❌ Handles custom packages"]
    
    CLI --> LB
    
    style CLI fill:#ffe6e6,stroke:#cc0000
    style LB fill:#fff0e6,stroke:#ff9900
```

### Target State (v1alpha2 with Platform + Providers)

```mermaid
graph TB
    CLI["<b>CLI (build.go)</b><br/>- Creates Kind cluster<br/>- Deploys controllers<br/>✅ Creates GiteaProvider CR<br/>🔲 Creates NginxGateway CR (NEEDS IMPL)<br/>🔲 Creates ArgoCDProvider CR (NEEDS IMPL)<br/>✅ Creates Platform CR (partial)"]
    
    Platform["<b>PlatformReconciler</b><br/>🔲 Establishes owner references (NEEDS IMPL)<br/>✅ Aggregates Git provider status<br/>✅ Aggregates Gateway provider status<br/>🔲 Aggregates GitOps provider status (NEEDS IMPL)<br/>🔲 Creates bootstrap GitRepository CRs (NEEDS IMPL)<br/>🔲 Creates ArgoCD Applications (NEEDS IMPL)<br/>✅ Updates Platform.Status"]
    
    GiteaProvider["<b>GiteaProvider CR</b><br/>✅ CRD exists<br/>✅ Controller OK<br/>🔲 Waits for owner ref (NEEDS IMPL)<br/>🔲 Discovers config from Platform (NEEDS IMPL)"]
    
    NginxGateway["<b>NginxGateway CR</b><br/>✅ CRD exists<br/>✅ Controller OK<br/>🔲 Waits for owner ref (NEEDS IMPL)<br/>🔲 Discovers config from Platform (NEEDS IMPL)"]
    
    ArgoCDProvider["<b>ArgoCDProvider CR</b><br/>✅ CRD exists<br/>✅ Controller OK<br/>🔲 Waits for owner ref (NEEDS IMPL)<br/>🔲 Discovers config from Platform (NEEDS IMPL)"]
    
    Gitea["Gitea<br/>Pods"]
    Nginx["Nginx<br/>Ingress Pods"]
    ArgoCD["ArgoCD<br/>Pods"]
    
    CLI --> Platform
    Platform -->|adds owner references| GiteaProvider
    Platform -->|adds owner references| NginxGateway
    Platform -->|adds owner references| ArgoCDProvider
    GiteaProvider -->|installs| Gitea
    NginxGateway -->|installs| Nginx
    ArgoCDProvider -->|installs| ArgoCD
    
    style CLI fill:#e1f5ff,stroke:#01579b
    style Platform fill:#fff9c4,stroke:#f57f17
    style GiteaProvider fill:#f3e5f5,stroke:#7b1fa2
    style NginxGateway fill:#fce4ec,stroke:#c2185b
    style ArgoCDProvider fill:#e0f2f1,stroke:#00695c
    style Gitea fill:#e8f5e9,stroke:#2e7d32
    style Nginx fill:#e8f5e9,stroke:#2e7d32
    style ArgoCD fill:#e8f5e9,stroke:#2e7d32
```

## Key Differences

### 1. **Localbuild Controller** (Old - To Be Removed)
- **Single monolithic controller** doing everything
- **Tight coupling** between components
- **Embedded installation logic** in Go code
- **Hard to customize** without recompiling
- **Sequential installation** (one after another)

### 2. **Platform + Provider Architecture** (New - Target)
- **Separation of concerns** (Platform orchestrates, Providers install)
- **Duck-typing** enables provider independence
- **Declarative CRs** for configuration
- **Easy to customize** via YAML
- **Parallel installation** where possible
- **Extensible** (can add new provider types)

## Migration Checklist

### Phase A: Owner Reference Pattern ⚠️ **CRITICAL PATH**

This is the foundation of the new architecture (spec lines 344-643).

```
Legend:
  ✅ = Complete
  🔲 = Not started
  🚧 = In progress
  ❌ = Blocked
```

**Platform Controller:**
- 🔲 Implement `ensureOwnerReference()` method
- 🔲 Add `ensureProviderOwnerReferences()` to Reconcile loop
- 🔲 Add RBAC for updating provider CRs
- 🔲 Add helper function `getPlatformOwnerReference()`

**GiteaProvider Controller:**
- 🔲 Add owner reference check at start of Reconcile()
- 🔲 Add "WaitingForPlatform" phase
- 🔲 Implement configuration discovery from Platform
- 🔲 Add requeue logic while waiting

**NginxGateway Controller:**
- 🔲 Add owner reference check at start of Reconcile()
- 🔲 Add "WaitingForPlatform" phase
- 🔲 Implement configuration discovery from Platform
- 🔲 Add requeue logic while waiting

**ArgoCDProvider Controller:**
- 🔲 Add owner reference check at start of Reconcile()
- 🔲 Add "WaitingForPlatform" phase
- 🔲 Implement configuration discovery from Platform
- 🔲 Add requeue logic while waiting

**Testing:**
- 🔲 Unit test: ensureOwnerReference()
- 🔲 Unit test: Provider wait-for-owner logic
- 🔲 Integration test: Platform → Provider lifecycle
- 🔲 E2E test: Full workflow with owner references

### Phase B: Platform Bootstrap Creation

**Platform Controller:**
- 🔲 Add `createBootstrapRepositories()` method
- 🔲 Move `reconcileGitRepo()` from Localbuild controller
- 🔲 Move `reconcileEmbeddedApp()` from Localbuild controller
- 🔲 Update to use duck-typed git provider access
- 🔲 Create ArgoCD Applications for bootstrap apps

**GitRepository Controller:**
- ✅ Already works with duck-typed providers (via localbuild)
- 🔲 Verify compatibility with Platform-created repos

**Testing:**
- 🔲 Unit test: Bootstrap repository creation
- 🔲 Integration test: GitRepository → Gitea sync
- 🔲 E2E test: Bootstrap apps in ArgoCD

### Phase C: CLI Integration

**build.go:**
- ✅ `createGiteaProvider()` - Already implemented
- 🔲 `createNginxGateway()` - Needs implementation
- 🔲 `createArgoCDProvider()` - Needs implementation
- ✅ `createPlatform()` - Exists but needs update
  - 🔲 Add Gateways reference
  - 🔲 Add GitOpsProviders reference

**Call sequence in Run():**
- Line 322: ✅ `createGiteaProvider()`
- Line 327: 🔲 `createNginxGateway()` - ADD THIS
- Line XXX: 🔲 `createArgoCDProvider()` - ADD THIS
- Line 330: ✅ `createPlatform()` - UPDATE THIS

**Testing:**
- 🔲 E2E test: `idpbuilder create` with v1alpha2
- 🔲 Verify all provider CRs created
- 🔲 Verify Platform references all providers

### Phase D: GitOps Provider Aggregation

**Platform Controller:**
- 🔲 Add `aggregateGitOpsProviders()` method
- 🔲 Call in Reconcile() after gateway aggregation
- 🔲 Update allReady logic to include gitops
- 🔲 Add RBAC for reading ArgoCDProvider

**Testing:**
- 🔲 Unit test: aggregateGitOpsProviders()
- 🔲 Verify Platform.Status includes GitOps providers
- 🔲 Verify Platform Ready condition

### Phase E: Custom Package Migration

**Decision Point:**
- 🔲 Option A: Move to Platform controller
- 🔲 Option B: Enhance existing CustomPackage controller

**Implementation (whichever chosen):**
- 🔲 Update to use duck-typed providers
- 🔲 Maintain priority handling
- 🔲 Support dirs, files, URLs

**Testing:**
- 🔲 Test with custom package directory
- 🔲 Test with custom package file
- 🔲 Test with custom package URL
- 🔲 Test package priority conflicts

### Phase F: Remove Localbuild

**build.go:**
- 🔲 Remove Localbuild CR creation (lines 281-318)
- 🔲 Update `isCompatible()` to use Platform CR
- 🔲 Remove references to `v1alpha1.Localbuild`

**controllers/run.go:**
- 🔲 Remove LocalbuildReconciler registration
- 🔲 Update controller setup comments

**Deletion:**
- 🔲 Delete `pkg/controllers/localbuild/` directory
- 🔲 Add deprecation notice to v1alpha1 CRD

**Documentation:**
- 🔲 Update user guide to use v1alpha2
- 🔲 Create migration guide from v1alpha1
- 🔲 Update API reference documentation

## Critical Path Analysis

```mermaid
graph TB
    PhaseA["<b>Phase A:</b><br/>Owner Reference Pattern"]
    PhaseB["<b>Phase B:</b><br/>Platform Bootstrap"]
    PhaseC["<b>Phase C:</b><br/>CLI Integration"]
    PhaseD["<b>Phase D:</b><br/>GitOps Aggregation"]
    PhaseE["<b>Phase E:</b><br/>Custom Packages"]
    PhaseF["<b>Phase F:</b><br/>Remove Localbuild"]
    
    PhaseA --> PhaseB
    PhaseA --> PhaseC
    PhaseB --> PhaseD
    PhaseC --> PhaseD
    PhaseD --> PhaseE
    PhaseE --> PhaseF
    
    style PhaseA fill:#ff9999,stroke:#cc0000,stroke-width:3px
    style PhaseB fill:#ffcc99,stroke:#ff9900
    style PhaseC fill:#ffcc99,stroke:#ff9900
    style PhaseD fill:#ffff99,stroke:#cccc00
    style PhaseE fill:#ccffcc,stroke:#00cc00
    style PhaseF fill:#99ccff,stroke:#0066cc
```

**Critical dependencies:**
1. **Owner Reference Pattern** must be completed first
   - Everything else depends on this
   - Providers won't work correctly without it
   
2. **CLI Integration** can proceed in parallel with Platform Bootstrap
   - Both depend on Owner Reference Pattern
   - Can be developed/tested independently
   
3. **Custom Packages** should wait until others are stable
   - Less critical path
   - More complex migration decision

4. **Remove Localbuild** is the final step
   - Only after everything else is complete and tested
   - Requires thorough validation

## Estimated Timeline

```mermaid
gantt
    title 4-Week Implementation Timeline
    dateFormat  YYYY-MM-DD
    section Week 1
    Owner Reference Pattern           :active, w1a, 2025-01-06, 2d
    Update Provider Controllers       :active, w1b, after w1a, 2d
    GitOps Aggregation               :active, w1c, after w1b, 1d
    section Week 2
    CLI Integration                   :w2a, 2025-01-13, 2d
    Platform Bootstrap                :w2b, after w2a, 3d
    section Week 3
    Custom Packages                   :w3a, 2025-01-20, 2d
    Integration Testing               :w3b, after w3a, 2d
    Performance Benchmarking          :w3c, after w3b, 1d
    section Week 4
    Remove Localbuild CR              :w4a, 2025-01-27, 2d
    Remove Localbuild Controller      :w4b, after w4a, 2d
    Documentation & Final Review      :w4c, after w4b, 1d
```

## Success Metrics

**Before (v1alpha1):**
- Time to cluster ready: ~2-3 minutes
- Components: 1 CR (Localbuild)
- Controllers: 1 (LocalbuildReconciler)
- Flexibility: Low (embedded logic)
- Customization: Requires recompilation

**After (v1alpha2):**
- Time to cluster ready: ~2-3 minutes (same or better)
- Components: 4 CRs (Platform + 3 Providers)
- Controllers: 4 (Platform + 3 Provider reconcilers)
- Flexibility: High (declarative CRs)
- Customization: YAML-based

## Validation Checklist

Before removing Localbuild controller, verify:

- [ ] `idpbuilder create` works without Localbuild CR
- [ ] All provider CRs created successfully
- [ ] Platform CR references all providers
- [ ] Owner references established correctly
- [ ] Providers wait for Platform before reconciling
- [ ] Configuration discovered from Platform
- [ ] GitRepository CRs created by Platform
- [ ] ArgoCD Applications created for bootstrap
- [ ] Custom packages work
- [ ] All integration tests pass
- [ ] Performance is comparable or better
- [ ] Documentation updated
- [ ] Migration guide complete

## Rollback Plan

If issues are discovered after Localbuild removal:

1. **Revert CLI changes** - Restore Localbuild CR creation
2. **Re-register controller** - Add LocalbuildReconciler back to run.go
3. **Restore controller code** - Undelete localbuild directory from git
4. **Document issues** - Create detailed issue report
5. **Fix and retry** - Address root cause before attempting again

## References

- **Spec:** `docs/specs/controller-architecture-spec.md`
- **Implementation Guide:** `docs/implementation/next-steps-remove-localbuild.md`
- **Phase 1.2 Status:** `docs/implementation/phase-1-2-final-status.md`
- **Examples:** `examples/v1alpha2/`
