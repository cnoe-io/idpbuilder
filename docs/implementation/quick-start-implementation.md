# Quick Start: Implementing the Next Steps

This is a practical guide for developers implementing the next steps to remove the Localbuild controller. Each section provides concrete code locations and changes.

## Prerequisites

- Familiarity with controller-runtime patterns
- Understanding of Kubernetes custom resources
- Go 1.21+
- Access to the repository

## Step 1: Owner Reference Pattern (Priority 1)

### 1.1 Update Platform Controller

**File:** `pkg/controllers/platform/platform_controller.go`

**Location:** After line 75 (after setting Pending phase)

```go
// Add this code block
// Establish owner references to all provider CRs
if err := r.ensureProviderOwnerReferences(ctx, platform); err != nil {
    logger.Error(err, "Failed to ensure provider owner references")
    return ctrl.Result{RequeueAfter: defaultRequeueTime}, nil
}
```

**Location:** End of file (new methods)

```go
// ensureProviderOwnerReferences ensures Platform is set as owner on all referenced providers
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

// ensureOwnerReference adds Platform as owner reference to a provider CR
func (r *PlatformReconciler) ensureOwnerReference(ctx context.Context, platform *v1alpha2.Platform, providerRef v1alpha2.ProviderReference) error {
    logger := log.FromContext(ctx)
    
    // Fetch provider as unstructured (works for any provider type)
    provider := &unstructured.Unstructured{}
    provider.SetGroupVersionKind(schema.GroupVersionKind{
        Group:   "idpbuilder.cnoe.io",
        Version: "v1alpha2",
        Kind:    providerRef.Kind,
    })
    
    key := types.NamespacedName{
        Name:      providerRef.Name,
        Namespace: providerRef.Namespace,
    }
    
    if err := r.Get(ctx, key, provider); err != nil {
        if errors.IsNotFound(err) {
            logger.Info("Provider not found yet, will retry", "name", providerRef.Name, "kind", providerRef.Kind)
            return nil // Don't error, provider might be created soon
        }
        return err
    }
    
    // Check if Platform is already an owner
    hasOwnerRef := false
    for _, ref := range provider.GetOwnerReferences() {
        if ref.UID == platform.UID {
            hasOwnerRef = true
            break
        }
    }
    
    if !hasOwnerRef {
        // Add Platform as owner reference
        ownerRef := metav1.OwnerReference{
            APIVersion: platform.APIVersion,
            Kind:       platform.Kind,
            Name:       platform.Name,
            UID:        platform.UID,
            Controller: func() *bool { b := false; return &b }(), // Not a controller owner
        }
        
        refs := provider.GetOwnerReferences()
        refs = append(refs, ownerRef)
        provider.SetOwnerReferences(refs)
        
        logger.Info("Adding Platform as owner reference", "provider", providerRef.Name, "kind", providerRef.Kind)
        if err := r.Update(ctx, provider); err != nil {
            return err
        }
    }
    
    return nil
}
```

**Update RBAC markers** at the top of file (after existing rbac markers):

```go
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=giteaproviders;nginxgateways;argocdproviders,verbs=update;patch
```

### 1.2 Update GiteaProvider Controller

**File:** `pkg/controllers/gitprovider/giteaprovider_controller.go`

**Location:** At the beginning of `Reconcile()` method, after fetching the resource (around line 50)

```go
// Add this helper function at the end of file first
func getPlatformOwnerReference(obj metav1.Object) *metav1.OwnerReference {
    for i := range obj.GetOwnerReferences() {
        ref := &obj.GetOwnerReferences()[i]
        if ref.APIVersion == "idpbuilder.cnoe.io/v1alpha2" && ref.Kind == "Platform" {
            return ref
        }
    }
    return nil
}

// Then add this at the beginning of Reconcile(), after fetching giteaProvider
platformRef := getPlatformOwnerReference(giteaProvider)
if platformRef == nil {
    logger.Info("Waiting for Platform to add owner reference")
    meta.SetStatusCondition(&giteaProvider.Status.Conditions, metav1.Condition{
        Type:    "Ready",
        Status:  metav1.ConditionFalse,
        Reason:  "WaitingForPlatform",
        Message: "Waiting for Platform resource to add owner reference",
    })
    giteaProvider.Status.Phase = "WaitingForPlatform"
    if err := r.Status().Update(ctx, giteaProvider); err != nil {
        logger.Error(err, "Failed to update status")
    }
    return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
}

// Get Platform resource for configuration discovery
platform := &v1alpha2.Platform{}
platformKey := types.NamespacedName{
    Name:      platformRef.Name,
    Namespace: giteaProvider.Namespace,
}
if err := r.Get(ctx, platformKey, platform); err != nil {
    logger.Error(err, "Failed to get Platform")
    return ctrl.Result{}, err
}

// Discover configuration from Platform (if not explicitly set)
if giteaProvider.Spec.AdminUser.Email == "" && platform.Spec.Domain != "" {
    giteaProvider.Spec.AdminUser.Email = "admin@" + platform.Spec.Domain
}
```

