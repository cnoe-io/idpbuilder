# Examples

This directory contains example YAML manifests for the Phase 1.1 controller-based architecture.

## Platform CR Examples

### Simple Platform
[`platform-simple.yaml`](./platform-simple.yaml) - A basic Platform CR that references a GiteaProvider.

### Complete Platform
[`platform-complete.yaml`](./platform-complete.yaml) - A complete example with both Platform and GiteaProvider CRs in a single file.

## GiteaProvider Examples

### Simple GiteaProvider
[`giteaprovider-simple.yaml`](./giteaprovider-simple.yaml) - A basic GiteaProvider CR with auto-generated credentials and organizations.

## Usage

1. First, ensure you have a Kubernetes cluster running:
```bash
kind create cluster
```

2. Apply the CRD manifests:
```bash
kubectl apply -f pkg/controllers/resources/idpbuilder.cnoe.io_giteaproviders.yaml
kubectl apply -f pkg/controllers/resources/idpbuilder.cnoe.io_platforms.yaml
```

3. Create the namespace for Gitea:
```bash
kubectl create namespace gitea
```

4. Apply the complete example:
```bash
kubectl apply -f examples/platform-complete.yaml
```

5. Check the status:
```bash
# Check GiteaProvider status
kubectl get giteaprovider -n gitea

# Check Platform status
kubectl get platform

# Get detailed status
kubectl describe platform my-platform
```

## Status Fields

### GiteaProvider Status

The GiteaProvider exposes the following duck-typed status fields that the Platform controller uses:

- `endpoint` - External URL for Gitea web UI and cloning
- `internalEndpoint` - Cluster-internal URL for API access
- `credentialsSecretRef` - Reference to the secret containing admin credentials
- `conditions` - Kubernetes-style conditions, including a "Ready" condition

### Platform Status

The Platform aggregates status from all referenced providers:

- `providers.gitProviders[]` - Summary of all Git provider statuses
- `conditions` - Overall platform conditions
- `phase` - Current phase (Pending, Ready, Failed)
