# Controller-Based Architecture Specification

**Version:** 1.0 Draft  
**Date:** December 19, 2025  
**Status:** Proposal  
**Authors:** IDP Builder Team

## Executive Summary

This document proposes a significant architectural evolution of the idpbuilder tool to transition from a CLI-driven installation model to a controller-based architecture. This change will enable idpbuilder to function as a true Kubernetes-native platform, where infrastructure components and application workloads are managed declaratively through Kubernetes Custom Resources (CRs) and reconciliation loops.

**Key Architectural Change**: The new design introduces a clear separation between the **idpbuilder CLI** and the **idpbuilder controllers**:

- **CLI**: Responsible for local infrastructure provisioning (Kind clusters), deploying controllers, and instantiating CRs for development use cases
- **Controllers**: Run in-cluster, manage provider lifecycle, and handle all reconciliation - can be deployed without CLI in production

This separation enables two deployment modes:
1. **CLI-Driven (Development)**: Simple `idpbuilder create` command for local development
2. **GitOps-Driven (Production)**: Controllers installed via Helm/kubectl, CRs managed via GitOps, no CLI required

### Goals

1. **Kubernetes-Native Management**: Enable all functionality to be managed through kubectl and GitOps tools like ArgoCD
2. **Separation of Concerns**: 
   - **CLI and Controllers**: Clear boundary between infrastructure provisioning (CLI) and platform reconciliation (controllers)
   - **Infrastructure and Services**: Delineate cluster provisioning from application/service management
3. **Production Readiness**: Support production workloads and virtualized control planes (e.g., vCluster, Cluster API)
4. **Deployment Flexibility**: Support both CLI-driven development workflows and GitOps-driven production deployments
5. **Extensibility**: Allow easier integration of additional services and customization by end users
6. **Operational Excellence**: Improve observability, debugging, and lifecycle management through standard Kubernetes patterns
7. **GitOps Native**: Enable deployment and management of controllers without any CLI interaction

### Non-Goals

1. Breaking changes to the CLI experience (backward compatibility maintained where feasible)
2. Removing the ability to run idpbuilder as a single binary for development use cases
3. Supporting non-Kubernetes infrastructure
4. Requiring CLI for production deployments (controllers must be deployable via standard Kubernetes methods)

## Current Architecture

### Overview

Today, idpbuilder operates in two distinct phases:

```
┌──────────────────────────────────────────────────────────────────┐
│                         CLI Phase                                 │
│  1. Parse flags                                                   │
│  2. Create Kind cluster                                           │
│  3. Start controller manager                                      │
│  4. Create Localbuild CR                                          │
│  5. Wait for ready state                                          │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│                    Controller Phase                               │
│  LocalbuildReconciler:                                            │
│    - Installs core packages (nginx, argocd, gitea)               │
│    - Creates GitRepository CRs                                    │
│    - Creates ArgoCD Applications                                  │
│                                                                   │
│  RepositoryReconciler:                                            │
│    - Creates Gitea repositories                                   │
│    - Populates repository content                                 │
│                                                                   │
│  CustomPackageReconciler:                                         │
│    - Processes custom packages                                    │
│    - Creates GitRepository CRs and ArgoCD apps                    │
└──────────────────────────────────────────────────────────────────┘
```

### Core Components Installation Flow

The `LocalbuildReconciler` currently performs the following:

1. **Embeds manifests** at compile time (via `//go:embed`)
2. **Direct installation** of nginx, ArgoCD, and Gitea through the Kubernetes API
3. **Configuration** through inline resource manipulation
4. **Status tracking** through deployment readiness checks
5. **GitOps handoff** by creating GitRepository CRs and ArgoCD Applications

### Key Problems with Current Architecture

1. **Tight Coupling**: Core package installation logic is embedded in the LocalbuildReconciler
2. **Limited Flexibility**: Difficult to customize or replace core components
3. **Infrastructure Blur**: No clear boundary between infrastructure (Kind cluster) and platform services
4. **Debugging Challenges**: Installation failures require binary debugging rather than kubectl inspection
5. **Upgrade Complexity**: Updating core components requires new binary releases
6. **Production Barriers**: Embedded installation approach doesn't align with production deployment patterns

## Proposed Architecture

### CLI and Controller Separation

In the new architecture, there is a clear separation between the **CLI** and the **idpbuilder controllers**:

#### CLI Responsibilities
The idpbuilder CLI serves as a developer-friendly tool for local development scenarios:

1. **Infrastructure Provisioning**: Creates and manages local Kubernetes clusters (Kind, etc.)
2. **Controller Deployment**: Deploys the idpbuilder controller manager to the cluster
3. **CR Instantiation**: Creates idpbuilder Custom Resources (Platform, Providers) for common use cases
4. **Developer Experience**: Provides a simple, opinionated workflow for getting started quickly
5. **Configuration**: Translates CLI flags into appropriate CRs and configurations

#### Controller Responsibilities
The idpbuilder controllers run as a deployment in the Kubernetes cluster and handle:

1. **Platform Orchestration**: Manages the Platform CR and coordinates provider installation
2. **Provider Lifecycle**: Installs, configures, and manages Git, Gateway, and GitOps providers
3. **GitOps Integration**: Creates and manages GitRepository CRs and ArgoCD Applications
4. **Status Management**: Updates status conditions and aggregates component health
5. **Reconciliation**: Continuously ensures desired state matches actual state

#### Two Installation Modes

**Mode 1: CLI-Driven (Development)**
- Use `idpbuilder create` command
- CLI provisions Kind cluster
- CLI deploys idpbuilder controllers
- CLI creates Platform and Provider CRs
- Optimized for quick local development

**Mode 2: GitOps-Driven (Production)**
- Pre-provision Kubernetes cluster (any distribution)
- Install idpbuilder controllers via Helm chart or manifests
- Deploy Platform and Provider CRs via GitOps (ArgoCD, Flux, etc.)
- No CLI required after initial controller installation
- Full declarative management

This separation enables:
- **Flexibility**: Use CLI for development, GitOps for production
- **Portability**: Controllers work on any Kubernetes cluster
- **GitOps Native**: Controllers can be managed declaratively
- **Simplicity**: CLI abstracts complexity for developers
- **Production Ready**: Controllers support enterprise deployment patterns

#### Provider Support

- **Git Provider**: Gitea (in-cluster), GitHub (external), GitLab (external or in-cluster)
- **Gateway Provider**: Nginx Ingress, Envoy Gateway, Istio Gateway

### High-Level Design

The new architecture introduces a composable, provider-based system where platform components are defined as separate Custom Resources with duck-typed interfaces. This enables:
- **Multiple provider implementations** running simultaneously
- **Pluggable Git providers**: Gitea (in-cluster), GitHub (external), GitLab
- **Pluggable Gateway providers**: Nginx Ingress, Envoy Gateway, Istio Gateway
- **GitOps management**: ArgoCD (default), Flux (future support)

Controllers run on the provisioned cluster and manage their respective providers:

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Infrastructure Layer                            │
│                                                                      │
│  Two Deployment Modes:                                              │
│                                                                      │
│  Mode 1 - CLI-Driven (Development):                                 │
│    • idpbuilder CLI provisions Kind cluster                         │
│    • CLI deploys idpbuilder-controllers (Helm/manifests)            │
│    • CLI creates Platform and Provider CRs                          │
│                                                                      │
│  Mode 2 - GitOps-Driven (Production):                               │
│    • Pre-provisioned Kubernetes cluster (any distribution)          │
│    • Install controllers via Helm chart or kubectl apply            │
│    • Deploy Platform/Provider CRs via GitOps (ArgoCD/Flux)          │
│    • No CLI required - fully declarative                            │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                 Platform Controllers (On-Cluster)                    │
│                                                                      │
│  PlatformReconciler:                                                 │
│    - Orchestrates platform bootstrap                                 │
│    - References provider CRs (Git, Gateway, GitOps)                 │
│    - Creates GitRepository CRs for bootstrap content                │
│    - Aggregates component status                                     │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  Git Provider Controllers (Duck-Typed)                         │ │
│  │                                                                │ │
│  │  GiteaProviderReconciler:                                      │ │
│  │    - Installs Gitea via Helm                                   │ │
│  │    - Creates organizations and admin users                     │ │
│  │    - Exposes: endpoint, internalEndpoint, credentialsSecretRef │ │
│  │                                                                │ │
│  │  GitHubProviderReconciler:                                     │ │
│  │    - Validates GitHub credentials and access                   │ │
│  │    - Manages organization/team configuration                   │ │
│  │    - Exposes: endpoint, internalEndpoint, credentialsSecretRef │ │
│  │                                                                │ │
│  │  GitLabProviderReconciler:                                     │ │
│  │    - Validates GitLab credentials and access                   │ │
│  │    - Manages groups and subgroups                              │ │
│  │    - Exposes: endpoint, internalEndpoint, credentialsSecretRef │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  Gateway Provider Controllers (Duck-Typed)                     │ │
│  │                                                                │ │
│  │  NginxGatewayReconciler:                                       │ │
│  │    - Installs Nginx Ingress Controller via Helm               │ │
│  │    - Creates IngressClass resource                             │ │
│  │    - Exposes: ingressClassName, loadBalancerEndpoint           │ │
│  │                                                                │ │
│  │  EnvoyGatewayReconciler:                                       │ │
│  │    - Installs Envoy Gateway via Helm                           │ │
│  │    - Creates GatewayClass and Gateway resources                │ │
│  │    - Exposes: ingressClassName, loadBalancerEndpoint           │ │
│  │                                                                │ │
│  │  IstioGatewayReconciler:                                       │ │
│  │    - Installs Istio control plane and gateway                  │ │
│  │    - Configures service mesh settings                          │ │
│  │    - Exposes: ingressClassName, loadBalancerEndpoint           │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌────────────────────────────────────────────────────────────────┐ │
│  │  GitOps Provider Controllers (Duck-Typed)                      │ │
│  │                                                                │ │
│  │  ArgoCDProviderReconciler:                                     │ │
│  │    - Installs ArgoCD via Helm                                  │ │
│  │    - Creates projects and admin credentials                    │ │
│  │    - Exposes: endpoint, internalEndpoint, credentialsSecretRef │ │
│  │                                                                │ │
│  │  FluxProviderReconciler:                                       │ │
│  │    - Installs Flux controllers via Helm                        │ │
│  │    - Configures source and kustomize controllers               │ │
│  │    - Exposes: endpoint, internalEndpoint, credentialsSecretRef │ │
│  └────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  GitRepositoryReconciler: (Enhanced)                                 │
│    - Works with ANY Git provider via duck-typed interface           │
│    - Creates repositories using provider's credentials              │
│    - Synchronizes content from multiple sources                     │
│                                                                      │
│  PackageReconciler: (Enhanced)                                       │
│    - Manages application packages                                   │
│    - Creates ArgoCD Applications referencing Git providers          │
│    - Handles package dependencies                                   │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                       Platform Services                              │
│  ┌───────────────┐  ┌──────────────┐  ┌──────────────────────┐     │
│  │ Git Providers │  │   Gateways   │  │  GitOps Providers    │     │
│  ├───────────────┤  ├──────────────┤  ├──────────────────────┤     │
│  │ • Gitea       │  │ • Nginx      │  │ • ArgoCD             │     │
│  │ • GitHub      │  │ • Envoy      │  │ • Flux               │     │
│  │ • GitLab      │  │ • Istio      │  │   (manages user apps │     │
│  │               │  │              │  │    via GitOps)       │     │
│  └───────────────┘  └──────────────┘  └──────────────────────┘     │
│                                                                      │
│  Multiple providers can coexist - e.g.:                              │
│    - Gitea for dev + GitHub for production                           │
│    - Nginx for public + Envoy for internal/service mesh             │
│    - ArgoCD for app deployment + Flux for infrastructure            │
└─────────────────────────────────────────────────────────────────────┘
```

**Key Architecture Principles:**

1. **Duck Typing**: Providers expose common status fields without requiring a shared interface type
2. **Composition**: Platform references multiple provider CRs by name and kind
3. **Extensibility**: New provider types can be added without modifying Platform CR
4. **Independence**: Each provider CR can exist and be managed independently
5. **Flexibility**: Components choose providers dynamically at runtime

### New Custom Resource Definitions

#### Platform CR

The Platform CR represents the entire IDP platform instance:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: localdev
  namespace: idpbuilder-system
spec:
  # Platform-wide configuration
  domain: cnoe.localtest.me
  ingressConfig:
    provider: nginx
    usePathRouting: false
    tlsSecretRef:
      name: platform-tls
      namespace: idpbuilder-system
  
  # Component specifications with provider references
  components:
    # Git Providers - references to provider CRs
    gitProviders:
      - name: gitea-local
        kind: GiteaProvider
        namespace: idpbuilder-system
      # Additional providers can be added
      # - name: github-external
      #   kind: GitHubProvider
      #   namespace: idpbuilder-system
    
    # Gateways - references to gateway provider CRs
    gateways:
      - name: nginx-gateway
        kind: NginxGateway
        namespace: idpbuilder-system
      # Additional gateways can be added
      # - name: envoy-gateway
      #   kind: EnvoyGateway
      #   namespace: idpbuilder-system
    
    # GitOps Providers - references to GitOps provider CRs
    gitOpsProviders:
      - name: argocd
        kind: ArgoCDProvider
        namespace: idpbuilder-system
      # Additional GitOps providers can be added
      # - name: flux
      #   kind: FluxProvider
      #   namespace: idpbuilder-system
  
  # GitOps bootstrap configuration
  bootstrap:
    gitServerRef:
      name: gitea
    repositories:
      - name: argocd-bootstrap
        path: hack/argo-cd
        autoSync: true
      - name: gitea-bootstrap
        path: hack/gitea
        autoSync: true
      - name: nginx-bootstrap
        path: hack/ingress-nginx
        autoSync: true

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
      reason: AllComponentsReady
      message: "All platform components are operational"
  
  # Provider statuses (aggregated from provider CRs)
  providers:
    gitProviders:
      - name: gitea-local
        kind: GiteaProvider
        ready: true
    gateways:
      - name: nginx-gateway
        kind: NginxGateway
        ready: true
    gitOpsProviders:
      - name: argocd
        kind: ArgoCDProvider
        ready: true
  
  observedGeneration: 1
  phase: Ready
```

