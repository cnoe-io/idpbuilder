# Controller-Based Architecture Specification

**Version:** 1.0 Draft  
**Date:** December 19, 2025  
**Status:** Proposal  
**Authors:** IDP Builder Team

## Executive Summary

This document proposes a significant architectural evolution of the idpbuilder tool to transition from a CLI-driven installation model to a controller-based architecture. This change will enable idpbuilder to function as a true Kubernetes-native platform, where infrastructure components and application workloads are managed declaratively through Kubernetes Custom Resources (CRs) and reconciliation loops.

### Goals

1. **Kubernetes-Native Management**: Enable all functionality to be managed through kubectl and GitOps tools like ArgoCD
2. **Separation of Concerns**: Clearly delineate infrastructure provisioning from application/service management
3. **Production Readiness**: Support production workloads and virtualized control planes (e.g., vCluster, Cluster API)
4. **Extensibility**: Allow easier integration of additional services and customization by end users
5. **Operational Excellence**: Improve observability, debugging, and lifecycle management through standard Kubernetes patterns

### Non-Goals

1. Breaking changes to the CLI experience (backward compatibility maintained where feasible)
2. Removing the ability to run idpbuilder as a single binary
3. Supporting non-Kubernetes infrastructure

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
│  CLI/Operator:                                                       │
│    - Provisions Kubernetes cluster (Kind, vCluster, etc.)           │
│    - Installs idpbuilder-controllers (Helm chart or manifest)       │
│    - Creates initial Platform CR                                    │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                 Platform Controllers (On-Cluster)                    │
│                                                                      │
│  PlatformReconciler:                                                 │
│    - Orchestrates platform bootstrap                                 │
│    - References provider CRs (Git, Gateway)                          │
│    - Creates ArgoCD component CR                                     │
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

**Objective**: Implement a working end-to-end platform controller architecture using the existing providers (Gitea, Nginx, ArgoCD) that replicates current idpbuilder functionality.

#### Scope:
- Implement **only** the providers that exist today: Gitea, Nginx, ArgoCD
- Migrate existing LocalbuildReconciler logic to new controller architecture
- Validate the duck-typing pattern works with real implementations
- Achieve feature parity with current CLI-driven installation

#### Tasks:

1. **Core CRD Definitions**
   - Define `Platform` CR (v1alpha2)
   - Define `GiteaProvider` CR
   - Define `NginxGateway` CR
   - Define `ArgoCDProvider` CR
   - Define duck-typed common status fields
   - Generate CRD manifests using controller-gen
   - **Skip**: GitHub, GitLab, Envoy, Istio, Flux CRDs (added later)

2. **Duck-Typing Infrastructure**
   - Create `pkg/util/provider/` package for duck-typed access
   - Implement `GetGitProviderStatus()` function using unstructured access
   - Implement `GetGatewayProviderStatus()` function
   - Implement `GetGitOpsProviderStatus()` function
   - Create shared condition helpers and status utilities

3. **Provider Controllers - Core Three**
   
   **GiteaProviderReconciler**:
   ```go
   // pkg/controllers/gitprovider/gitea_controller.go
   type GiteaProviderReconciler struct {
       client.Client
       Scheme *runtime.Scheme
       HelmClient *helm.Client
       GiteaClientFactory func(baseURL, token string) (*gitea.Client, error)
   }

   func (r *GiteaProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // Migrate logic from current LocalbuildReconciler.reconcileGitea()
       // 1. Fetch GiteaProvider CR
       // 2. Install Gitea via Helm (reuse existing embedded manifests)
       // 3. Create admin user and organization
       // 4. Update status with duck-typed fields
   }
   ```

   **NginxGatewayReconciler**:
   ```go
   // pkg/controllers/gatewayprovider/nginx_controller.go
   type NginxGatewayReconciler struct {
       client.Client
       Scheme *runtime.Scheme
       HelmClient *helm.Client
   }

   func (r *NginxGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // Migrate logic from current LocalbuildReconciler.reconcileNginx()
       // 1. Fetch NginxGateway CR
       // 2. Install Nginx Ingress via Helm (reuse existing manifests)
       // 3. Configure IngressClass
       // 4. Update status with duck-typed fields
   }
   ```

   **ArgoCDProviderReconciler**:
   ```go
   // pkg/controllers/gitopsprovider/argocd_controller.go
   type ArgoCDProviderReconciler struct {
       client.Client
       Scheme *runtime.Scheme
       HelmClient *helm.Client
   }

   func (r *ArgoCDProviderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // Migrate logic from current LocalbuildReconciler.reconcileArgoCD()
       // 1. Fetch ArgoCDProvider CR
       // 2. Install ArgoCD via Helm (reuse existing manifests)
       // 3. Create admin credentials and projects
       // 4. Update status with duck-typed fields
   }
   ```

