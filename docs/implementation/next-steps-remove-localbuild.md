# Next Steps Required to Remove the Localbuild Controller

## Executive Summary

Based on the controller-architecture-spec.md, the goal is to fully transition from the Localbuild controller (v1alpha1) to the new controller-based architecture (v1alpha2). This document identifies the **concrete next steps** required to complete this migration and remove the Localbuild controller.

## Current State Analysis

### What's Already Implemented ✅

**Phase 1.1: GiteaProvider** (Complete)
- ✅ GiteaProvider CRD (`api/v1alpha2/giteaprovider_types.go`)
- ✅ GiteaProviderReconciler (`pkg/controllers/gitprovider/giteaprovider_controller.go`)
- ✅ Git provider duck-typing utilities (`pkg/util/provider/git.go`)
- ✅ Controller registered in `pkg/controllers/run.go`

**Phase 1.2: NginxGateway** (Complete)
- ✅ NginxGateway CRD (`api/v1alpha2/nginxgateway_types.go`)
- ✅ NginxGatewayReconciler (`pkg/controllers/gatewayprovider/nginxgateway_controller.go`)
- ✅ Gateway provider duck-typing utilities (`pkg/util/provider/gateway.go`)
- ✅ Controller registered in `pkg/controllers/run.go`

**Phase 1.2: Platform Controller** (Partial)
- ✅ Platform CRD (`api/v1alpha2/platform_types.go`)
- ✅ PlatformReconciler basic implementation (`pkg/controllers/platform/platform_controller.go`)
- ✅ Aggregates Git and Gateway provider status
- ❌ **MISSING**: Owner reference pattern implementation
- ❌ **MISSING**: Configuration discovery pattern
- ❌ **MISSING**: Bootstrap repository creation

**Phase 1.3: ArgoCDProvider** (Complete)
- ✅ ArgoCDProvider CRD (`api/v1alpha2/argocdprovider_types.go`)
- ✅ ArgoCDProviderReconciler (`pkg/controllers/gitopsprovider/argocdprovider_controller.go`)
- ✅ GitOps provider duck-typing utilities (`pkg/util/provider/gitops.go`)
- ✅ Controller registered (needs verification)

**CLI Integration** (Partial)
- ✅ CLI creates GiteaProvider CR (`pkg/build/build.go::createGiteaProvider()`)
- ✅ CLI creates Platform CR (`pkg/build/build.go::createPlatform()`)
- ❌ **STILL CREATES**: Localbuild CR (line 281-318 in build.go)
- ❌ **MISSING**: Creation of NginxGateway CR
- ❌ **MISSING**: Creation of ArgoCDProvider CR
- ❌ **MISSING**: Platform references to Gateway and GitOps providers

## Critical Missing Pieces

### 1. Owner Reference Pattern ⚠️ **PRIORITY 1**

According to the spec (lines 344-643), providers should wait for the Platform to establish ownership before beginning reconciliation. This is **NOT IMPLEMENTED**.

**Required Changes:**

#### Platform Controller (`pkg/controllers/platform/platform_controller.go`)
```go
// Add after line 75 (after setting Pending phase)
// Establish owner references to all provider CRs
if err := r.ensureProviderOwnerReferences(ctx, platform); err != nil {
    logger.Error(err, "Failed to ensure provider owner references")
    return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
}

// Add new method
func (r *PlatformReconciler) ensureProviderOwnerReferences(ctx context.Context, platform *v1alpha2.Platform) error {
    // Process all Git providers
    for _, providerRef := range platform.Spec.Components.GitProviders {
        if err := r.ensureOwnerReference(ctx, platform, providerRef); err != nil {
            return err
        }
    }
    
    // Process all Gateway providers
    for _, providerRef := range platform.Spec.Components.Gateways {
        if err := r.ensureOwnerReference(ctx, platform, providerRef); err != nil {
            return err
        }
    }
    
    // Process all GitOps providers
    for _, providerRef := range platform.Spec.Components.GitOpsProviders {
        if err := r.ensureOwnerReference(ctx, platform, providerRef); err != nil {
            return err
        }
    }
    
    return nil
}

func (r *PlatformReconciler) ensureOwnerReference(ctx context.Context, platform *v1alpha2.Platform, providerRef v1alpha2.ProviderReference) error {
    // Implementation from spec lines 563-611
    // Fetch provider as unstructured, add Platform as owner reference
}
```

**Required RBAC Addition:**
```go
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders;nginxgateways;argocdproviders,verbs=update;patch
```