#### Git Provider CRs (Duck-Typed)

Git providers are defined as separate CR types that share common status fields, allowing other controllers to interact with them uniformly regardless of implementation.

**Common Status Fields (Duck-Typed Interface):**

All Git provider CRs must expose these status fields:
```yaml
status:
  # Standard conditions
  conditions:
    - type: Ready
      status: "True"
  
  # Common fields for Git operations
  endpoint: string           # External URL for web UI and cloning
  internalEndpoint: string   # Cluster-internal URL for API access
  credentialsSecretRef:      # Secret containing access credentials
    name: string
    namespace: string
    key: string              # Key within secret (e.g., "token", "password")
```

##### GiteaProvider CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GiteaProvider
metadata:
  name: gitea-local
  namespace: idpbuilder-system
spec:
  # Deployment namespace
  namespace: gitea
  version: 1.24.3
  
  # Installation method
  installMethod:
    type: Helm
    helm:
      repository: https://dl.gitea.com/charts/
      chart: gitea
      version: 12.1.2
  
  # Gitea-specific configuration
  config:
    ingress:
      enabled: true
      className: nginx
      host: gitea.cnoe.localtest.me
    
    persistence:
      enabled: true
      size: 10Gi
    
    database:
      type: sqlite  # Options: sqlite, postgres, mysql
  
  # Admin user configuration
  adminUser:
    username: giteaAdmin
    email: admin@cnoe.localtest.me
    passwordSecretRef:
      name: gitea-admin-secret
      namespace: gitea
      key: password
    autoGenerate: true
  
  # Organizations to create
  organizations:
    - name: idpbuilder
      description: IDP Builder Bootstrap Organization

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  endpoint: https://gitea.cnoe.localtest.me
  internalEndpoint: http://gitea-http.gitea.svc.cluster.local:3000
  credentialsSecretRef:
    name: gitea-admin-secret
    namespace: gitea
    key: token
  
  # Gitea-specific status
  installed: true
  version: 1.24.3
  phase: Ready
  adminUser:
    username: giteaAdmin
```

##### GitHubProvider CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GitHubProvider
metadata:
  name: github-external
  namespace: idpbuilder-system
spec:
  # GitHub organization
  organization: my-organization
  
  # GitHub API endpoint (for GitHub Enterprise)
  endpoint: https://api.github.com  # Default for github.com
  
  # Credentials for GitHub API access
  credentialsSecretRef:
    name: github-credentials
    namespace: idpbuilder-system
    key: token  # GitHub Personal Access Token or App private key
  
  # Authentication method
  authType: token  # Options: token, app
  
  # For GitHub App authentication
  appAuth:
    appID: "123456"
    installationID: "789012"
    privateKeySecretRef:
      name: github-app-key
      namespace: idpbuilder-system
      key: private-key.pem
  
  # Repository defaults
  repositoryDefaults:
    visibility: private  # Options: public, private, internal
    autoInit: true
    defaultBranch: main
    allowSquashMerge: true
    allowMergeCommit: true
    allowRebaseMerge: true
  
  # Team configuration
  teams:
    - name: platform-team
      permission: admin  # Options: pull, push, admin, maintain

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  endpoint: https://github.com/my-organization
  internalEndpoint: https://api.github.com
  credentialsSecretRef:
    name: github-credentials
    namespace: idpbuilder-system
    key: token
  
  # GitHub-specific status
  organization: my-organization
  authenticated: true
  rateLimit:
    remaining: 4999
    limit: 5000
    resetAt: "2025-12-19T11:00:00Z"
```

##### GitLabProvider CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GitLabProvider
metadata:
  name: gitlab-external
  namespace: idpbuilder-system
spec:
  # GitLab instance URL
  endpoint: https://gitlab.com  # Or self-hosted GitLab URL
  
  # Group path
  group: my-group
  
  # Credentials for GitLab API access
  credentialsSecretRef:
    name: gitlab-credentials
    namespace: idpbuilder-system
    key: token  # GitLab Personal Access Token or Group Access Token
  
  # For self-hosted GitLab with custom CA
  caSecretRef:
    name: gitlab-ca-cert
    namespace: idpbuilder-system
    key: ca.crt
  
  # Repository defaults
  repositoryDefaults:
    visibility: private  # Options: public, private, internal
    defaultBranch: main
    initializeWithReadme: true
    cicdEnabled: true
  
  # Subgroup configuration
  subgroups:
    - name: platform
      description: Platform repositories

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  endpoint: https://gitlab.com/my-group
  internalEndpoint: https://gitlab.com/api/v4
  credentialsSecretRef:
    name: gitlab-credentials
    namespace: idpbuilder-system
    key: token
  
  # GitLab-specific status
  group: my-group
  groupID: 12345
  authenticated: true
```

#### Gateway Provider CRs (Duck-Typed)

Gateway providers are defined as separate CR types that share common status fields, allowing other controllers to route traffic through them uniformly regardless of implementation.

**Common Status Fields (Duck-Typed Interface):**

All Gateway provider CRs must expose these status fields:
```yaml
status:
  # Standard conditions
  conditions:
    - type: Ready
      status: "True"
  
  # Common fields for gateway operations
  ingressClassName: string   # Ingress class name to use in Ingress resources
  loadBalancerEndpoint: string # External endpoint for accessing services
  internalEndpoint: string   # Cluster-internal API endpoint
```

##### NginxGateway CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: NginxGateway
metadata:
  name: nginx-gateway
  namespace: idpbuilder-system
spec:
  # Deployment namespace
  namespace: ingress-nginx
  version: 1.13.0
  
  # Installation method
  installMethod:
    type: Helm
    helm:
      repository: https://kubernetes.github.io/ingress-nginx
      chart: ingress-nginx
      version: 4.11.0
  
  # Nginx-specific configuration
  config:
    controller:
      service:
        type: NodePort
        nodePorts:
          http: 30080
          https: 30443
      
      resources:
        limits:
          cpu: 100m
          memory: 90Mi
        requests:
          cpu: 100m
          memory: 90Mi
      
      admissionWebhooks:
        enabled: true
      
      config:
        use-forwarded-headers: "true"
        compute-full-forwarded-for: "true"
  
  # Ingress class configuration
  ingressClass:
    name: nginx
    isDefault: true
  
  # TLS configuration
  defaultTLS:
    secretRef:
      name: platform-tls
      namespace: idpbuilder-system

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  ingressClassName: nginx
  loadBalancerEndpoint: http://172.18.0.2
  internalEndpoint: http://ingress-nginx-controller.ingress-nginx.svc.cluster.local
  
  # Nginx-specific status
  installed: true
  version: 1.13.0
  phase: Ready
  controller:
    replicas: 1
    readyReplicas: 1
```

##### EnvoyGateway CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: EnvoyGateway
metadata:
  name: envoy-gateway
  namespace: idpbuilder-system
spec:
  # Deployment namespace
  namespace: envoy-gateway-system
  version: v1.0.0
  
  # Installation method
  installMethod:
    type: Helm
    helm:
      repository: https://gateway-envoyproxy.io
      chart: gateway-helm
      version: v1.0.0
  
  # Envoy Gateway-specific configuration
  config:
    provider:
      type: Kubernetes
    
    gateway:
      controllerName: gateway.envoyproxy.io/gatewayclass-controller
    
    resources:
      limits:
        cpu: 500m
        memory: 256Mi
      requests:
        cpu: 100m
        memory: 128Mi
  
  # Gateway class configuration
  gatewayClass:
    name: envoy
    isDefault: false
  
  # Listener configuration
  listeners:
    - name: http
      protocol: HTTP
      port: 80
    - name: https
      protocol: HTTPS
      port: 443
      tls:
        mode: Terminate
        certificateRefs:
          - name: platform-tls
            namespace: idpbuilder-system

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  ingressClassName: envoy
  loadBalancerEndpoint: http://172.18.0.3
  internalEndpoint: http://envoy-gateway.envoy-gateway-system.svc.cluster.local
  
  # Envoy-specific status
  installed: true
  version: v1.0.0
  phase: Ready
  gatewayClass: envoy
  gateway:
    ready: true
    addresses:
      - type: IPAddress
        value: 172.18.0.3
```

##### IstioGateway CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: IstioGateway
metadata:
  name: istio-gateway
  namespace: idpbuilder-system
spec:
  # Istio profile
  profile: default  # Options: default, demo, minimal, ambient
  
  # Deployment namespace for Istio control plane
  namespace: istio-system
  version: 1.24.0
  
  # Installation method
  installMethod:
    type: Helm
    helm:
      repository: https://istio-release.storage.googleapis.com/charts
      chart: gateway
      version: 1.24.0
  
  # Istio-specific configuration
  config:
    pilot:
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 128Mi
    
    gateways:
      istio-ingressgateway:
        enabled: true
        type: NodePort
        ports:
          - name: http
            nodePort: 31080
            port: 80
          - name: https
            nodePort: 31443
            port: 443
  
  # Gateway resource configuration
  gateway:
    name: istio-ingressgateway
    selector:
      istio: ingressgateway
  
  # Service mesh features
  mesh:
    mtls:
      mode: PERMISSIVE  # Options: STRICT, PERMISSIVE, DISABLE
    
    observability:
      tracing: true
      metrics: true

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  ingressClassName: istio
  loadBalancerEndpoint: http://172.18.0.4
  internalEndpoint: http://istio-ingressgateway.istio-system.svc.cluster.local
  
  # Istio-specific status
  installed: true
  version: 1.24.0
  phase: Ready
  profile: default
  controlPlane:
    ready: true
    version: 1.24.0
  gateway:
    ready: true
    addresses:
      - type: IPAddress
        value: 172.18.0.4
```

**Alternative: Envoy Gateway Provider Configuration**

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Gateway
metadata:
  name: gateway
  namespace: idpbuilder-system
spec:
  provider: envoy
  
  # Envoy Gateway provider configuration
  envoy:
    namespace: envoy-gateway-system
    version: v1.0.0
    
    installMethod:
      type: Helm
      helm:
        repository: oci://docker.io/envoyproxy/gateway-helm
        chart: gateway-helm
        version: v1.0.0
    
    config:
      deployment:
        replicas: 1
        resources:
          limits:
            cpu: 500m
            memory: 512Mi
      
      # Gateway class configuration
      gatewayClass:
        name: idp-gateway
        controllerName: gateway.envoyproxy.io/gatewayclass-controller

status:
  conditions:
    - type: Ready
      status: "True"
  provider: envoy
  installed: true
  version: v1.0.0
  phase: Ready
  gatewayClassName: idp-gateway
```

#### GitOps Provider CRs (Duck-Typed)

GitOps providers are defined as separate CR types that share common status fields, allowing other controllers to create and manage GitOps applications uniformly regardless of implementation.

**Common Status Fields (Duck-Typed Interface):**

All GitOps provider CRs must expose these status fields:
```yaml
status:
  # Standard conditions
  conditions:
    - type: Ready
      status: "True"
  
  # Common fields for GitOps operations
  endpoint: string              # External URL for web UI
  internalEndpoint: string      # Cluster-internal API endpoint
  credentialsSecretRef:         # Admin credentials
    name: string
    namespace: string
    key: string
```

##### ArgoCDProvider CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: ArgoCDProvider
metadata:
  name: argocd
  namespace: idpbuilder-system
spec:
  # Deployment namespace
  namespace: argocd
  version: v2.12.0
  
  # Installation method
  installMethod:
    type: Helm
    helm:
      repository: https://argoproj.github.io/argo-helm
      chart: argo-cd
      version: 7.0.0
  
  # ArgoCD-specific configuration
  config:
    server:
      ingress:
        enabled: true
        ingressClassName: nginx
        hosts:
          - argocd.cnoe.localtest.me
      extraArgs:
        - --insecure
    
    controller:
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 256Mi
    
    # Disable components for lightweight deployment
    notifications:
      enabled: false
    dex:
      enabled: false
  
  # Admin credentials
  adminCredentials:
    secretRef:
      name: argocd-admin-secret
      namespace: argocd
    # Auto-generate if not provided
    autoGenerate: true
  
  # Projects to create
  projects:
    - name: default
      description: Default project
    - name: platform
      description: Platform components

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  endpoint: https://argocd.cnoe.localtest.me
  internalEndpoint: http://argocd-server.argocd.svc.cluster.local
  credentialsSecretRef:
    name: argocd-admin-secret
    namespace: argocd
    key: password
  
  # ArgoCD-specific status
  installed: true
  version: v2.12.0
  phase: Ready
  serverHealth:
    status: Healthy
  applicationController:
    ready: true
```