### 1.3 Update NginxGateway Controller

**File:** `pkg/controllers/gatewayprovider/nginxgateway_controller.go`

**Add the same pattern** as GiteaProvider:
- Add `getPlatformOwnerReference()` helper
- Add owner reference check at start of `Reconcile()`
- Add configuration discovery from Platform

### 1.4 Update ArgoCDProvider Controller

**File:** `pkg/controllers/gitopsprovider/argocdprovider_controller.go`

**Add the same pattern** as GiteaProvider and NginxGateway.

### 1.5 Test Step 1

```bash
# Generate CRDs
make manifests

# Build
make build

# Run unit tests
go test ./pkg/controllers/platform/... -v
go test ./pkg/controllers/gitprovider/... -v
go test ./pkg/controllers/gatewayprovider/... -v
go test ./pkg/controllers/gitopsprovider/... -v

# Test with example CRs (requires cluster)
kubectl apply -f examples/v1alpha2/giteaprovider.yaml
kubectl apply -f examples/v1alpha2/nginxgateway.yaml
kubectl apply -f examples/v1alpha2/argocdprovider.yaml

# Verify they're waiting
kubectl get giteaproviders -A
kubectl get nginxgateways -A
kubectl get argocdproviders -A

# All should show Phase: WaitingForPlatform

# Create Platform
kubectl apply -f examples/v1alpha2/platform-with-all-providers.yaml

# Verify owner references added
kubectl get giteaproviders <name> -n <namespace> -o yaml | grep ownerReferences -A 5

# Verify providers start reconciling
kubectl get giteaproviders,nginxgateways,argocdproviders -A
# Should transition from WaitingForPlatform to Installing/Ready
```

## Step 2: GitOps Provider Aggregation (Priority 2)

### 2.1 Update Platform Controller

**File:** `pkg/controllers/platform/platform_controller.go`

**Location:** After `aggregateGateways()` call (around line 100)

```go
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
```

**Location:** End of file (new method)

```go
// aggregateGitOpsProviders aggregates status from all GitOps providers
func (r *PlatformReconciler) aggregateGitOpsProviders(ctx context.Context, platform *v1alpha2.Platform) ([]v1alpha2.ProviderStatusSummary, bool, error) {
    logger := log.FromContext(ctx)
    summaries := []v1alpha2.ProviderStatusSummary{}
    allReady := true
    
    for _, gitopsProviderRef := range platform.Spec.Components.GitOpsProviders {
        // Fetch provider using unstructured client to support duck-typing
        gvk := schema.GroupVersionKind{
            Group:   "idpbuilder.cnoe.io",
            Version: "v1alpha2",
            Kind:    gitopsProviderRef.Kind,
        }
        
        providerObj := &unstructured.Unstructured{}
        providerObj.SetGroupVersionKind(gvk)
        
        err := r.Get(ctx, types.NamespacedName{
            Name:      gitopsProviderRef.Name,
            Namespace: gitopsProviderRef.Namespace,
        }, providerObj)
        
        if err != nil {
            if errors.IsNotFound(err) {
                logger.Info("GitOps provider not found", "name", gitopsProviderRef.Name, "kind", gitopsProviderRef.Kind)
                summaries = append(summaries, v1alpha2.ProviderStatusSummary{
                    Name:  gitopsProviderRef.Name,
                    Kind:  gitopsProviderRef.Kind,
                    Ready: false,
                })
                allReady = false
                continue
            }
            return nil, false, fmt.Errorf("getting gitops provider %s: %w", gitopsProviderRef.Name, err)
        }
        
        // Extract status using duck-typing
        ready, err := provider.IsGitOpsProviderReady(providerObj)
        if err != nil {
            logger.Error(err, "Failed to check gitops provider readiness", "name", gitopsProviderRef.Name)
            ready = false
        }
        
        summaries = append(summaries, v1alpha2.ProviderStatusSummary{
            Name:  gitopsProviderRef.Name,
            Kind:  gitopsProviderRef.Kind,
            Ready: ready,
        })
        
        if !ready {
            allReady = false
        }
    }
    
    return summaries, allReady, nil
}
```

**Update RBAC**:

```go
//+kubebuilder:rbac:groups=idpbuilder.cnoe.io,resources=argocdproviders,verbs=get;list;watch
```

### 2.2 Test Step 2

```bash
# Run tests
go test ./pkg/controllers/platform/... -v

# Create Platform with GitOps provider
kubectl apply -f examples/v1alpha2/platform-with-all-providers.yaml

# Check Platform status
kubectl get platform <name> -n <namespace> -o yaml

# Verify gitOpsProviders section in status
# Should show ArgoCDProvider with ready: true/false
```