#### GiteaProvider Controller (`pkg/controllers/gitprovider/giteaprovider_controller.go`)
```go
// Add at beginning of Reconcile() method after fetching resource
// Check for Platform owner reference
platformRef := getPlatformOwnerReference(giteaProvider)
if platformRef == nil {
    meta.SetStatusCondition(&giteaProvider.Status.Conditions, metav1.Condition{
        Type:    "Ready",
        Status:  metav1.ConditionFalse,
        Reason:  "WaitingForPlatform",
        Message: "Waiting for Platform resource to add owner reference",
    })
    giteaProvider.Status.Phase = "WaitingForPlatform"
    r.Status().Update(ctx, giteaProvider)
    return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// Discover configuration from Platform
platform := &v1alpha2.Platform{}
if err := r.Get(ctx, types.NamespacedName{
    Name:      platformRef.Name,
    Namespace: giteaProvider.Namespace,
}, platform); err != nil {
    return ctrl.Result{}, err
}

// Apply discovered configuration (e.g., protocol from TLS settings)
// Implementation from spec lines 436-470
```

#### NginxGateway Controller (`pkg/controllers/gatewayprovider/nginxgateway_controller.go`)
```go
// Same pattern as GiteaProvider - wait for owner reference and discover config
```

#### ArgoCDProvider Controller (`pkg/controllers/gitopsprovider/argocdprovider_controller.go`)
```go
// Same pattern as GiteaProvider - wait for owner reference and discover config
```

### 2. Bootstrap Repository Creation ⚠️ **PRIORITY 2**

Currently handled by Localbuild controller (lines 280-415 in `localbuild/controller.go`). Must be moved to Platform controller.

**Required Changes:**

#### Platform Controller
```go
// After all providers are ready, create bootstrap repositories
func (r *PlatformReconciler) createBootstrapRepositories(ctx context.Context, platform *v1alpha2.Platform) error {
    // Get first git provider
    if len(platform.Spec.Components.GitProviders) == 0 {
        return fmt.Errorf("no git provider configured")
    }
    
    // Create GitRepository CRs for each bootstrap app
    bootStrapApps := []string{"argocd", "gitea", "nginx"}
    for _, appName := range bootStrapApps {
        repo := &v1alpha1.GitRepository{
            ObjectMeta: metav1.ObjectMeta{
                Name:      appName,
                Namespace: platform.Namespace,
            },
        }
        
        _, err := controllerutil.CreateOrUpdate(ctx, r.Client, repo, func() error {
            // Set Platform as owner
            if err := controllerutil.SetControllerReference(platform, repo, r.Scheme); err != nil {
                return err
            }
            
            // Configure repo spec using duck-typed git provider
            // Implementation from localbuild/controller.go::reconcileGitRepo()
            return nil
        })
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

#### Move Logic From Localbuild
```go
// Move these functions from localbuild/controller.go to platform/platform_controller.go:
// - reconcileGitRepo() - lines 739-849
// - reconcileEmbeddedApp() - lines 362-415
// - ReconcileArgoAppsWithGitea() - lines 280-360
```

### 3. Update CLI to Create All Provider CRs ⚠️ **PRIORITY 3**

Currently the CLI only creates GiteaProvider and Platform. It needs to create all provider CRs.

**Required Changes to `pkg/build/build.go`:**

```go
// Add creation of NginxGateway CR
func (b *Build) createNginxGateway(ctx context.Context, kubeClient client.Client) error {
    nginxGateway := &v1alpha2.NginxGateway{
        ObjectMeta: metav1.ObjectMeta{
            Name:      b.name + "-nginx",
            Namespace: "ingress-nginx",
        },
    }
    
    _, err := controllerutil.CreateOrUpdate(ctx, kubeClient, nginxGateway, func() error {
        nginxGateway.Spec = v1alpha2.NginxGatewaySpec{
            Namespace: "ingress-nginx",
            Version:   "1.13.0",
            // Configuration from b.cfg
        }
        return nil
    })
    
    return err
}

// Add creation of ArgoCDProvider CR
func (b *Build) createArgoCDProvider(ctx context.Context, kubeClient client.Client) error {
    argocdProvider := &v1alpha2.ArgoCDProvider{
        ObjectMeta: metav1.ObjectMeta{
            Name:      b.name + "-argocd",
            Namespace: "argocd",
        },
    }
    
    _, err := controllerutil.CreateOrUpdate(ctx, kubeClient, argocdProvider, func() error {
        argocdProvider.Spec = v1alpha2.ArgoCDProviderSpec{
            Namespace: "argocd",
            Version:   "v2.12.0",
            AdminCredentials: v1alpha2.ArgoCDAdminCredentials{
                AutoGenerate: true,
            },
        }
        return nil
    })
    
    return err
}