##### FluxProvider CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: FluxProvider
metadata:
  name: flux
  namespace: idpbuilder-system
spec:
  # Deployment namespace
  namespace: flux-system
  version: v2.4.0
  
  # Installation method
  installMethod:
    type: Helm
    helm:
      repository: https://fluxcd-community.github.io/helm-charts
      chart: flux2
      version: 2.14.0
  
  # Flux-specific configuration
  config:
    # Source Controller configuration
    sourceController:
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 256Mi
    
    # Kustomize Controller configuration
    kustomizeController:
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 256Mi
    
    # Helm Controller configuration
    helmController:
      resources:
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 256Mi
    
    # Notification Controller (optional)
    notificationController:
      enabled: true
  
  # Multi-tenancy configuration
  multitenancy:
    enabled: true
    defaultServiceAccount: flux-reconciler
  
  # Image automation (optional)
  imageAutomation:
    enabled: false

status:
  conditions:
    - type: Ready
      status: "True"
      lastTransitionTime: "2025-12-19T10:00:00Z"
  
  # Common fields (duck-typed interface)
  endpoint: https://flux-dashboard.cnoe.localtest.me
  internalEndpoint: http://flux-source-controller.flux-system.svc.cluster.local
  credentialsSecretRef:
    name: flux-admin-secret
    namespace: flux-system
    key: token
  
  # Flux-specific status
  installed: true
  version: v2.4.0
  phase: Ready
  controllers:
    sourceController:
      ready: true
    kustomizeController:
      ready: true
    helmController:
      ready: true
    notificationController:
      ready: true
```

#### Gitea Component CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GiteaComponent
metadata:
  name: gitea
  namespace: idpbuilder-system
  ownerReferences:
    - apiVersion: idpbuilder.cnoe.io/v1alpha1
      kind: Platform
      name: localdev
spec:
  namespace: gitea
  version: 1.24.3
  
  installMethod:
    type: Helm
    helm:
      repository: https://dl.gitea.com/charts/
      chart: gitea
      version: 12.1.2
  
  config:
    ingress:
      enabled: true
      className: nginx
      hosts:
        - host: gitea.cnoe.localtest.me
    
    persistence:
      enabled: true
      size: 10Gi
      storageClass: standard
    
    postgresql:
      enabled: false
    
    # Use SQLite for development
    database:
      builtIn:
        sqlite:
          enabled: true
  
  # Admin user configuration
  adminUser:
    username: giteaAdmin
    email: admin@cnoe.localtest.me
    passwordSecretRef:
      name: gitea-admin-secret
      namespace: gitea
      key: password
    autoGenerate: true
  
  # Organizations to create
  organizations:
    - name: idpbuilder
      description: IDP Builder Bootstrap Organization
  
status:
  conditions:
    - type: Ready
      status: "True"
  installed: true
  version: 1.24.3
  phase: Ready
  endpoint: https://gitea.cnoe.localtest.me
  internalEndpoint: http://gitea-http.gitea.svc.cluster.local:3000
  adminUser:
    username: giteaAdmin
    secretRef:
      name: gitea-admin-secret
      namespace: gitea
```

#### Nginx Component CR

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: NginxComponent
metadata:
  name: nginx
  namespace: idpbuilder-system
spec:
  namespace: ingress-nginx
  version: 1.13.0
  
  installMethod:
    type: Helm
    helm:
      repository: https://kubernetes.github.io/ingress-nginx
      chart: ingress-nginx
      version: 4.11.0
  
  config:
    controller:
      service:
        type: NodePort
        nodePorts:
          http: 30080
          https: 30443
      
      resources:
        limits:
          cpu: 100m
          memory: 90Mi
        requests:
          cpu: 100m
          memory: 90Mi
      
      admissionWebhooks:
        enabled: true
  
  # TLS configuration
  defaultTLS:
    secretRef:
      name: platform-tls
      namespace: idpbuilder-system
  
status:
  conditions:
    - type: Ready
      status: "True"
  installed: true
  version: 1.13.0
  phase: Ready
  loadBalancerIP: 172.18.0.2
```

### Enhanced GitRepository CR

Enhance the existing GitRepository CR to support more advanced scenarios:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GitRepository
metadata:
  name: platform-configs
  namespace: idpbuilder-system
spec:
  # Git server reference (supports multiple servers)
  gitServerRef:
    kind: GiteaComponent
    name: gitea
    namespace: idpbuilder-system
  
  # Repository configuration
  name: platform-configs
  organization: idpbuilder
  description: Platform configuration repository
  private: false
  autoInit: true
  
  # Content sources (multiple sources can be combined)
  sources:
    - type: Embedded
      path: pkg/controllers/localbuild/resources/argocd
      targetPath: argocd/
    
    - type: Local
      path: /path/to/local/manifests
      targetPath: custom/
    
    - type: Git
      url: https://github.com/external/repo.git
      ref: main
      path: configs/
      targetPath: external/
  
  # Sync behavior
  sync:
    autoSync: true
    interval: 5m
  
status:
  conditions:
    - type: Ready
      status: "True"
  url: https://gitea.cnoe.localtest.me/idpbuilder/platform-configs.git
  cloneURL: http://gitea-http.gitea.svc.cluster.local:3000/idpbuilder/platform-configs.git
  lastSyncTime: "2025-12-19T10:00:00Z"
  commitSHA: abc123def456
```

### New Package CR

Replace CustomPackage with a more comprehensive Package CR:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Package
metadata:
  name: backstage
  namespace: idpbuilder-system
  annotations:
    cnoe.io/package-priority: "100"
spec:
  # Package source
  source:
    gitRepository:
      name: backstage-manifests
      namespace: idpbuilder-system
      path: manifests/
  
  # ArgoCD Application configuration
  argocd:
    enabled: true
    project: default
    destination:
      server: https://kubernetes.default.svc
      namespace: backstage
    
    syncPolicy:
      automated:
        prune: true
        selfHeal: true
      syncOptions:
        - CreateNamespace=true
  
  # Dependencies (ensures ordering)
  dependencies:
    - kind: ArgoCDProvider
      name: argocd
    - kind: GiteaComponent
      name: gitea
    - kind: Package
      name: postgresql
  
  # Health checks
  healthChecks:
    - type: Deployment
      namespace: backstage
      name: backstage
      timeout: 10m
  
status:
  conditions:
    - type: Ready
      status: "True"
  phase: Synced
  argoApplication:
    name: backstage
    namespace: argocd
    syncStatus: Synced
    healthStatus: Healthy
```

## Implementation Plan

This implementation plan follows an **iterative, end-to-end approach**. Instead of implementing all APIs first and then all controllers, we'll implement a narrow vertical slice that migrates existing idpbuilder functionality to the new architecture. Once that's working end-to-end, we'll add alternative providers incrementally.

### Phase 1: Core End-to-End Implementation (Existing Providers)

**Objective**: Implement a working end-to-end platform controller architecture using the existing providers (Gitea, Nginx, ArgoCD) that replicates current idpbuilder functionality. This phase is broken into iterative sub-phases where each sub-phase delivers end-to-end functionality that can be implemented, validated, and merged independently.

#### Scope:
- Implement **only** the providers that exist today: Gitea, Nginx, ArgoCD
- Migrate existing LocalbuildReconciler logic to new controller architecture incrementally
- Validate the duck-typing pattern works with real implementations
- Achieve feature parity with current CLI-driven installation
- Each sub-phase builds on the previous one and can be merged independently

#### Implementation Approach

Instead of implementing all APIs and controllers simultaneously, we implement vertical slices that deliver end-to-end functionality incrementally. Each sub-phase implements:
1. The necessary CRD definitions for that provider
2. The provider controller implementation
3. The Platform CR changes to support that provider
4. Duck-typing infrastructure for that provider type
5. Tests and validation for that specific provider
6. Documentation updates

This allows each sub-phase to be:
- **Implemented** independently by a developer
- **Validated** with end-to-end tests
- **Merged** into main without breaking existing functionality
- **Used** immediately by early adopters

---

### Sub-Phase 1.1: Platform CR + GiteaProvider (Git Provider Foundation)

**Objective**: Establish the foundation with Platform CR and the first provider (GiteaProvider), proving the duck-typing pattern and basic orchestration.

#### Tasks:

1. **Core Infrastructure Setup**
   - Create `api/v1alpha2/` directory structure for new API version
   - Set up code generation tools (controller-gen, deepcopy-gen)
   - Create base types and interfaces for duck-typing
   - Add necessary dependencies (controller-runtime, kubebuilder markers)

2. **Platform CR Definition (Initial)**
   ```yaml
   apiVersion: idpbuilder.cnoe.io/v1alpha2
   kind: Platform
   metadata:
     name: localdev
     namespace: idpbuilder-system
   spec:
     domain: cnoe.localtest.me
     components:
       gitProviders:
         - name: gitea-local
           kind: GiteaProvider
           namespace: idpbuilder-system
   status:
     conditions:
       - type: Ready
         status: "True"
     providers:
       gitProviders:
         - name: gitea-local
           kind: GiteaProvider
           ready: true
   ```
   - Define minimal Platform CR with only gitProviders support
   - Implement status aggregation for git providers
   - Add conditions and phase tracking
   - Generate CRD manifests

3. **GiteaProvider CR Definition**
   ```yaml
   apiVersion: idpbuilder.cnoe.io/v1alpha2
   kind: GiteaProvider
   metadata:
     name: gitea-local
     namespace: idpbuilder-system
   spec:
     namespace: gitea
     version: 1.24.3
     adminUser:
       username: giteaAdmin
       email: admin@cnoe.localtest.me
       autoGenerate: true
     organizations:
       - name: idpbuilder
         description: IDP Builder Bootstrap Organization
   status:
     conditions:
       - type: Ready
         status: "True"
     # Duck-typed fields
     endpoint: https://gitea.cnoe.localtest.me
     internalEndpoint: http://gitea-http.gitea.svc.cluster.local:3000
     credentialsSecretRef:
       name: gitea-admin-secret
       namespace: gitea
       key: token
   ```
   - Define GiteaProvider CR with spec and status
   - Implement duck-typed status fields (endpoint, internalEndpoint, credentialsSecretRef)
   - Add Gitea-specific configuration
   - Generate CRD manifests

4. **Duck-Typing Infrastructure for Git Providers**
   ```go
   // pkg/util/provider/git.go
   package provider

   import (
       "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
   )

   // GitProviderStatus represents the duck-typed interface for Git providers
   type GitProviderStatus struct {
       Endpoint             string
       InternalEndpoint     string
       CredentialsSecretRef SecretReference
       Ready                bool
   }

   type SecretReference struct {
       Name      string
       Namespace string
       Key       string
   }

   // GetGitProviderStatus extracts duck-typed status from any Git provider CR
   func GetGitProviderStatus(obj *unstructured.Unstructured) (*GitProviderStatus, error) {
       // Extract status fields using unstructured access
       // Return GitProviderStatus or error if required fields missing
   }
   ```
   - Create provider utility package
   - Implement GitProviderStatus struct
   - Implement GetGitProviderStatus() using unstructured access
   - Add validation and error handling
   - Create unit tests for duck-typing

5. **GiteaProviderReconciler Implementation**
   ```go
   // pkg/controllers/gitprovider/gitea_controller.go
   package gitprovider

   import (
       "context"
       ctrl "sigs.k8s.io/controller-runtime"
       "sigs.k8s.io/controller-runtime/pkg/client"
       
       idpbuilderv1alpha2 "github.com/cnoe-io/idpbuilder/api/v1alpha2"
   )

   type GiteaProviderReconciler struct {
       client.Client
       Scheme *runtime.Scheme
   }

   func (r *GiteaProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // 1. Fetch GiteaProvider CR
       // 2. Install Gitea using existing embedded manifests (reuse from LocalbuildReconciler)
       // 3. Wait for Gitea to be ready
       // 4. Create admin user and organization using Gitea API
       // 5. Generate and store credentials in secret
       // 6. Update status with duck-typed fields (endpoint, internalEndpoint, credentialsSecretRef)
       // 7. Set Ready condition
   }

   func (r *GiteaProviderReconciler) SetupWithManager(mgr ctrl.Manager) error {
       return ctrl.NewControllerManagedBy(mgr).
           For(&idpbuilderv1alpha2.GiteaProvider{}).
           Complete(r)
   }
   ```
   - **MIGRATE** Gitea installation logic from `pkg/controllers/localbuild/gitea.go`
   - **MIGRATE** Gitea embedded manifests from `pkg/controllers/localbuild/resources/gitea/`
   - **MIGRATE** Gitea client integration and API calls from LocalbuildReconciler
   - Reuse existing embedded manifests without changes
   - Implement Gitea client integration
   - Create admin user and organization
   - Update duck-typed status fields
   - Add proper error handling and conditions

6. **Migration and Cleanup**
   - **REMOVE** Gitea-related code from `pkg/controllers/localbuild/controller.go`:
     - Remove `reconcileGitea()` function
     - Remove Gitea installation calls from main reconcile loop
   - **REMOVE** `pkg/controllers/localbuild/gitea.go` file (after migration complete)
   - **REMOVE** Gitea-related fields from Localbuild CR status (if any)
   - Update LocalbuildReconciler to skip Gitea installation
   - Add deprecation warnings for Localbuild CR usage

6. **PlatformReconciler Implementation (Minimal)**
   ```go
   // pkg/controllers/platform/controller.go
   package platform

   type PlatformReconciler struct {
       client.Client
       Scheme *runtime.Scheme
   }

   func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // 1. Fetch Platform CR
       // 2. For each gitProvider reference:
       //    a. Fetch provider CR (unstructured)
       //    b. Extract status using GetGitProviderStatus()
       //    c. Aggregate ready status
       // 3. Update Platform status with aggregated provider status
       // 4. Set Platform Ready condition based on all providers
   }
   ```
   - Implement basic Platform reconciliation
   - Support only gitProviders initially
   - Use duck-typing to access provider status
   - Aggregate provider status into Platform status
   - Update conditions

7. **Testing**
   - Unit tests for GiteaProviderReconciler
   - Unit tests for PlatformReconciler
   - Unit tests for duck-typing utilities
   - Integration test: Create Platform CR → GiteaProvider created → Gitea installed → Status updated
   - Validation: `kubectl get platform,giteaprovider -n idpbuilder-system`
   - **Verify** old Localbuild CR still works (for backward compatibility)
   - **Verify** Gitea functionality removed from LocalbuildReconciler works correctly

**Deliverables**:
- ✅ Platform CR (v1alpha2) with gitProviders support only
- ✅ GiteaProvider CRD and controller
- ✅ Duck-typing infrastructure for Git providers
- ✅ PlatformReconciler supporting git providers
- ✅ Unit and integration tests
- ✅ CRD manifests generated
- ✅ Documentation for GiteaProvider
- ✅ **MIGRATED**: Gitea installation logic from LocalbuildReconciler
- ✅ **REMOVED**: Gitea-related code from `pkg/controllers/localbuild/`

**Success Criteria**:
- Can create a Platform CR that references a GiteaProvider
- GiteaProvider installs Gitea successfully
- Platform status correctly aggregates GiteaProvider status
- Duck-typing access to GiteaProvider status works
- All tests pass
- Can be merged and deployed independently
- **Old Localbuild CR no longer installs Gitea** (functionality moved to GiteaProvider)
- Backward compatibility maintained (Localbuild CR still functions for other components)

**Validation Steps**:
```bash
# 1. Apply CRDs
kubectl apply -f config/crd/

