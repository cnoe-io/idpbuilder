# V1Alpha2 Examples

This directory contains example configurations for the v1alpha2 controller-based architecture.

## Overview

The v1alpha2 architecture introduces a modular, provider-based system where platform components are managed through separate Custom Resources:

- **Platform**: Orchestrates the overall IDP platform by referencing provider CRs
- **GiteaProvider**: Manages Gitea Git server installation and configuration
- **NginxGateway**: Manages Nginx Ingress Controller installation and configuration

## Quick Start

### 1. Create Provider CRs

First, create the individual provider CRs:

```bash
# Create Gitea provider
kubectl apply -f giteaprovider.yaml

# Create Nginx gateway provider
kubectl apply -f nginxgateway.yaml
```

### 2. Create Platform CR

Then create the Platform CR that references these providers:

```bash
kubectl apply -f platform-with-gateway.yaml
```

### 3. Check Status

Monitor the installation progress:

```bash
# Check Platform status
kubectl get platform -n idpbuilder-system

# Check individual provider statuses
kubectl get giteaprovider -n idpbuilder-system
kubectl get nginxgateway -n idpbuilder-system

# View detailed status
kubectl describe platform localdev -n idpbuilder-system
```

## Architecture

The Platform CR aggregates status from all referenced providers using duck-typing:

```
Platform CR (localdev)
  ├── Git Providers
  │   └── gitea-local (GiteaProvider)
  └── Gateways
      └── nginx-gateway (NginxGateway)
```

Each provider controller:
1. Installs its respective component (Gitea, Nginx, etc.)
2. Waits for the component to be ready
3. Updates its status with duck-typed fields
4. Sets the Ready condition when operational

The Platform controller:
1. Monitors all referenced providers
2. Aggregates their readiness status
3. Updates Platform status to Ready when all providers are ready

## Duck-Typed Status Fields

### Git Providers (e.g., GiteaProvider)
- `endpoint`: External URL for web UI and cloning
- `internalEndpoint`: Cluster-internal URL for API access
- `credentialsSecretRef`: Secret containing access credentials

### Gateway Providers (e.g., NginxGateway)
- `ingressClassName`: Name of the ingress class to use
- `loadBalancerEndpoint`: External endpoint for accessing services
- `internalEndpoint`: Cluster-internal API endpoint

All providers expose a `Ready` condition in their status.

## Files

- **giteaprovider.yaml**: Example GiteaProvider configuration
- **nginxgateway.yaml**: Example NginxGateway configuration
- **platform-with-gateway.yaml**: Example Platform CR referencing both providers

## Migration from V1Alpha1

The v1alpha2 architecture is designed to coexist with v1alpha1. The existing Localbuild CR continues to work, but the new provider-based architecture offers:

- Better separation of concerns
- Easier customization of individual components
- Support for alternative providers (e.g., GitHub, GitLab, Envoy)
- Improved observability through kubectl
- GitOps-friendly declarative management

## Next Steps

After the platform is ready, you can:

1. Access Gitea: Check the `endpoint` field in GiteaProvider status
2. Deploy applications using the Nginx Ingress
3. Add more providers to the Platform CR as they become available