// Update createPlatform() to reference all providers
func (b *Build) createPlatform(ctx context.Context, kubeClient client.Client) error {
    platform := &v1alpha2.Platform{
        ObjectMeta: metav1.ObjectMeta{
            Name:      b.name + "-platform",
            Namespace: "default",
        },
    }
    
    _, err := controllerutil.CreateOrUpdate(ctx, kubeClient, platform, func() error {
        platform.Spec = v1alpha2.PlatformSpec{
            Domain: b.cfg.Host,
            Components: v1alpha2.PlatformComponents{
                GitProviders: []v1alpha2.ProviderReference{
                    {
                        Name:      b.name + "-gitea",
                        Kind:      "GiteaProvider",
                        Namespace: util.GiteaNamespace,
                    },
                },
                Gateways: []v1alpha2.ProviderReference{
                    {
                        Name:      b.name + "-nginx",
                        Kind:      "NginxGateway",
                        Namespace: "ingress-nginx",
                    },
                },
                GitOpsProviders: []v1alpha2.ProviderReference{
                    {
                        Name:      b.name + "-argocd",
                        Kind:      "ArgoCDProvider",
                        Namespace: "argocd",
                    },
                },
            },
        }
        return nil
    })
    
    return err
}

// In Run() method, add calls before creating Platform:
// Line 322-327, add:
setupLog.V(1).Info("Creating nginxgateway resource")
if err := b.createNginxGateway(ctx, kubeClient); err != nil {
    return fmt.Errorf("creating nginxgateway resource: %w", err)
}

setupLog.V(1).Info("Creating argocdprovider resource")
if err := b.createArgoCDProvider(ctx, kubeClient); err != nil {
    return fmt.Errorf("creating argocdprovider resource: %w", err)
}
```

### 4. Remove Localbuild CR Creation ⚠️ **PRIORITY 4**

After the above changes are complete and tested, remove Localbuild CR creation.

**Required Changes to `pkg/build/build.go`:**

```go
// DELETE lines 281-318 (Localbuild CR creation)
// This entire block:
localBuild := v1alpha1.Localbuild{
    ObjectMeta: metav1.ObjectMeta{
        Name: b.name,
    },
}
// ... through ...
if err != nil {
    if b.statusReporter != nil {
        b.statusReporter.FailStep("resources", err)
    }
    return fmt.Errorf("creating localbuild resource: %w", err)
}
```

### 5. Deprecate and Remove Localbuild Controller ⚠️ **PRIORITY 5**

After the CLI no longer creates Localbuild CRs, remove the controller.

**Files to Delete:**
```
pkg/controllers/localbuild/
├── controller.go
├── argo.go
├── argo_test.go
├── gitea.go
├── gitea_test.go
├── installer.go
├── nginxgateway.go
└── resources/
```

**Files to Modify:**
```go
// pkg/controllers/run.go
// DELETE LocalbuildReconciler registration (lines ~50-70)

// api/v1alpha1/localbuild_types.go
// ADD deprecation notice in comments
// Consider keeping for migration period with conversion webhook
```

### 6. Custom Package Handling Migration ⚠️ **PRIORITY 6**

Move custom package reconciliation from Localbuild controller to Platform controller or keep in separate CustomPackage controller.

**Required Changes:**

Either:

**Option A: Move to Platform Controller**
```go
// Move these methods from localbuild/controller.go to platform/platform_controller.go:
// - reconcileCustomPkg() - lines 514-629
// - reconcileCustomPkgUrl() - lines 631-672
// - reconcileCustomPkgDir() - lines 674-702
// - reconcileCustomPkgFile() - lines 704-737
```

**Option B: Enhance Existing CustomPackage Controller**
```go
// The CustomPackage controller already exists
// Just ensure it works with duck-typed providers
// Update it to discover git provider from Platform
```

### 7. Update Platform Controller to Aggregate GitOps Providers

**Required Changes to `pkg/controllers/platform/platform_controller.go`:**

```go
// After line 100 (after aggregating gateways), add:

// Aggregate GitOps Providers  
gitopsStatuses, gitopsReady, err := r.aggregateGitOpsProviders(ctx, platform)
if err != nil {
    logger.Error(err, "Failed to aggregate gitops providers")
    return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
}
platform.Status.Providers.GitOpsProviders = gitopsStatuses
if !gitopsReady {
    allReady = false
}