# 2. Create GiteaProvider
kubectl apply -f examples/giteaprovider.yaml

# 3. Create Platform referencing GiteaProvider
kubectl apply -f examples/platform-gitea-only.yaml

# 4. Verify Gitea installation
kubectl get pods -n gitea
kubectl get giteaprovider gitea-local -n idpbuilder-system -o yaml

# 5. Verify Platform status
kubectl get platform localdev -n idpbuilder-system -o yaml

# 6. Access Gitea
curl https://gitea.cnoe.localtest.me
```

---

### Sub-Phase 1.2: Add NginxGateway Provider (Gateway Support)

**Objective**: Add gateway provider support to Platform CR and implement NginxGateway, enabling ingress functionality.

#### Tasks:

1. **Platform CR Update - Add Gateway Support**
   ```yaml
   apiVersion: idpbuilder.cnoe.io/v1alpha2
   kind: Platform
   metadata:
     name: localdev
     namespace: idpbuilder-system
   spec:
     domain: cnoe.localtest.me
     components:
       gitProviders:
         - name: gitea-local
           kind: GiteaProvider
           namespace: idpbuilder-system
       gateways:  # NEW
         - name: nginx-gateway
           kind: NginxGateway
           namespace: idpbuilder-system
   status:
     providers:
       gitProviders:
         - name: gitea-local
           kind: GiteaProvider
           ready: true
       gateways:  # NEW
         - name: nginx-gateway
           kind: NginxGateway
           ready: true
   ```
   - Update Platform CR spec to include gateways field
   - Update Platform CR status to aggregate gateway status
   - Update PlatformReconciler to handle gateways
   - Regenerate CRD manifests

2. **NginxGateway CR Definition**
   ```yaml
   apiVersion: idpbuilder.cnoe.io/v1alpha2
   kind: NginxGateway
   metadata:
     name: nginx-gateway
     namespace: idpbuilder-system
   spec:
     namespace: ingress-nginx
     version: 1.13.0
     ingressClass:
       name: nginx
       isDefault: true
   status:
     conditions:
       - type: Ready
         status: "True"
     # Duck-typed fields
     ingressClassName: nginx
     loadBalancerEndpoint: http://172.18.0.2
     internalEndpoint: http://ingress-nginx-controller.ingress-nginx.svc.cluster.local
   ```
   - Define NginxGateway CR with spec and status
   - Implement duck-typed status fields for gateways
   - Generate CRD manifests

3. **Duck-Typing Infrastructure for Gateway Providers**
   ```go
   // pkg/util/provider/gateway.go
   package provider

   type GatewayProviderStatus struct {
       IngressClassName      string
       LoadBalancerEndpoint  string
       InternalEndpoint      string
       Ready                 bool
   }

   func GetGatewayProviderStatus(obj *unstructured.Unstructured) (*GatewayProviderStatus, error) {
       // Extract gateway status fields using unstructured access
   }
   ```
   - Create GatewayProviderStatus struct
   - Implement GetGatewayProviderStatus() function
   - Add unit tests for gateway duck-typing

4. **NginxGatewayReconciler Implementation**
   ```go
   // pkg/controllers/gatewayprovider/nginx_controller.go
   type NginxGatewayReconciler struct {
       client.Client
       Scheme *runtime.Scheme
   }

   func (r *NginxGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // 1. Fetch NginxGateway CR
       // 2. Install Nginx Ingress using existing embedded manifests
       // 3. Wait for Nginx to be ready
       // 4. Get LoadBalancer endpoint
       // 5. Update status with duck-typed fields
       // 6. Set Ready condition
   }
   ```
   - **MIGRATE** Nginx installation logic from `pkg/controllers/localbuild/nginx.go`
   - **MIGRATE** Nginx embedded manifests from `pkg/controllers/localbuild/resources/nginx/`
   - Reuse existing embedded manifests without changes
   - Get LoadBalancer/NodePort endpoint
   - Update duck-typed status fields
   - Create IngressClass resource

5. **Migration and Cleanup**
   - **REMOVE** Nginx-related code from `pkg/controllers/localbuild/controller.go`:
     - Remove `reconcileNginx()` function
     - Remove Nginx installation calls from main reconcile loop
   - **REMOVE** `pkg/controllers/localbuild/nginx.go` file (after migration complete)
   - **REMOVE** Nginx-related fields from Localbuild CR status (if any)
   - Update LocalbuildReconciler to skip Nginx installation

6. **PlatformReconciler Update**
   - Update to handle gateways field in Platform spec
   - Aggregate gateway provider status
   - Update Platform status with gateway information
   - Ensure both git providers and gateways are tracked

7. **GiteaProvider Integration with Gateway**
   - Update GiteaProvider to create Ingress resource
   - Use gateway's ingressClassName from Platform
   - Set up Gitea ingress using the gateway
   - Test Gitea accessibility through Nginx

8. **Testing**
   - Unit tests for NginxGatewayReconciler
   - Unit tests for gateway duck-typing
   - Integration test: Platform with Gitea + Nginx
   - Validate Gitea accessible through Nginx ingress
   - Test Platform status aggregation with multiple provider types
   - **Verify** old Localbuild CR no longer installs Nginx
   - **Verify** Nginx functionality removed from LocalbuildReconciler

**Deliverables**:
- ✅ NginxGateway CRD and controller
- ✅ Updated Platform CR with gateway support
- ✅ Duck-typing infrastructure for gateways
- ✅ Updated PlatformReconciler
- ✅ Gitea accessible via Nginx ingress
- ✅ Tests for gateway functionality
- ✅ Documentation for NginxGateway
- ✅ **MIGRATED**: Nginx installation logic from LocalbuildReconciler
- ✅ **REMOVED**: Nginx-related code from `pkg/controllers/localbuild/`

**Success Criteria**:
- Platform CR can reference both git and gateway providers
- NginxGateway installs Nginx Ingress successfully
- Gitea is accessible through Nginx ingress
- Platform status aggregates both provider types
- Duck-typing works for gateways
- All tests pass
- Can be merged independently
- **Old Localbuild CR no longer installs Nginx** (functionality moved to NginxGateway)
- Backward compatibility maintained (Localbuild CR still functions for remaining components)

**Validation Steps**:
```bash
# 1. Apply updated CRDs
kubectl apply -f config/crd/

# 2. Create NginxGateway
kubectl apply -f examples/nginxgateway.yaml

# 3. Update Platform to include gateway
kubectl apply -f examples/platform-gitea-nginx.yaml

# 4. Verify Nginx installation
kubectl get pods -n ingress-nginx
kubectl get nginxgateway nginx-gateway -n idpbuilder-system -o yaml

# 5. Verify Platform status includes both providers
kubectl get platform localdev -n idpbuilder-system -o yaml

# 6. Test Gitea access through ingress
curl -k https://gitea.cnoe.localtest.me
```

---

### Sub-Phase 1.3: Add ArgoCDProvider (GitOps Support)

**Objective**: Add GitOps provider support to Platform CR and implement ArgoCDProvider, enabling GitOps-based deployments.

#### Tasks:

1. **Platform CR Update - Add GitOps Provider Support**
   ```yaml
   apiVersion: idpbuilder.cnoe.io/v1alpha2
   kind: Platform
   metadata:
     name: localdev
     namespace: idpbuilder-system
   spec:
     domain: cnoe.localtest.me
     components:
       gitProviders:
         - name: gitea-local
           kind: GiteaProvider
           namespace: idpbuilder-system
       gateways:
         - name: nginx-gateway
           kind: NginxGateway
           namespace: idpbuilder-system
       gitOpsProviders:  # NEW
         - name: argocd
           kind: ArgoCDProvider
           namespace: idpbuilder-system
   status:
     providers:
       gitProviders:
         - name: gitea-local
           kind: GiteaProvider
           ready: true
       gateways:
         - name: nginx-gateway
           kind: NginxGateway
           ready: true
       gitOpsProviders:  # NEW
         - name: argocd
           kind: ArgoCDProvider
           ready: true
   ```
   - Update Platform CR spec to include gitOpsProviders field
   - Update Platform CR status to aggregate GitOps provider status
   - Update PlatformReconciler to handle GitOps providers
   - Regenerate CRD manifests

2. **ArgoCDProvider CR Definition**
   ```yaml
   apiVersion: idpbuilder.cnoe.io/v1alpha2
   kind: ArgoCDProvider
   metadata:
     name: argocd
     namespace: idpbuilder-system
   spec:
     namespace: argocd
     version: v2.12.0
     adminCredentials:
       autoGenerate: true
     projects:
       - name: default
       - name: platform
   status:
     conditions:
       - type: Ready
         status: "True"
     # Duck-typed fields
     endpoint: https://argocd.cnoe.localtest.me
     internalEndpoint: http://argocd-server.argocd.svc.cluster.local
     credentialsSecretRef:
       name: argocd-admin-secret
       namespace: argocd
       key: password
   ```
   - Define ArgoCDProvider CR with spec and status
   - Implement duck-typed status fields for GitOps providers
   - Generate CRD manifests

3. **Duck-Typing Infrastructure for GitOps Providers**
   ```go
   // pkg/util/provider/gitops.go
   package provider

   type GitOpsProviderStatus struct {
       Endpoint             string
       InternalEndpoint     string
       CredentialsSecretRef SecretReference
       Ready                bool
   }

   func GetGitOpsProviderStatus(obj *unstructured.Unstructured) (*GitOpsProviderStatus, error) {
       // Extract GitOps provider status fields using unstructured access
   }
   ```
   - Create GitOpsProviderStatus struct
   - Implement GetGitOpsProviderStatus() function
   - Add unit tests for GitOps duck-typing

4. **ArgoCDProviderReconciler Implementation**
   ```go
   // pkg/controllers/gitopsprovider/argocd_controller.go
   type ArgoCDProviderReconciler struct {
       client.Client
       Scheme *runtime.Scheme
   }

   func (r *ArgoCDProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // 1. Fetch ArgoCDProvider CR
       // 2. Install ArgoCD using existing embedded manifests
       // 3. Wait for ArgoCD to be ready
       // 4. Create admin credentials and projects
       // 5. Update status with duck-typed fields
       // 6. Set Ready condition
   }
   ```
   - **MIGRATE** ArgoCD installation logic from `pkg/controllers/localbuild/argo.go`
   - **MIGRATE** ArgoCD embedded manifests from `pkg/controllers/localbuild/resources/argocd/`
   - Reuse existing embedded manifests without changes
   - Create admin credentials
   - Create ArgoCD projects
   - Update duck-typed status fields
   - Set up ArgoCD ingress using gateway

5. **Migration and Cleanup**
   - **REMOVE** ArgoCD-related code from `pkg/controllers/localbuild/controller.go`:
     - Remove `reconcileArgoCD()` function
     - Remove ArgoCD installation calls from main reconcile loop
     - Remove ArgoCD Application creation logic
   - **REMOVE** `pkg/controllers/localbuild/argo.go` file (after migration complete)
   - **REMOVE** ArgoCD-related fields from Localbuild CR status (if any)
   - Update LocalbuildReconciler to skip ArgoCD installation
   - **At this point, LocalbuildReconciler should be essentially empty** and can be deprecated

6. **Enhanced GitRepositoryReconciler**
   ```go
   // pkg/controllers/gitrepository/controller.go
   func (r *GitRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // 1. Fetch GitRepository CR
       // 2. Get Git provider reference from GitRepository
       // 3. Fetch provider CR using unstructured client
       // 4. Extract provider status using GetGitProviderStatus()
       // 5. Create repository using provider credentials
       // 6. Sync content from embedded sources
       // 7. Update GitRepository status
   }
   ```
   - Update existing GitRepositoryReconciler to use duck-typing
   - Support gitServerRef that can point to any Git provider kind
   - Use GetGitProviderStatus() to get credentials
   - **MIGRATE** repository creation logic to work with duck-typed providers
   - Reuse existing content sync logic

7. **PlatformReconciler Update - Bootstrap Integration**
   - Update to handle gitOpsProviders field
   - **MIGRATE** bootstrap repository creation from LocalbuildReconciler
   - **MIGRATE** ArgoCD Application creation from LocalbuildReconciler
   - Create GitRepository CRs for bootstrap content
   - Create ArgoCD Applications using providers
   - Aggregate GitOps provider status
   - Coordinate between git, gateway, and GitOps providers

8. **Bootstrap GitRepository CRs**
   - Create GitRepository CR for ArgoCD bootstrap
   - Create GitRepository CR for Gitea bootstrap
   - Create GitRepository CR for Nginx bootstrap
   - PlatformReconciler creates these automatically
   - **MIGRATE** embedded content references from LocalbuildReconciler

9. **Testing**
   - Unit tests for ArgoCDProviderReconciler
   - Unit tests for GitOps duck-typing
   - Unit tests for enhanced GitRepositoryReconciler
   - Integration test: Full Platform with all three providers
   - Validate ArgoCD Applications created and synced
   - Test GitRepository creation in Gitea via duck-typing
   - **Verify** LocalbuildReconciler no longer installs any components
   - **Verify** all functionality migrated to new provider controllers

**Deliverables**:
- ✅ ArgoCDProvider CRD and controller
- ✅ Updated Platform CR with GitOps provider support
- ✅ Duck-typing infrastructure for GitOps providers
- ✅ Enhanced GitRepositoryReconciler with duck-typing
- ✅ Updated PlatformReconciler with bootstrap logic
- ✅ Tests for full stack
- ✅ Documentation for ArgoCDProvider and GitRepository
- ✅ **MIGRATED**: ArgoCD installation logic from LocalbuildReconciler
- ✅ **MIGRATED**: Bootstrap repository and Application creation from LocalbuildReconciler
- ✅ **REMOVED**: ArgoCD-related code from `pkg/controllers/localbuild/`
- ✅ **REMOVED**: All component installation logic from LocalbuildReconciler

**Success Criteria**:
- Platform CR can reference git, gateway, and GitOps providers
- ArgoCDProvider installs ArgoCD successfully
- GitRepository controller works with GiteaProvider via duck-typing
- Bootstrap repositories created in Gitea
- ArgoCD Applications created and synced
- Platform status aggregates all three provider types
- All tests pass
- Full feature parity with existing LocalbuildReconciler
- Can be merged independently
- **LocalbuildReconciler is effectively deprecated** - all component installation moved to provider controllers
- **Old Localbuild CR is maintained for backward compatibility** but internally delegates to Platform CR (or shows deprecation warnings)

**Validation Steps**:
```bash
# 1. Apply updated CRDs
kubectl apply -f config/crd/