4. **Enhanced GitRepositoryReconciler**
   ```go
   // pkg/controllers/gitrepository/controller.go
   func (r *GitRepositoryReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // Update to use duck-typed Git provider access
       // 1. Fetch GitRepository CR
       // 2. Get Git provider using GetGitProviderStatus()
       // 3. Create repository using provider credentials
       // 4. Sync content from sources
   }
   ```
   - Update to work with any Git provider via duck-typing
   - Migrate existing Gitea client code
   - Reuse existing content sync logic
   - **Skip**: Complex multi-source merging (added later)

5. **PlatformReconciler Implementation**
   ```go
   // pkg/controllers/platform/controller.go
   type PlatformReconciler struct {
       client.Client
       Scheme *runtime.Scheme
   }

   func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
       // Orchestrate provider lifecycle
       // 1. Fetch Platform CR
       // 2. Ensure provider CRs exist (create if missing)
       // 3. Wait for providers to be Ready
       // 4. Create GitRepository CRs for bootstrap
       // 5. Aggregate status from providers
       // 6. Update Platform status
   }
   ```

6. **Helm Integration**
   - Add Helm SDK dependencies (helm.sh/helm/v3)
   - Create Helm client wrapper in `pkg/util/helm/`
   - Implement chart installation/upgrade/deletion
   - Support embedded charts and remote repositories

7. **Testing Framework**
   - Set up envtest for controller unit tests
   - Create test fixtures for core providers
   - Add integration tests using real Helm charts
   - Establish CI test harness

**Deliverables**:
- ✅ Working Platform + 3 provider CRDs
- ✅ Functional reconcilers for Gitea, Nginx, ArgoCD
- ✅ Duck-typing infrastructure validated
- ✅ Enhanced GitRepository controller
- ✅ Test coverage >60%
- ✅ Feature parity with existing LocalbuildReconciler

**Success Criteria**:
- Can create a Platform CR and have all three components install successfully
- GitRepository CRs work with GiteaProvider via duck-typing
- Status aggregation shows platform health
- Existing embedded manifests reused without changes

### Phase 2: Component Controllers 

**Objective**: Implement individual component controllers with full lifecycle management

#### Tasks:

1. **GitOps Provider Controllers**

   Each GitOps provider has its own dedicated reconciler:

---

### Phase 2: CLI Integration & Migration Path

**Objective**: Update the CLI to use the new controller architecture and provide migration path for existing users.

#### Tasks:

1. **Controller Deployment**
   - Create Helm chart for idpbuilder controllers
     ```
     charts/idpbuilder/
       Chart.yaml
       values.yaml
       templates/
         deployment.yaml
         rbac.yaml
         crds/
     ```
   - Build controller container image
   - Support air-gapped installation with embedded images

2. **Update `idpbuilder create` Command**
   ```go
   // pkg/cmd/create/root.go
   func createPlatform(ctx context.Context, opts *CreateOptions) error {
       // 1. Create Kind cluster (unchanged)
       // 2. Install idpbuilder controllers (Helm chart or static manifests)
       // 3. Create provider CRs (GiteaProvider, NginxGateway, ArgoCDProvider)
       // 4. Create Platform CR referencing providers
       // 5. Wait for Platform to be Ready
       // 6. Display access info
   }
   ```
   - Maintain backward compatibility with existing flags
   - Auto-generate provider CRs from CLI flags
   - Wait for controller reconciliation
   - Display endpoints and credentials as before

3. **Flag Mapping**
   - Map `--package-dir` to Platform spec
   - Map `--custom-package` to Package CRs
   - Map ingress/TLS flags to provider configurations
   - Maintain all existing CLI flags

4. **Migration Command**
   ```bash
   idpbuilder migrate --cluster-name localdev
   ```
   - Detect existing Localbuild CR
   - Extract configuration
   - Install controllers
   - Create equivalent provider CRs + Platform CR
   - Validate migration success
   - Provide rollback option

