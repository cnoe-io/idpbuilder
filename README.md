Table of Contents
=================

* [IDP Builder](#idp-builder)
  * [About](#about)
  * [Quickstart](#quickstart)
    * [Build](#build)
    * [Run](#run)
    * [Use](#use)
  * [Architecture](#architecture)
* [Extending the IDP builder](#extending-the-idpbuilder)
* [Developer notes](#developer-notes)

# IDP Builder

Internal development platform binary launcher.

> **WORK IN PROGRESS**: This tool is in a pre-release stage and is under active development.

## About

Spin up a complete internal developer platform using industry standard technologies like Kubernetes, Argo, and backstage with only Docker required as a dependency.

This can be useful in several ways:
* Create a single binary which can demonstrate an IDP reference implementation.
* Use within CI to perform integration testing.
* Use as a local development environment for IDP engineers.

## Quickstart:

### Build

`make`

### Run

`./idpbuilder create --build-name localdev`

You can also define the kubernetes version to image and which corresponds to the kind pre-built [image](https://github.com/kubernetes-sigs/kind/releases).
`./idpbuilder create --kube-version v1.27.3`

If it is needed to expose some extra Ports between the docker container and the kubernetes host, they can be declared as such
`./idpbuilder create --extra-ports 22:32222`

It is also possible to use your own kind config file
`./idpbuilder create --build-name local --kind-config ./my-kind.yaml`

**NOTE**: Be sure to include in your kind config the section `containerdConfigPatches` where the registry hostname includes the name specified with the parameter: `--build-name`
```yaml
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5001"]
    endpoint = ["http://idpbuilder-<localBuildName>-registry:5000"]
```

### Use

#### GUI

GUIs for core packages are available at the following addresses:  

* ArgoCD: https://argocd.cnoe.localtest.me:8443/
* Backstage: https://backstage.cnoe.localtest.me:8443/
* Gitea: https://gitea.cnoe.localtest.me:8443/


## Architecture

idpbuilder is made of two phases: CLI and Kubernetes controllers.

![idpbuilder.png](docs/images/idpbuilder.png)

### CLI

When the idpbuilder binary is executed, it starts with the CLI phase.

1. This is the phase where command flags are parsed and translated into relevant Go structs' fields. Most notably the [`LocalBuild`](https://github.com/cnoe-io/idpbuilder/blob/main/api/v1alpha1/localbuild_types.go) struct.
2. Create a Kind cluster, then update the kubeconfig file.
3. Once the kind cluster is started and relevant fields are populated, Kubernetes controllers are started:
  *  `LocalbuildReconciler` responsible for bootstrapping the cluster with absolute necessary packages. Creates Custom Resources (CRs) and installs embedded manifests.
  *  `RepositoryReconciler` responsible for creating and managing Gitea repository and repository contents.
  *  `CustomPackageReconciler` responsible for managing custom packages.
4. They are all managed by a single Kubernetes controller manager.
5. Once controllers are started, CRs corresponding to these controllers are created. For example for Backstage, it creates a GitRepository CR and ArgoCD Application.
6. CLI then waits for these CRs to be ready.

### Controllers

During this phase, controllers act on CRs created by the CLI phase. Resources such as Gitea repositories and ArgoCD applications are created. 

#### LocalbuildReconciler

`LocalbuildReconciler` bootstraps the cluster using embedded manifests. Embedded manifests are yaml files that are baked into the binary at compile time.
1. Install core packages. They are essential services that are needed for the user experiences we want to enable:
  * Gitea. This is the in-cluster Git server that hosts Git repositories.
  * Ingress-nginx. This is necessary to expose services inside the cluster to the users.
  * ArgoCD. This is used as the packaging mechanism. Its primary purpose is to deploy manifests from gitea repositories.
2. Once they are installed, it creates `GitRepository` CRs for core packages. This CR represents the git repository on the Gitea server.
3. Create ArgoCD applications for the apps. Point them to the Gitea repositories. From here on, ArgoCD manages the core packages.

Once core packages are installed, it creates the other embedded applications: Backstage and Crossplane.
1. Create `GitRepository` CRs for the apps.
2. Create ArgoCD applications for the apps. Point them to the Gitea repositories.


#### RepositoryReconciler

`RepositoryReconciler` creates Gitea repositories.
The content of the repositories can either be sourced from Embedded file system or local file system.

#### CustomPackageReconciler

`CustomPackageReconciler` parses the specified ArgoCD application files. If they specify repository URL with the scheme `cnoe://`,
it creates `GitRepository` CR with source specified as local, then creates ArgoCD application with the repository URL replaced.

For example, if an ArgoCD application is specified as the following.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  source:
    repoURL: cnoe://busybox
```

Then, the actual object created is this.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  source:
    repoURL: http://my-gitea-http.gitea.svc.cluster.local:3000/giteaAdmin/idpbuilder-localdev-my-app-busybox.git
```

## Developer notes

If you want to contribute and extend the existing project, make sure that you have installed go (>= 1.20) and cloned this project.
Next, you can build it `make` or launch the `main.go` within your IDE or locally `./idpbuilder`.

You can override the kind configuration generated by the idpbuilder. For that purpose, look to the
console to grab the config and save it in a file:
```text
########################### Our kind config ############################
# Kind kubernetes release images https://github.com/kubernetes-sigs/kind/releases
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
...
#########################   config end    ############################
```
Next, import it `./idpbuilder create --kindConfig <path to the config file>`


## Extending the IDP builder

We are actively working to include more patterns and examples of extending idpbuilder to get started easily.