# 2. Create ArgoCDProvider
kubectl apply -f examples/argocdprovider.yaml

# 3. Update Platform to include all providers
kubectl apply -f examples/platform-full.yaml

# 4. Verify ArgoCD installation
kubectl get pods -n argocd
kubectl get argocdprovider argocd -n idpbuilder-system -o yaml

# 5. Verify GitRepository creation
kubectl get gitrepositories -n idpbuilder-system

# 6. Check Gitea for created repositories
curl -k https://gitea.cnoe.localtest.me/idpbuilder/

# 7. Verify ArgoCD Applications
kubectl get applications -n argocd

# 8. Verify Platform status includes all providers
kubectl get platform localdev -n idpbuilder-system -o yaml

# 9. Access ArgoCD UI
curl -k https://argocd.cnoe.localtest.me
```

---

### Sub-Phase 1.4: Integration, Testing, and Documentation

**Objective**: Ensure all components work together seamlessly, achieve comprehensive test coverage, and provide complete documentation.

#### Tasks:

1. **End-to-End Integration Testing**
   - Create comprehensive E2E test suite
   - Test Platform CR creation from scratch
   - Test provider lifecycle (create, update, delete)
   - Test provider failure and recovery scenarios
   - Test Platform status aggregation accuracy
   - Test concurrent reconciliation
   - Validate existing embedded manifests still work

2. **CLI Integration Preparation**
   - Design CLI flag mapping to new CRs
   - Plan controller deployment strategy
   - Design migration from Localbuild to Platform
   - Create example YAML files for users
   - Document CLI changes needed

3. **Performance and Resource Testing**
   - Measure controller resource usage
   - Compare with existing LocalbuildReconciler performance
   - Optimize reconciliation loops
   - Add resource limits to controller deployment
   - Test on different cluster sizes

4. **Error Handling and Observability**
   - Ensure all error paths have clear messages
   - Add proper logging throughout controllers
   - Implement event emission for key actions
   - Add status conditions for all failure modes
   - Create troubleshooting guide

5. **Documentation**
   - API reference for Platform, GiteaProvider, NginxGateway, ArgoCDProvider
   - Controller architecture deep-dive
   - Duck-typing pattern explanation
   - Migration guide from v1alpha1 to v1alpha2
   - Examples for common use cases
   - Developer guide for extending providers
   - Update main README with new architecture

6. **Code Quality**
   - Ensure test coverage >70%
   - Run linters and fix all issues
   - Code review checklist
   - Security review of credentials handling
   - Dependency audit

7. **Backward Compatibility Validation**
   - Ensure existing Localbuild CR still works (if kept)
   - Validate existing embedded manifests compatibility
   - Test with existing package definitions
   - Document any breaking changes

8. **Final Migration Cleanup**
   - **VERIFY** all component installation logic removed from LocalbuildReconciler
   - **VERIFY** `pkg/controllers/localbuild/gitea.go` removed
   - **VERIFY** `pkg/controllers/localbuild/nginx.go` removed
   - **VERIFY** `pkg/controllers/localbuild/argo.go` removed
   - Update LocalbuildReconciler to either:
     - Show deprecation warnings and delegate to Platform CR, OR
     - Maintain minimal compatibility shim for backward compatibility
   - Document deprecation path for Localbuild CR
   - Clean up any unused code or dependencies

**Deliverables**:
- ✅ Comprehensive E2E test suite
- ✅ Test coverage >70%
- ✅ Complete API documentation
- ✅ Architecture documentation
- ✅ Migration guide
- ✅ Example YAML files
- ✅ Troubleshooting guide
- ✅ Performance benchmarks
- ✅ Security review completed
- ✅ **FINAL CLEANUP**: All component installation code removed from LocalbuildReconciler
- ✅ **FINAL CLEANUP**: Individual component files (`gitea.go`, `nginx.go`, `argo.go`) removed from localbuild package

**Success Criteria**:
- All E2E tests pass consistently
- Test coverage exceeds 70%
- Documentation is complete and accurate
- Performance is equal or better than existing implementation
- No critical security issues
- Ready for Phase 2 (CLI integration)
- **All component installation code successfully migrated** from LocalbuildReconciler to provider controllers
- **LocalbuildReconciler cleaned up** with only backward compatibility shim remaining (if needed)

---

## Phase 1 Summary

**Overall Objective**: Deliver a working controller-based architecture with all three core providers (Gitea, Nginx, ArgoCD) that can be deployed and managed through Kubernetes CRs.

**Total Deliverables**:
- ✅ Platform CR (v1alpha2) with full provider support
- ✅ GiteaProvider CRD and controller
- ✅ NginxGateway CRD and controller
- ✅ ArgoCDProvider CRD and controller
- ✅ Duck-typing infrastructure for all provider types
- ✅ Enhanced GitRepositoryReconciler
- ✅ PlatformReconciler with full orchestration
- ✅ Test coverage >70%
- ✅ Complete documentation
- ✅ Feature parity with existing LocalbuildReconciler
- ✅ **MIGRATED**: All component installation logic from LocalbuildReconciler to provider controllers
- ✅ **REMOVED**: Component-specific code from `pkg/controllers/localbuild/` (gitea.go, nginx.go, argo.go)
- ✅ **CLEANED**: LocalbuildReconciler reduced to minimal compatibility shim (or deprecated)

**Migration Strategy Per Sub-Phase**:
- **Sub-Phase 1.1**: Migrate Gitea → Remove Gitea code from LocalbuildReconciler
- **Sub-Phase 1.2**: Migrate Nginx → Remove Nginx code from LocalbuildReconciler
- **Sub-Phase 1.3**: Migrate ArgoCD → Remove ArgoCD code from LocalbuildReconciler
- **Sub-Phase 1.4**: Final cleanup → Verify all migrations complete, remove old files

**Key Benefits of Iterative Approach**:
- Each sub-phase can be developed, tested, and merged independently
- Early validation of duck-typing pattern in Sub-Phase 1.1
- Progressive complexity (git → gateway → GitOps)
- Incremental migration reduces risk
- Faster feedback loops
- Lower risk of large merge conflicts
- Ability to course-correct between sub-phases
- Early adopters can start using partial functionality
- **Old functionality removed progressively** as new controllers prove stable

**Estimated Timeline**:
- Sub-Phase 1.1 (Platform + Gitea): 2-3 weeks
- Sub-Phase 1.2 (+ Nginx): 1-2 weeks
- Sub-Phase 1.3 (+ ArgoCD): 2-3 weeks
- Sub-Phase 1.4 (Integration + Docs): 1-2 weeks
- **Total Phase 1**: 6-10 weeks

### Phase 2: Component Controllers 

**Objective**: Implement individual component controllers with full lifecycle management

#### Tasks:

1. **GitOps Provider Controllers**

   Each GitOps provider has its own dedicated reconciler:

---

### Phase 2: CLI Integration & Controller Deployment

**Objective**: Implement CLI integration to support both development (CLI-driven) and production (GitOps-driven) deployment modes, with clear separation between CLI and controllers.

#### Tasks:

1. **Controller Packaging**
   - Create Helm chart for idpbuilder controllers
     ```
     charts/idpbuilder-controllers/
       Chart.yaml
       values.yaml
       templates/
         deployment.yaml          # Controller manager deployment
         rbac.yaml                # Service accounts, roles, bindings
         crds/                    # Platform, Provider CRDs
         namespace.yaml           # idpbuilder-system namespace
     ```
   - Build controller container image with all reconcilers
   - Support air-gapped installation with embedded images
   - Publish Helm chart to registry for production use
   - Create static manifest bundle (kubectl apply -f) as alternative

2. **CLI Update for Controller Deployment**
   ```go
   // pkg/cmd/create/root.go
   func createPlatform(ctx context.Context, opts *CreateOptions) error {
       // Phase 1: Infrastructure (CLI Responsibility)
       // 1. Create Kind cluster
       cluster := createKindCluster(opts.ClusterName)
       
       // Phase 2: Controller Installation (CLI Responsibility)
       // 2. Deploy idpbuilder controllers
       deployControllers(cluster, opts)  // Via Helm or embedded manifests
       
       // 3. Wait for controller manager to be ready
       waitForControllerReady(cluster)
       
       // Phase 3: Platform Creation (CLI Responsibility)
       // 4. Create provider CRs based on CLI flags
       createProviderCRs(cluster, opts)  // GiteaProvider, NginxGateway, etc.
       
       // 5. Create Platform CR referencing providers
       createPlatformCR(cluster, opts)
       
       // Phase 4: Wait and Display (CLI Responsibility)
       // 6. Wait for Platform to be Ready (controllers handle reconciliation)
       waitForPlatformReady(cluster)
       
       // 7. Display access info (endpoints, credentials)
       displayAccessInfo(cluster)
   }
   ```
   - CLI focuses on infrastructure and initial setup
   - Controllers handle all component installation and reconciliation
   - Maintain backward compatibility with existing flags
   - Auto-generate provider CRs from CLI flags
   - Display endpoints and credentials as before

3. **GitOps Installation Documentation**
   Create comprehensive documentation for GitOps-driven installation:
   
   **Production Installation Guide** (`docs/production-installation.md`):
   ```bash
   # Step 1: Install idpbuilder controllers (choose one method)
   
   # Method A: Helm (Recommended)
   helm repo add idpbuilder https://cnoe-io.github.io/idpbuilder
   helm install idpbuilder-controllers idpbuilder/idpbuilder-controllers \
     --namespace idpbuilder-system --create-namespace
   
   # Method B: kubectl with static manifests
   kubectl apply -f https://github.com/cnoe-io/idpbuilder/releases/latest/download/install.yaml
   
   # Step 2: Create Provider CRs via GitOps
   # Add to your GitOps repository (ArgoCD, Flux, etc.)
   # See examples/production/giteaprovider.yaml
   # See examples/production/nginxgateway.yaml
   # See examples/production/argocdprovider.yaml
   
   # Step 3: Create Platform CR via GitOps
   # Add to your GitOps repository
   # See examples/production/platform.yaml
   
   # The controllers will handle everything else!
   ```
   
   **Key Documentation Points**:
   - No CLI required for production deployments
   - Controllers are standard Kubernetes deployments
   - All configuration via CRs (declarative)
   - Full GitOps compatibility
   - RBAC requirements for controllers
   - Resource requirements and limits

4. **Flag Mapping for Development Mode**
   - Map `--package-dir` to Platform spec bootstrap repositories
   - Map `--custom-package` to Package CRs
   - Map `--port` to NginxGateway service port configuration
   - Map ingress/TLS flags to provider configurations
   - Map `--protocol` to Platform domain configuration
   - Maintain all existing CLI flags for backward compatibility

5. **Migration Command**
   ```bash
   idpbuilder migrate --cluster-name localdev
   ```
   - Detect existing Localbuild CR
   - Extract configuration
   - Deploy controllers (if not already present)
   - Create equivalent provider CRs + Platform CR
   - Validate migration success
   - Provide rollback option
   - Support dry-run mode

6. **Deployment Mode Detection**
   - CLI detects if controllers are already installed
   - If controllers present: Skip installation, only create CRs
   - If controllers absent: Install controllers, then create CRs
   - Environment variable to force specific mode
   - Clear logging of which mode is active

7. **Controller Health and Monitoring**
   - CLI checks controller health before creating CRs
   - Display controller logs on failure
   - Provide troubleshooting commands
   - Add `idpbuilder get controllers` command to check status

**Deliverables**:
- ✅ Helm chart for idpbuilder controllers
- ✅ Controller container image in registry
- ✅ Updated CLI with controller deployment logic
- ✅ CLI-to-Controller separation implemented
- ✅ Production installation documentation
- ✅ GitOps deployment examples
- ✅ Migration tool and documentation
- ✅ Updated CLI help and examples
- ✅ Clear documentation of two deployment modes
- ✅ Installation guide for bypassing CLI entirely

**Success Criteria**:
- CLI users can use new architecture transparently (development mode)
- Platform teams can deploy controllers without CLI (production mode)
- Controllers work identically in both modes
- Clear separation: CLI for infra/setup, controllers for reconciliation
- Full GitOps compatibility demonstrated
- Migration path tested and documented

---

### Phase 3: GitHub Provider (First Alternative)

**Objective**: Add GitHub as an alternative Git provider to validate the duck-typing pattern with external services.

#### Tasks:

1. **GitHubProvider CRD**
   - Define `GitHubProvider` CR in `api/v1alpha2/`
   - Implement common duck-typed status fields
   - Add GitHub-specific configuration (org, teams, auth)

2. **GitHubProviderReconciler**
   ```go
   // pkg/controllers/gitprovider/github_controller.go
   type GitHubProviderReconciler struct {
       client.Client
       Scheme *runtime.Scheme
   }

   func (r *GitHubProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // 1. Fetch GitHubProvider CR
       // 2. Validate GitHub credentials (no installation needed)
       // 3. Verify organization access
       // 4. Update status with duck-typed fields
   }
   ```

3. **GitHub Client Integration**
   - Add github.com/google/go-github dependency
   - Implement authentication (token, GitHub App)
   - Create repository management functions
   - Handle API rate limiting

4. **GitRepository Enhancement**
   - Update GitRepository controller to support GitHub
   - Implement GitHub-specific repository creation
   - Handle visibility settings (public/private)
   - Test with both Gitea and GitHub

5. **Testing**
   - Unit tests with mocked GitHub client
   - Integration tests (optional, may require real GitHub token)
   - Validate duck-typing works across providers

**Deliverables**:
- ✅ GitHubProvider CRD and controller
- ✅ GitHub client integration
- ✅ GitRepository works with both Gitea and GitHub
- ✅ Documentation for using GitHub as Git provider
- ✅ CLI support for `--git-provider=github`

---

### Phase 4: GitLab Provider (Second Alternative)

**Objective**: Add GitLab as another Git provider option.

#### Tasks:

1. **GitLabProvider CRD and Controller**
   - Define `GitLabProvider` CR
   - Implement GitLabProviderReconciler
   - Add gitlab.com/gitlab-org/api/client-go dependency
   - Support both gitlab.com and self-hosted GitLab

2. **Testing & Validation**
   - Validate duck-typing works with three Git providers
   - Ensure GitRepository controller handles all three
   - Document GitLab-specific configuration

**Deliverables**:
- ✅ GitLabProvider CRD and controller
- ✅ GitRepository supports three Git providers
- ✅ Documentation and examples

---

### Phase 5: Envoy Gateway Provider (First Alternative Gateway)

**Objective**: Add Envoy Gateway as an alternative ingress option.

#### Tasks:

1. **EnvoyGateway CRD and Controller**
   - Define `EnvoyGateway` CR
   - Implement EnvoyGatewayReconciler
   - Install Envoy Gateway via Helm
   - Configure Gateway API resources

2. **Multi-Gateway Support in Platform**
   - Allow Platform to reference multiple gateways
   - Components choose gateway via annotations
   - Test Nginx + Envoy running simultaneously

**Deliverables**:
- ✅ EnvoyGateway CRD and controller
- ✅ Platform supports multiple gateways
- ✅ Documentation for Envoy Gateway

---

### Phase 6: Istio Gateway Provider (Service Mesh Gateway)

**Objective**: Add Istio as a service mesh gateway option.

#### Tasks:

1. **IstioGateway CRD and Controller**
   - Define `IstioGateway` CR
   - Implement IstioGatewayReconciler
   - Install Istio via Helm
   - Configure Istio Gateway resources

**Deliverables**:
- ✅ IstioGateway CRD and controller
- ✅ Documentation for Istio integration

---

### Phase 7: Flux Provider (Alternative GitOps)

**Objective**: Add Flux as an alternative to ArgoCD for GitOps.

#### Tasks:

1. **FluxProvider CRD and Controller**
   - Define `FluxProvider` CR
   - Implement FluxProviderReconciler
   - Install Flux controllers via Helm
   - Create Flux source and sync resources

2. **Multi-GitOps Support**
   - Allow Platform to use multiple GitOps providers
   - Package controller supports both ArgoCD and Flux
   - Document use cases (ArgoCD for apps, Flux for infra)

**Deliverables**:
- ✅ FluxProvider CRD and controller
- ✅ Multi-GitOps documentation

---

### Phase 8: Production Features & Stabilization

**Objective**: Add production-ready features and comprehensive testing.

#### Tasks:

1. **High Availability**
   - Support multiple replicas for components
   - Leader election for controllers
   - Database persistence for Gitea
   - ArgoCD HA configuration

2. **Monitoring & Observability**
   - Prometheus metrics for all controllers
   - Component health dashboards
   - Alert rules for component failures
   - OpenTelemetry integration

3. **Security Enhancements**
   - RBAC for component CRs
   - Secret management improvements
   - TLS everywhere
   - Pod security standards

4. **Multi-Cluster Support**
   - Support vCluster as infrastructure provider
   - Support Cluster API
   - Remote cluster management

5. **Package Ecosystem**
   - Package catalog / marketplace
   - Package versioning
   - Package dependencies graph

6. **Comprehensive Testing**
   - E2E test coverage >80%
   - Chaos testing
   - Performance testing
   - Upgrade/downgrade testing

7. **Documentation**
   - Complete API reference
   - Architecture deep-dives
   - Operator guide
   - Developer guide
   - Troubleshooting runbooks

**Deliverables**:
- ✅ Production-ready release
- ✅ Complete documentation
- ✅ Test coverage >80%
- ✅ Release artifacts

---

## Phase Timeline & Milestones

**Phase 1** (Weeks 1-10): Core end-to-end with existing providers (Iterative)
- Sub-Phase 1.1 (Weeks 1-3): Platform + GiteaProvider
  - Milestone: Can deploy Platform CR with Gitea provider via duck-typing
- Sub-Phase 1.2 (Weeks 4-5): Add NginxGateway
  - Milestone: Gitea accessible through Nginx ingress
- Sub-Phase 1.3 (Weeks 6-8): Add ArgoCDProvider
  - Milestone: Full GitOps functionality with all three providers
- Sub-Phase 1.4 (Weeks 9-10): Integration, testing, and documentation
  - Milestone: Production-ready Phase 1 with >70% test coverage

**Phase 2** (Weeks 11-14): CLI integration and migration
- Milestone: CLI users can use new architecture transparently

**Phase 3** (Weeks 15-16): GitHub provider
- Milestone: Users can choose GitHub instead of Gitea

**Phase 4** (Weeks 17-18): GitLab provider
- Milestone: Three Git provider options available

**Phase 5** (Weeks 19-20): Envoy Gateway
- Milestone: Multiple gateway providers supported

**Phase 6** (Weeks 21-22): Istio Gateway
- Milestone: Service mesh integration available

**Phase 7** (Weeks 23-24): Flux provider
- Milestone: Multiple GitOps providers supported

**Phase 8** (Weeks 25-30): Production features & stabilization
- Milestone: v1.0.0 release candidate
   - Display endpoints and credentials
   - Show ArgoCD application health
   - Improve error messages

**Deliverables**:
- Updated CLI with backward compatibility
- idpbuilder Helm chart
- Controller container image
- Updated CLI documentation

### Phase 5: Migration & Compatibility 

**Objective**: Ensure smooth migration path and backward compatibility

#### Tasks:

1. **Migration Tool**
   ```bash
   idpbuilder migrate --cluster-name localdev
   ```
   - Detect existing Localbuild CR
   - Extract configuration
   - Create equivalent Platform CR
   - Migrate without downtime
   - Rollback capability

2. **Backward Compatibility Layer**
   - Keep LocalbuildReconciler functional (deprecated)
   - Add deprecation warnings
   - Support both architectures simultaneously
   - Document migration path

3. **Version Detection**
   - Detect cluster architecture version
   - Auto-select appropriate reconciliation path
   - Prevent mixing incompatible versions
   - Clear error messages for version mismatches

4. **Documentation**
   - Migration guide
   - Architectural comparison document
   - Troubleshooting guide
   - FAQ for common issues

**Deliverables**:
- Migration tool
- Backward compatibility maintained
- Comprehensive migration documentation
- Deprecation timeline

### Phase 6: Advanced Features 

**Objective**: Add production-ready features and extensions

#### Tasks:

1. **Multi-Cluster Support**
   - Support vCluster as infrastructure provider
   - Support Cluster API
   - Remote cluster management
   - Cluster inventory tracking

2. **High Availability**
   - Support multiple replicas for components
   - Leader election for controllers
   - Database persistence for Gitea
   - ArgoCD HA configuration

3. **Monitoring & Observability**
   - Prometheus metrics for controllers
   - Component health dashboards
   - Alert rules for component failures
   - OpenTelemetry integration

4. **Security Enhancements**
   - RBAC for component CRs
   - Secret management improvements
   - TLS everywhere
   - Pod security standards

5. **Package Ecosystem**
   - Package catalog / marketplace
   - Package versioning
   - Package dependencies graph
   - Package discovery

**Deliverables**:
- Multi-cluster support
- HA deployment options
- Monitoring dashboards
- Security hardening
- Package ecosystem foundation

### Phase 7: Testing & Stabilization 

**Objective**: Comprehensive testing and stabilization for production use

#### Tasks:

1. **Testing**
   - E2E test coverage >80%
   - Chaos testing (component failures, network issues)
   - Performance testing (large scale deployments)
   - Upgrade/downgrade testing
   - Multi-platform testing (Linux, macOS, Windows)

2. **Documentation**
   - Complete API reference
   - Architecture deep-dives
   - Operator guide
   - Developer guide
   - Troubleshooting runbooks

3. **Release Preparation**
   - Semantic versioning strategy
   - Changelog generation
   - Release notes
   - Upgrade compatibility matrix
   - Support policy

**Deliverables**:
- Production-ready release
- Complete documentation
- Test coverage >80%
- Release artifacts

## Migration Strategy

### For Existing Users

#### Option 1: In-Place Migration (Recommended)

For users with existing clusters:

```bash
# 1. Backup existing configuration
idpbuilder get config > backup-config.yaml