5. **Backward Compatibility**
   - Keep LocalbuildReconciler functional (deprecated)
   - Add deprecation warnings to logs
   - Support both architectures side-by-side
   - Auto-detect which mode to use

**Deliverables**:
- ✅ Updated CLI maintaining backward compatibility
- ✅ idpbuilder Helm chart for controllers
- ✅ Controller container image in registry
- ✅ Migration tool and documentation
- ✅ Updated CLI help and examples

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

**Phase 1** (Weeks 1-6): Core end-to-end with existing providers
- Milestone: Can deploy Platform CR with Gitea + Nginx + ArgoCD

**Phase 2** (Weeks 7-10): CLI integration and migration
- Milestone: CLI users can use new architecture transparently

**Phase 3** (Weeks 11-12): GitHub provider
- Milestone: Users can choose GitHub instead of Gitea

**Phase 4** (Weeks 13-14): GitLab provider
- Milestone: Three Git provider options available

**Phase 5** (Weeks 15-16): Envoy Gateway
- Milestone: Multiple gateway providers supported

**Phase 6** (Weeks 17-18): Istio Gateway
- Milestone: Service mesh integration available

**Phase 7** (Weeks 19-20): Flux provider
- Milestone: Multiple GitOps providers supported

**Phase 8** (Weeks 21-26): Production features & stabilization
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

- **v0.8.0 (Phase 1-2, Months 1-3)**: New architecture introduced with core providers (Gitea, Nginx, ArgoCD), CLI integration, and migration tool
- **v0.9.0 (Phase 3-4, Months 4-5)**: Alternative Git providers (GitHub, GitLab) added, old architecture marked deprecated with warnings
- **v0.10.0 (Phase 5-6, Months 6-7)**: Alternative gateways (Envoy, Istio) added
- **v0.11.0 (Phase 7, Month 8)**: Flux provider added, old LocalbuildReconciler removed (migration tool still available)
- **v0.12.0 (Phase 8, Months 9-11)**: Production features and stabilization
- **v1.0.0 (Month 12)**: First stable release with full provider ecosystem

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
- [ ] Migration tool successfully converts existing clusters
- [ ] Package system works with new architecture

### Quality Criteria

- [ ] Test coverage >70% (unit + integration)
- [ ] E2E tests passing for all major scenarios
- [ ] Documentation complete (API reference, guides, runbooks)
- [ ] Performance parity or better than current implementation
- [ ] Zero critical bugs in beta testing

### Adoption Criteria

- [ ] 50+ users successfully migrate to new architecture
- [ ] Positive community feedback
- [ ] No major blockers reported
- [ ] Third-party integrations demonstrated
- [ ] Production deployments validated

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

### A. Example End-to-End Flow

```bash
# User creates a cluster
idpbuilder create --name localdev

# Behind the scenes:
# 1. CLI creates Kind cluster
# 2. CLI installs idpbuilder controllers (Helm chart)
# 3. CLI creates Platform CR:

# First create provider CRs, then reference them in Platform
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

# 4. Provider reconcilers see new provider CRs
# 5. GiteaProviderReconciler installs Gitea via Helm
# 6. GiteaProviderReconciler creates admin user and organizations
# 7. GiteaProviderReconciler updates status with duck-typed fields
# 8. NginxGatewayReconciler installs Nginx Ingress via Helm
# 9. NginxGatewayReconciler creates IngressClass resource
# 10. NginxGatewayReconciler updates status with duck-typed fields
# 11. ArgoCDProviderReconciler installs ArgoCD via Helm
# 12. ArgoCDProviderReconciler creates projects and admin credentials
# 13. ArgoCDProviderReconciler updates status with duck-typed fields
# 14. PlatformReconciler sees all providers are Ready
# 15. PlatformReconciler creates GitRepository CRs for bootstrap
# 16. GitRepositoryReconciler uses GiteaProvider duck-typed interface
# 17. GitRepositoryReconciler creates repos with embedded content
# 18. PlatformReconciler creates ArgoCD Applications using providers
# 19. ArgoCD syncs applications from Gitea
# 20. PlatformReconciler updates Platform status to Ready
# 21. CLI displays success message and access information
```

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