## Step 3: CLI Creates All Provider CRs (Priority 3)

### 3.1 Add NginxGateway Creation

**File:** `pkg/build/build.go`

**Location:** After `createGiteaProvider()` method (around line 403)

```go
// createNginxGateway creates a NginxGateway CR
func (b *Build) createNginxGateway(ctx context.Context, kubeClient client.Client) error {
    // Ensure ingress-nginx namespace exists
    if err := k8s.EnsureNamespace(ctx, kubeClient, "ingress-nginx"); err != nil {
        return fmt.Errorf("ensuring ingress-nginx namespace: %w", err)
    }
    
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
        }
        return nil
    })
    
    return err
}
```

### 3.2 Add ArgoCDProvider Creation

**File:** `pkg/build/build.go`

**Location:** After `createNginxGateway()` method

```go
// createArgoCDProvider creates an ArgoCDProvider CR
func (b *Build) createArgoCDProvider(ctx context.Context, kubeClient client.Client) error {
    // Ensure argocd namespace exists
    if err := k8s.EnsureNamespace(ctx, kubeClient, "argocd"); err != nil {
        return fmt.Errorf("ensuring argocd namespace: %w", err)
    }
    
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
```

### 3.3 Update Platform Creation

**File:** `pkg/build/build.go`

**Location:** Replace `createPlatform()` method (around line 405)

```go
// createPlatform creates a Platform CR that references all providers
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
```

### 3.4 Update Run() to Call New Methods

**File:** `pkg/build/build.go`

**Location:** In `Run()` method, after `createGiteaProvider()` call (around line 327)

```go
// After this:
setupLog.V(1).Info("Creating giteaprovider resource")
if err := b.createGiteaProvider(ctx, kubeClient); err != nil {
    if b.statusReporter != nil {
        b.statusReporter.FailStep("resources", err)
    }
    return fmt.Errorf("creating giteaprovider resource: %w", err)
}

// Add this:
setupLog.V(1).Info("Creating nginxgateway resource")
if err := b.createNginxGateway(ctx, kubeClient); err != nil {
    if b.statusReporter != nil {
        b.statusReporter.FailStep("resources", err)
    }
    return fmt.Errorf("creating nginxgateway resource: %w", err)
}

setupLog.V(1).Info("Creating argocdprovider resource")
if err := b.createArgoCDProvider(ctx, kubeClient); err != nil {
    if b.statusReporter != nil {
        b.statusReporter.FailStep("resources", err)
    }
    return fmt.Errorf("creating argocdprovider resource: %w", err)
}
```

### 3.5 Test Step 3

```bash
# Build CLI
make build

# Run create
./bin/idpbuilder create --name test

# Verify all CRs created
kubectl get giteaproviders -A
kubectl get nginxgateways -A
kubectl get argocdproviders -A
kubectl get platforms -A

# Check Platform references
kubectl get platform test-platform -n default -o yaml

# Should show all three provider types referenced
```

## Step 4: Remove Localbuild CR Creation (Final Step)

**ONLY DO THIS AFTER STEPS 1-3 ARE COMPLETE AND TESTED**

### 4.1 Remove Localbuild Creation from CLI

**File:** `pkg/build/build.go`

**Location:** In `Run()` method, DELETE lines 281-318

```go
// DELETE THIS ENTIRE BLOCK:
localBuild := v1alpha1.Localbuild{
    ObjectMeta: metav1.ObjectMeta{
        Name: b.name,
    },
}

cliStartTime := time.Now().Format(time.RFC3339Nano)

setupLog.V(1).Info("Creating localbuild resource")
_, err = controllerutil.CreateOrUpdate(ctx, kubeClient, &localBuild, func() error {
    if localBuild.ObjectMeta.Annotations == nil {
        localBuild.ObjectMeta.Annotations = map[string]string{}
    }
    localBuild.ObjectMeta.Annotations[v1alpha1.CliStartTimeAnnotation] = cliStartTime
    localBuild.Spec = v1alpha1.LocalbuildSpec{
        BuildCustomization: b.cfg,
        PackageConfigs: v1alpha1.PackageConfigsSpec{
            Argo: v1alpha1.ArgoPackageConfigSpec{
                Enabled: true,
            },
            EmbeddedArgoApplications: v1alpha1.EmbeddedArgoApplicationsPackageConfigSpec{
                Enabled: true,
            },
            CustomPackageDirs:        b.customPackageDirs,
            CustomPackageFiles:       b.customPackageFiles,
            CustomPackageUrls:        b.customPackageUrls,
            CorePackageCustomization: b.packageCustomization,
        },
    }

    return nil
})
if err != nil {
    if b.statusReporter != nil {
        b.statusReporter.FailStep("resources", err)
    }
    return fmt.Errorf("creating localbuild resource: %w", err)
}
```