# 2. Upgrade idpbuilder binary
brew upgrade idpbuilder

# 3. Run migration
idpbuilder migrate --cluster-name localdev --auto-approve

# 4. Verify migration
kubectl get platform -n idpbuilder-system
kubectl get argocdcomponent,giteacomponent,nginxcomponent
```

The migration process:
1. Analyzes existing Localbuild CR
2. Creates corresponding Platform and Component CRs
3. Installs controller manager
4. Transitions management to controllers
5. Marks Localbuild CR as deprecated
6. Validates all services remain available

#### Option 2: Recreate Cluster

For a clean start:

```bash
# 1. Export applications and data
idpbuilder backup --output ./backup

# 2. Delete existing cluster
idpbuilder delete --cluster-name localdev

# 3. Create new cluster with updated binary
idpbuilder create --restore-from ./backup
```

### Breaking Changes

The following changes will require user action:

1. **API Version Bump**: `v1alpha1` → `v1alpha2` (or `v1beta1`)
   - Old CRDs will continue to work during deprecation period
   - Conversion webhooks provided for automatic migration

2. **Controller Deployment**: Controllers now run in-cluster
   - Users relying on CLI-only operation need to adapt
   - Minimal changes for standard use cases

3. **Configuration Structure**: Platform CR replaces Localbuild CR
   - Automated migration available
   - Manual migration documented

### Deprecation Timeline

- **v0.8.0 (Phase 1, Months 1-3)**: New architecture introduced with core providers (Gitea, Nginx, ArgoCD) implemented iteratively
  - Sub-phases allow early adoption and feedback
  - Each sub-phase can be merged independently
- **v0.9.0 (Phase 2, Month 4)**: CLI integration and migration tool added, old architecture marked deprecated with warnings
- **v0.10.0 (Phases 3-4, Months 5-6)**: Alternative Git providers (GitHub, GitLab) added
- **v0.11.0 (Phases 5-6, Months 7-8)**: Alternative gateways (Envoy, Istio) added, old LocalbuildReconciler removed (migration tool still available)
- **v0.12.0 (Phase 7, Month 9)**: Flux provider added
- **v1.0.0 (Phase 8, Months 10-12)**: Production features, stabilization, and first stable release with full provider ecosystem

## Benefits & Impact

### Benefits

1. **Kubernetes-Native**
   - Everything manageable via kubectl
   - Full GitOps compatibility
   - Standard Kubernetes patterns (CRs, controllers, reconciliation)

2. **Operational Excellence**
   - Better observability (conditions, events, metrics)
   - Easier debugging (kubectl describe, logs)
   - Standard troubleshooting approaches

3. **Flexibility**
   - Easy component customization
   - Support for alternative components
   - Plugin architecture possible

4. **Production-Ready**
   - HA configurations supported
   - Proper separation of concerns
   - Infrastructure-agnostic

5. **Extensibility**
   - Third-party controllers can integrate
   - Package ecosystem enablement
   - Community contributions easier

### Impact on Users

#### Developers (IDP Consumers)

- **Minimal impact**: Day-to-day usage unchanged
- **Benefit**: Better reliability and easier troubleshooting
- **Action needed**: None (transparent migration)

#### Platform Engineers (IDP Operators)

- **Moderate impact**: Need to understand new architecture
- **Benefit**: Much more control and customization capability
- **Action needed**: Learn new CRs and controller concepts

#### Contributors

- **High impact**: Significant code restructuring
- **Benefit**: Cleaner architecture, easier to contribute
- **Action needed**: Understand new controller patterns

## Risks & Mitigation

### Risk: Increased Complexity

**Description**: Controller-based architecture adds operational complexity.

**Mitigation**:
- Maintain simple CLI experience for basic use cases
- Provide migration automation
- Comprehensive documentation and examples
- Pre-configured defaults for common scenarios

### Risk: Migration Challenges

**Description**: Users may face issues migrating existing clusters.

**Mitigation**:
- Automated migration tool with rollback
- Extended deprecation period (6+ months)
- Side-by-side architecture support
- Dedicated migration documentation and support

### Risk: Performance Concerns

**Description**: Additional controllers consume more resources.

**Mitigation**:
- Optimize controller resource requests
- Implement leader election for single-active instances
- Provide resource tuning guidance
- Benchmark against current implementation

### Risk: Breaking Existing Workflows

**Description**: Users with automation may break.

**Mitigation**:
- Maintain CLI backward compatibility
- Version detection and auto-adaptation
- Clear upgrade guides
- Communication via release notes and deprecation warnings

## Success Criteria

### Functional Criteria

- [ ] Platform CR successfully orchestrates component installation
- [ ] All component controllers (ArgoCD, Gitea, Nginx) fully functional
- [ ] GitOps hand-off working (ArgoCD manages components)
- [ ] CLI backward compatibility maintained
- [ ] CLI successfully deploys controllers in development mode
- [ ] Controllers work when deployed via GitOps (no CLI)
- [ ] Migration tool successfully converts existing clusters
- [ ] Package system works with new architecture
- [ ] Both deployment modes (CLI-driven and GitOps-driven) validated

### Quality Criteria

- [ ] Test coverage >70% (unit + integration)
- [ ] E2E tests passing for all major scenarios
- [ ] E2E tests for both CLI-driven and GitOps-driven modes
- [ ] Documentation complete (API reference, guides, runbooks)
- [ ] Production installation guide complete
- [ ] Performance parity or better than current implementation
- [ ] Zero critical bugs in beta testing

### Adoption Criteria

- [ ] 50+ users successfully migrate to new architecture
- [ ] Production deployments using GitOps mode validated
- [ ] Development workflows using CLI mode validated
- [ ] Positive community feedback
- [ ] No major blockers reported
- [ ] Third-party integrations demonstrated
- [ ] Controllers demonstrated working on multiple Kubernetes distributions

## Open Questions

1. **Helm vs. Kustomize**: Should we support both installation methods?
   - Recommendation: Start with Helm, add Kustomize later if requested

2. **Controller Deployment**: Should controllers be optional for basic use cases?
   - Recommendation: Always deploy controllers, but make them lightweight

3. **Component Scope**: Should we add more components (e.g., Crossplane)?
   - Recommendation: Start with core three, make extensible for additions

4. **API Versioning**: v1alpha2 vs. v1beta1?
   - Recommendation: v1alpha2 to indicate significant change, move to v1beta1 after stabilization

5. **Multi-tenancy**: Should we support multiple Platform CRs in one cluster?
   - Recommendation: Yes, but document single-platform-per-cluster as primary use case

6. **State Management**: How to handle component state during migrations?
   - Recommendation: GitRepository CRs maintain state, components are stateless where possible

## Appendix

### A. Example End-to-End Flows

#### Mode 1: CLI-Driven Development Workflow

```bash
# Developer uses the CLI for quick local setup
idpbuilder create --name localdev