// Add new method (similar to aggregateGitProviders and aggregateGateways):
func (r *PlatformReconciler) aggregateGitOpsProviders(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
    // Implementation similar to aggregateGitProviders
    // Use provider.IsGitOpsProviderReady() for duck-typing
}
```

**Required RBAC Addition:**
```go
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders,verbs=get;list;watch
```

## Implementation Order

To minimize risk and ensure each step can be validated independently:

### Step 1: Owner Reference Pattern (Week 1)
1. Implement `ensureOwnerReference()` in Platform controller
2. Update GiteaProvider controller to wait for owner reference
3. Update NginxGateway controller to wait for owner reference  
4. Update ArgoCDProvider controller to wait for owner reference
5. Add configuration discovery from Platform
6. Test with example CRs

**Validation:**
- Create providers first, verify they wait in "WaitingForPlatform" phase
- Create Platform, verify it adds owner references
- Verify providers transition to reconciling

### Step 2: Platform Aggregates GitOps (Week 1)
1. Add `aggregateGitOpsProviders()` to Platform controller
2. Update Platform status to include GitOps providers
3. Test Platform with all three provider types

**Validation:**
- Platform.Status.Providers shows all three types
- Platform is Ready only when all providers are Ready

### Step 3: CLI Creates All Providers (Week 2)
1. Add `createNginxGateway()` to build.go
2. Add `createArgoCDProvider()` to build.go
3. Update `createPlatform()` to reference all providers
4. Keep Localbuild CR creation for now (compatibility)

**Validation:**
- Run `idpbuilder create`
- Verify all provider CRs are created
- Verify Platform CR references all providers
- Verify old Localbuild path still works

### Step 4: Bootstrap Repositories (Week 2-3)
1. Move GitRepository creation from Localbuild to Platform controller
2. Move ArgoCD Application creation to Platform controller
3. Test GitOps bootstrap flow with duck-typed providers

**Validation:**
- GitRepository CRs created by Platform
- ArgoCD Applications created by Platform
- Bootstrap apps sync successfully

### Step 5: Custom Packages (Week 3)
1. Decide Option A or B for custom package handling
2. Implement chosen approach
3. Test with custom packages

**Validation:**
- Custom packages work with v1alpha2 architecture
- Package priority handling works

### Step 6: Remove Localbuild (Week 4)
1. Remove Localbuild CR creation from build.go
2. Test thoroughly with v1alpha2 only
3. Add deprecation warning to Localbuild controller
4. After migration period, delete Localbuild controller

**Validation:**
- CLI works without Localbuild CR
- All features work with v1alpha2 architecture
- Integration tests pass

## Testing Strategy

### Unit Tests
- [ ] Test `ensureOwnerReference()` in Platform controller
- [ ] Test provider wait-for-owner-reference logic
- [ ] Test configuration discovery from Platform
- [ ] Test bootstrap repository creation
- [ ] Test Platform aggregation with all provider types

### Integration Tests
- [ ] Test full workflow: CLI → Providers → Platform → Bootstrap
- [ ] Test provider independence (can create in any order)
- [ ] Test Platform handles provider failures gracefully
- [ ] Test upgrade path (v1alpha1 to v1alpha2)

### E2E Tests
- [ ] Run `idpbuilder create` and verify all components work
- [ ] Test with custom packages
- [ ] Test GitOps workflow (no CLI after initial setup)
- [ ] Performance comparison with old architecture

## Success Criteria

- [ ] Platform controller implements owner reference pattern
- [ ] All providers wait for Platform before reconciling
- [ ] Platform creates bootstrap repositories
- [ ] CLI creates Platform and all provider CRs
- [ ] CLI does NOT create Localbuild CR
- [ ] Localbuild controller is removed
- [ ] All integration tests pass
- [ ] Feature parity with v1alpha1 architecture
- [ ] Documentation updated
- [ ] Migration guide complete

## Timeline Estimate

- **Week 1**: Owner reference pattern + GitOps aggregation
- **Week 2**: CLI updates + bootstrap repositories  
- **Week 3**: Custom packages + testing
- **Week 4**: Remove Localbuild + final validation

**Total**: ~4 weeks for complete migration

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| Breaking existing users | High | Keep Localbuild controller until v1alpha2 is proven stable |
| Performance regression | Medium | Benchmark and compare with v1alpha1 |
| Missing functionality | High | Comprehensive feature parity checklist |
| Complex migration | Medium | Detailed step-by-step guide with rollback plan |

## References

- Controller Architecture Spec: `docs/specs/controller-architecture-spec.md`
- Phase 1.2 Status: `docs/implementation/phase-1-2-final-status.md`
- Example CRs: `examples/v1alpha2/`