### 4.2 Update isCompatible()

**File:** `pkg/build/build.go`

**Location:** Replace `isCompatible()` method

```go
func (b *Build) isCompatible(ctx context.Context, kubeClient client.Client) (bool, error) {
    // Check Platform CR instead of Localbuild
    platform := v1alpha2.Platform{
        ObjectMeta: metav1.ObjectMeta{
            Name:      b.name + "-platform",
            Namespace: "default",
        },
    }
    
    err := kubeClient.Get(ctx, client.ObjectKeyFromObject(&platform), &platform)
    if err != nil {
        if k8serrors.IsNotFound(err) {
            return true, nil
        }
        return false, err
    }
    
    // Check domain compatibility
    if platform.Spec.Domain != b.cfg.Host {
        return false, fmt.Errorf("provided domain and existing configuration are incompatible. "+
            "existing: %s, given: %s", platform.Spec.Domain, b.cfg.Host)
    }
    
    return true, nil
}
```

### 4.3 Test Step 4

```bash
# Build with changes
make build

# Create cluster
./bin/idpbuilder create --name test2

# Verify NO Localbuild CR exists
kubectl get localbuilds
# Should show "No resources found"

# Verify Platform CR exists
kubectl get platforms -A

# Verify all functionality works
kubectl get pods -A
kubectl get applications -n argocd
```

## Troubleshooting Common Issues

### Issue: Providers stuck in "WaitingForPlatform"

**Solution:**
```bash
# Check if Platform exists
kubectl get platforms -A

# Check if Platform references the provider
kubectl get platform <name> -n <namespace> -o yaml | grep -A 10 "gitProviders\|gateways\|gitOpsProviders"

# Check Platform controller logs
kubectl logs -n idpbuilder-system deployment/idpbuilder-controller-manager | grep Platform
```

### Issue: Owner reference not being added

**Solution:**
```bash
# Check Platform controller RBAC
kubectl get clusterrole idpbuilder-controller-manager-role -o yaml | grep -A 5 "giteaproviders\|nginxgateways\|argocdproviders"

# Should have update and patch verbs

# Check Platform controller logs for errors
kubectl logs -n idpbuilder-system deployment/idpbuilder-controller-manager | grep "ensureOwnerReference"
```

### Issue: Build fails after removing Localbuild

**Solution:**
```bash
# Make sure you removed all references
grep -r "Localbuild" pkg/build/

# Check imports
grep -r "v1alpha1.Localbuild" .

# Make sure you didn't break compatibility check
# Test with fresh cluster
```

## Testing Checklist

Before committing each step:

**Step 1 (Owner Reference):**
- [ ] Platform controller compiles
- [ ] Provider controllers compile
- [ ] Unit tests pass
- [ ] Providers wait for Platform
- [ ] Owner references are added
- [ ] Configuration is discovered

**Step 2 (GitOps Aggregation):**
- [ ] Platform status includes GitOps providers
- [ ] Platform Ready condition works
- [ ] Unit tests pass

**Step 3 (CLI Integration):**
- [ ] All provider CRs created by CLI
- [ ] Platform references all providers
- [ ] Integration test passes

**Step 4 (Remove Localbuild):**
- [ ] CLI works without Localbuild
- [ ] All functionality preserved
- [ ] E2E test passes

## Next Steps After Implementation

1. **Documentation:**
   - Update user guide
   - Create migration guide
   - Update API docs

2. **Examples:**
   - Create complete example YAML
   - Add to examples/v1alpha2/

3. **Testing:**
   - Add integration tests
   - Add E2E tests
   - Performance benchmarking

4. **Cleanup:**
   - Remove localbuild controller code
   - Update deprecation notices
   - Archive old examples

## Getting Help

If you get stuck:

1. Check the spec: `docs/specs/controller-architecture-spec.md`
2. Review implementation guide: `docs/implementation/next-steps-remove-localbuild.md`
3. Look at phase 1.2 status: `docs/implementation/phase-1-2-final-status.md`
4. Check examples: `examples/v1alpha2/`
5. Ask in team chat or create an issue

## Command Reference

```bash
# Build
make build

# Generate CRDs
make manifests

# Generate DeepCopy
make generate

# Run tests
make test

# Run specific package tests
go test ./pkg/controllers/platform/... -v

# Run controller locally (requires cluster)
make run

# Deploy to cluster
make deploy

# Create example resources
kubectl apply -f examples/v1alpha2/

# Check controller logs
kubectl logs -n idpbuilder-system deployment/idpbuilder-controller-manager -f

# Delete everything
kubectl delete -f examples/v1alpha2/
```