# Behind the scenes - CLI responsibilities:
# 1. CLI creates Kind cluster
# 2. CLI deploys idpbuilder controllers to cluster (Helm or manifests)
# 3. CLI waits for controller manager to be ready
# 4. CLI creates provider CRs based on defaults:

cat <<EOF | kubectl apply -f -
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GiteaProvider
metadata:
  name: gitea
  namespace: idpbuilder-system
spec:
  namespace: gitea
  adminUser:
    autoGenerate: true
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: NginxGateway
metadata:
  name: nginx
  namespace: idpbuilder-system
spec:
  namespace: ingress-nginx
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: ArgoCDProvider
metadata:
  name: argocd
  namespace: idpbuilder-system
spec:
  namespace: argocd
  adminCredentials:
    autoGenerate: true
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: localdev
  namespace: idpbuilder-system
spec:
  domain: cnoe.localtest.me
  components:
    gitProviders:
      - name: gitea
        kind: GiteaProvider
        namespace: idpbuilder-system
    gateways:
      - name: nginx
        kind: NginxGateway
        namespace: idpbuilder-system
    gitOpsProviders:
      - name: argocd
        kind: ArgoCDProvider
        namespace: idpbuilder-system
EOF

# Behind the scenes - Controller responsibilities:
# 5. GiteaProviderReconciler (running in cluster) installs Gitea via Helm
# 6. GiteaProviderReconciler creates admin user and organizations
# 7. GiteaProviderReconciler updates status with duck-typed fields
# 8. NginxGatewayReconciler (running in cluster) installs Nginx Ingress
# 9. NginxGatewayReconciler creates IngressClass resource
# 10. NginxGatewayReconciler updates status with duck-typed fields
# 11. ArgoCDProviderReconciler (running in cluster) installs ArgoCD
# 12. ArgoCDProviderReconciler creates projects and admin credentials
# 13. ArgoCDProviderReconciler updates status with duck-typed fields
# 14. PlatformReconciler sees all providers are Ready
# 15. PlatformReconciler creates GitRepository CRs for bootstrap content
# 16. GitRepositoryReconciler uses GiteaProvider duck-typed interface
# 17. GitRepositoryReconciler creates repos with embedded content in Gitea
# 18. PlatformReconciler creates ArgoCD Applications referencing repos
# 19. ArgoCD syncs applications from Gitea repositories
# 20. PlatformReconciler updates Platform status to Ready

# Behind the scenes - CLI responsibilities (continued):
# 21. CLI monitors Platform status until Ready
# 22. CLI retrieves endpoints and credentials from provider status
# 23. CLI displays success message and access information to user

# Output shown to user:
# ✓ Cluster created successfully
# ✓ Controllers deployed
# ✓ Platform ready
# 
# Access your platform:
#   Gitea:  https://gitea.cnoe.localtest.me
#   ArgoCD: https://argocd.cnoe.localtest.me
#   Admin credentials: kubectl get secret -n idpbuilder-system
```

#### Mode 2: GitOps-Driven Production Workflow

```bash
# Platform team installs to existing production cluster
# No CLI required - everything is declarative

# Step 1: Install idpbuilder controllers (done once per cluster)
helm repo add idpbuilder https://cnoe-io.github.io/idpbuilder
helm install idpbuilder-controllers idpbuilder/idpbuilder-controllers \
  --namespace idpbuilder-system \
  --create-namespace \
  --values values-prod.yaml

# Or using static manifests:
kubectl apply -f https://github.com/cnoe-io/idpbuilder/releases/latest/download/install.yaml

# Step 2: Add Platform and Provider CRs to GitOps repo
# In your GitOps repository (managed by ArgoCD/Flux):

# File: platform/providers/gitea.yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GiteaProvider
metadata:
  name: gitea-prod
  namespace: idpbuilder-system
spec:
  namespace: gitea
  version: 1.24.3
  config:
    persistence:
      enabled: true
      size: 100Gi
      storageClass: fast-ssd
    postgresql:
      enabled: true
  adminUser:
    username: admin
    email: platform-team@company.com
    passwordSecretRef:
      name: gitea-admin-secret
      namespace: gitea
  organizations:
    - name: platform-team

---
# File: platform/providers/nginx.yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: NginxGateway
metadata:
  name: nginx-prod
  namespace: idpbuilder-system
spec:
  namespace: ingress-nginx
  version: 1.13.0
  config:
    controller:
      replicaCount: 3
      service:
        type: LoadBalancer

---
# File: platform/providers/argocd.yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: ArgoCDProvider
metadata:
  name: argocd-prod
  namespace: idpbuilder-system
spec:
  namespace: argocd
  version: v2.12.0
  config:
    server:
      replicas: 3
    controller:
      replicas: 3

---
# File: platform/platform.yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: production
  namespace: idpbuilder-system
spec:
  domain: idp.company.com
  components:
    gitProviders:
      - name: gitea-prod
        kind: GiteaProvider
        namespace: idpbuilder-system
    gateways:
      - name: nginx-prod
        kind: NginxGateway
        namespace: idpbuilder-system
    gitOpsProviders:
      - name: argocd-prod
        kind: ArgoCDProvider
        namespace: idpbuilder-system

# Step 3: Commit and push to GitOps repo
git add platform/
git commit -m "Deploy idpbuilder platform"
git push

# Step 4: ArgoCD/Flux syncs the changes
# Controllers (already running in cluster) see new CRs and reconcile
# Same reconciliation logic as CLI mode - controllers are identical

# Step 5: Monitor via kubectl (no CLI needed)
kubectl get platform -n idpbuilder-system
kubectl get giteaprovider,nginxgateway,argocdprovider -n idpbuilder-system
kubectl describe platform production -n idpbuilder-system

# All operations are declarative and auditable through Git
# No CLI binary required on production systems
# Platform team manages everything through GitOps workflow
```

#### Key Differences Between Modes

| Aspect | CLI-Driven (Development) | GitOps-Driven (Production) |
|--------|-------------------------|---------------------------|
| **Infrastructure** | CLI provisions Kind | Pre-existing cluster |
| **Controller Install** | CLI deploys via Helm/manifests | Helm/kubectl/GitOps |
| **CR Creation** | CLI creates based on flags | GitOps repository |
| **Workflow** | Single command (`idpbuilder create`) | Declarative Git commits |
| **Use Case** | Local dev, testing, demos | Production, staging, teams |
| **Prerequisites** | Docker | Kubernetes cluster |
| **Auditability** | CLI logs | Git history |
| **Scalability** | Single user | Multi-cluster, multi-team |

**Important**: Controllers are identical in both modes. The separation is only in how infrastructure is provisioned and how CRs are created initially. Once controllers are running, they operate the same way regardless of deployment mode.

### B. Component Interaction Diagram

```
┌───────────────────────────────────────────────────────────────────┐
│                          Platform CR                               │
│  Spec: References to provider CRs                                 │
│  Status: Aggregated health of all providers                       │
└───────────────┬───────────────────────────────────────────────────┘
                │ (references)
                ├──────────────┬────────────────┬──────────────┐
                │              │                │              │
                ▼              ▼                ▼              ▼
     ┌──────────────┐  ┌─────────────┐  ┌──────────────┐  ┌─────────┐
     │ GiteaProvider│  │ NginxGateway│  │ ArgoCDProvider│  │ Package │
     │      CR      │  │     CR      │  │      CR       │  │   CRs   │
     └──────┬───────┘  └──────┬──────┘  └──────┬───────┘  └────┬────┘
            │                 │                 │               │
            │ (manages)       │ (manages)       │ (manages)     │
            ▼                 ▼                 ▼               │
      ┌─────────┐       ┌──────────┐      ┌─────────┐         │
      │  Gitea  │       │ Ingress  │      │ ArgoCD  │         │
      │ Server  │       │  Nginx   │      │ Server  │         │
      │  Pods   │       │  Pods    │      │  Pods   │         │
      └────┬────┘       └──────────┘      └────┬────┘         │
           │                                    │              │
           │ (hosts)                            │ (manages)    │
           ▼                                    └──────────────┘
      ┌─────────┐                                     │
      │   Git   │◄────────────────────────────────────┘
      │  Repos  │              (syncs from)
      └─────────┘

Duck-Typed Interfaces:
- Git Providers: endpoint, internalEndpoint, credentialsSecretRef
- Gateway Providers: ingressClassName, loadBalancerEndpoint, internalEndpoint
- GitOps Providers: endpoint, internalEndpoint, credentialsSecretRef

Other controllers (GitRepository, Package) use duck-typing to access
any provider implementation without tight coupling.
```

### C. Resource Naming Conventions

- **Platform CR**: `<cluster-name>` (e.g., `localdev`)
- **Provider CRs**: 
  - Git Providers: `<name>` (e.g., `gitea`, `github-prod`, `gitlab-dev`)
  - Gateway Providers: `<name>-gateway` (e.g., `nginx-gateway`, `envoy-gateway`)
  - GitOps Providers: `<name>` (e.g., `argocd`, `flux`)
- **Namespace for controllers**: `idpbuilder-system`
- **Provider deployment namespaces**: Provider-specific (e.g., `argocd`, `gitea`, `ingress-nginx`, `flux-system`)
- **GitRepository CRs**: `<component>-bootstrap` (e.g., `argocd-bootstrap`)
- **Package CRs**: User-defined (e.g., `backstage`, `crossplane`)

### D. API Group Versioning

- **Current**: `idpbuilder.cnoe.io/v1alpha1`
- **Proposed**: `idpbuilder.cnoe.io/v1alpha2` (new architecture)
- **Future**: `idpbuilder.cnoe.io/v1beta1` (after stabilization)
- **Target**: `idpbuilder.cnoe.io/v1` (GA release)

### E. Controller Permissions (RBAC)

Each controller requires specific permissions:

**PlatformReconciler**:
- Full access to Platform CRs
- Read access to all provider CRs (duck-typed via unstructured)
- Read status from all providers

**Git Provider Reconcilers**:
- **GiteaProviderReconciler**:
  - Full access to GiteaProvider CRs
  - Create/Update/Delete Gitea namespaces
  - Create secrets for admin credentials
  - HTTP access to Gitea API (via ServiceAccount)
- **GitHubProviderReconciler**:
  - Full access to GitHubProvider CRs
  - Read GitHub credentials from secrets
  - HTTP access to GitHub API
- **GitLabProviderReconciler**:
  - Full access to GitLabProvider CRs
  - Read GitLab credentials from secrets
  - HTTP access to GitLab API

**Gateway Provider Reconcilers**:
- **NginxGatewayReconciler**:
  - Full access to NginxGateway CRs
  - Create/Update/Delete ingress-nginx namespaces
  - Manage IngressClass resources
  - Manage TLS secrets
  - ValidatingWebhookConfiguration access
- **EnvoyGatewayReconciler**:
  - Full access to EnvoyGateway CRs
  - Create/Update/Delete envoy-gateway namespaces
  - Manage GatewayClass and Gateway resources
- **IstioGatewayReconciler**:
  - Full access to IstioGateway CRs
  - Create/Update/Delete istio-system namespaces
  - Manage Istio CRDs and Gateway resources

**GitOps Provider Reconcilers**:
- **ArgoCDProviderReconciler**:
  - Full access to ArgoCDProvider CRs
  - Create/Update/Delete ArgoCD namespaces
  - Install ArgoCD CRDs
  - Create secrets for credentials
- **FluxProviderReconciler**:
  - Full access to FluxProvider CRs
  - Create/Update/Delete Flux namespaces
  - Install Flux CRDs
  - Create secrets for credentials

### F. Monitoring & Observability

**Metrics to Export** (Prometheus format):

```
# Platform-level metrics
idpbuilder_platform_components_total{platform="localdev"}
idpbuilder_platform_components_ready{platform="localdev", component="argocd"}
idpbuilder_platform_reconcile_duration_seconds{platform="localdev"}
idpbuilder_platform_reconcile_errors_total{platform="localdev"}

# Component-level metrics
idpbuilder_component_install_duration_seconds{component="argocd", version="v2.12.0"}
idpbuilder_component_ready{component="argocd", phase="Ready"}
idpbuilder_component_helm_releases_total{component="argocd"}
idpbuilder_component_helm_failures_total{component="argocd"}

# GitRepository metrics
idpbuilder_gitrepository_syncs_total{name="argocd-bootstrap"}
idpbuilder_gitrepository_sync_failures_total{name="argocd-bootstrap"}
idpbuilder_gitrepository_last_sync_timestamp{name="argocd-bootstrap"}
```

**Events to Emit**:

- PlatformCreated
- ComponentInstallStarted
- ComponentInstallCompleted
- ComponentInstallFailed
- ComponentUpgradeStarted
- ComponentHealthCheckPassed
- ComponentHealthCheckFailed
- GitRepositoryCreated
- GitRepositorySyncCompleted
- PackageDeployed

### G. Configuration Examples

**Minimal Configuration** (Development with Gitea and Nginx):

First, create the provider CRs:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GiteaProvider
metadata:
  name: gitea-local
  namespace: idpbuilder-system
spec:
  namespace: gitea
  version: 1.24.3
  adminUser:
    autoGenerate: true
  organizations:
    - name: idpbuilder
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: NginxGateway
metadata:
  name: nginx-gateway
  namespace: idpbuilder-system
spec:
  namespace: ingress-nginx
  version: 1.13.0
```

Then reference them in the Platform:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: dev
  namespace: idpbuilder-system
spec:
  domain: cnoe.localtest.me
  components:
    gitProviders:
      - name: gitea-local
        kind: GiteaProvider
        namespace: idpbuilder-system
    gateways:

      - name: nginx-gateway
      kind: NginxGateway
      namespace: idpbuilder-system
```

**GitHub + Envoy Gateway Configuration** (External Git with modern gateway):

First, create the provider CRs:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GitHubProvider
metadata:
  name: github-external
  namespace: idpbuilder-system
spec:
  organization: my-company
  endpoint: https://api.github.com
  credentialsSecretRef:
    name: github-token
    namespace: idpbuilder-system
    key: token
  repositoryDefaults:
    visibility: private
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: EnvoyGateway
metadata:
  name: envoy-gateway
  namespace: idpbuilder-system
spec:
  namespace: envoy-gateway-system
  version: v1.0.0
```

Then reference them in the Platform:

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: dev-external
  namespace: idpbuilder-system
spec:
  domain: cnoe.localtest.me
  
  components:
    argocd:
      enabled: true
    
    gitProviders:
      - name: github-external
        kind: GitHubProvider
        namespace: idpbuilder-system
    
    gateways:

    
      - name: envoy-gateway
      kind: EnvoyGateway
      namespace: idpbuilder-system
```

**Multi-Provider Configuration** (Multiple Git providers):

```yaml
# Define multiple Git providers
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GiteaProvider
metadata:
  name: gitea-dev
  namespace: idpbuilder-system
spec:
  namespace: gitea
  adminUser:
    autoGenerate: true
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: GitHubProvider
metadata:
  name: github-prod
  namespace: idpbuilder-system
spec:
  organization: my-company
  credentialsSecretRef:
    name: github-prod-token
    namespace: idpbuilder-system
    key: token
---
# Platform references both
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: hybrid
  namespace: idpbuilder-system
spec:
  domain: cnoe.localtest.me
  components:
    # Use multiple Git providers
    gitProviders:
      - name: gitea-dev
        kind: GiteaProvider
        namespace: idpbuilder-system
      - name: github-prod
        kind: GitHubProvider
        namespace: idpbuilder-system
    
    gateways:

    
      - name: nginx-gateway
      kind: NginxGateway
      namespace: idpbuilder-system
```



**Multi-Gateway Configuration** (Multiple ingress controllers for different purposes):

```yaml
# Define multiple gateway providers
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: NginxGateway
metadata:
  name: nginx-public
  namespace: idpbuilder-system
spec:
  namespace: ingress-nginx-public
  version: 1.13.0
  serviceType: LoadBalancer
  class: nginx-public
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: EnvoyGateway
metadata:
  name: envoy-internal
  namespace: idpbuilder-system
spec:
  namespace: envoy-gateway-internal
  version: v1.0.0
  serviceType: ClusterIP
  class: envoy-internal
---
# Platform references both gateways
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: multi-gateway
  namespace: idpbuilder-system
spec:
  domain: cnoe.localtest.me
  components:
    gitProviders:
      - name: gitea-dev
        kind: GiteaProvider
        namespace: idpbuilder-system
    
    # Use multiple gateway providers
    gateways:
      - name: nginx-public
        kind: NginxGateway
        namespace: idpbuilder-system
      - name: envoy-internal
        kind: EnvoyGateway
        namespace: idpbuilder-system
```

In this setup:
- Nginx handles public-facing services (external LoadBalancer)
- Envoy handles internal services (ClusterIP, service mesh)
- Platform components can choose which gateway to use
- Different ingress classes allow routing to different controllers



**Multi-GitOps Provider Configuration** (ArgoCD for applications, Flux for infrastructure):

```yaml
# Define multiple GitOps providers
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: ArgoCDProvider
metadata:
  name: argocd-apps
  namespace: idpbuilder-system
spec:
  namespace: argocd
  version: v2.12.0
  adminCredentials:
    autoGenerate: true
  projects:
    - name: applications
      description: User applications
---
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: FluxProvider
metadata:
  name: flux-infra
  namespace: idpbuilder-system
spec:
  namespace: flux-system
  version: v2.4.0
  config:
    sourceController:
      resources:
        requests:
          cpu: 100m
          memory: 256Mi
---
# Platform references both GitOps providers
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: multi-gitops
  namespace: idpbuilder-system
spec:
  domain: cnoe.localtest.me
  components:
    gitProviders:
      - name: gitea-dev
        kind: GiteaProvider
        namespace: idpbuilder-system
    
    gateways:
      - name: nginx-gateway
        kind: NginxGateway
        namespace: idpbuilder-system
    
    # Use multiple GitOps providers
    gitOpsProviders:
      - name: argocd-apps
        kind: ArgoCDProvider
        namespace: idpbuilder-system
      - name: flux-infra
        kind: FluxProvider
        namespace: idpbuilder-system
```

In this setup:
- ArgoCD manages application deployments
- Flux manages infrastructure and platform components
- Each GitOps provider operates independently
- Different teams can use different GitOps tools

**Production Configuration** (High Availability):

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Platform
metadata:
  name: production
  namespace: idpbuilder-system
spec:
  domain: idp.example.com
  
  ingressConfig:
    provider: nginx
    usePathRouting: false
    tlsSecretRef:
      name: wildcard-tls
      namespace: idpbuilder-system
  
  components:
    argocd:
      enabled: true
      namespace: argocd
      version: v2.12.0
      helmChart:
        repository: https://argoproj.github.io/argo-helm
        version: 7.0.0
      values:
        server:
          replicas: 3
          resources:
            limits:
              cpu: 2
              memory: 2Gi
        controller:
          replicas: 3
        repoServer:
          replicas: 3
        redis-ha:
          enabled: true
        notifications:
          enabled: true
        dex:
          enabled: true
    
    gitea:
      enabled: true
      namespace: gitea
      values:
        persistence:
          enabled: true
          size: 100Gi
          storageClass: fast-ssd
        postgresql:
          enabled: true
          primary:
            persistence:
              size: 50Gi
        redis:
          enabled: true
        replicas: 3
    
    nginx:
      enabled: true
      namespace: ingress-nginx
      values:
        controller:
          replicaCount: 3
          resources:
            limits:
              cpu: 1
              memory: 1Gi
          service:
            type: LoadBalancer
```

---

## Conclusion

This architectural evolution represents a significant maturity milestone for idpbuilder. By embracing Kubernetes-native patterns and separating infrastructure concerns from application management, we enable idpbuilder to serve both development and production use cases effectively.

The migration path preserves backward compatibility while providing a clear upgrade path. The phased implementation plan minimizes risk and allows for iterative refinement based on user feedback.

We believe this architecture positions idpbuilder as a truly production-ready platform that can scale from local development to multi-cluster production environments while maintaining its core value proposition: simplicity and ease of use.

---

**Document Approval**:
- [ ] Technical Lead Review
- [ ] Product Owner Review  
- [ ] Community Feedback Period (2 weeks)
- [ ] Final Approval

**Next Steps**:
1. Share with community for feedback
2. Refine based on feedback
3. Create detailed GitHub issues for each phase
4. Begin Phase 1 implementation
