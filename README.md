Table of Contents
=================

* [IDP Builder](#idp-builder)
  * [About](#about)
  * [Quickstart](#quickstart)
    * [Running the idpbuilder](#running-the-idpbuilder)
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

## Core Components and versions
The idpbuilder ships with a number of core components that are embedded in the binary. Here is the list and the version of those components:

| Name     | Version |
| -------- | ------- |
| Argo CD  | v2.10.7 |
| Gitea    | v9.5.1  |
| Nginx    | v1.8.1  |

## Quickstart

### Download and install the idpbuilder

Download the latest release with the commands:

```bash
version=$(curl -Ls -o /dev/null -w %{url_effective} https://github.com/cnoe-io/idpbuilder/releases/latest)
version=${version##*/}
curl -L -o ./idpbuilder.tar.gz "https://github.com/cnoe-io/idpbuilder/releases/download/${version}/idpbuilder-$(uname | awk '{print tolower($0)}')-$(uname -m | sed 's/x86_64/amd64/').tar.gz"
tar xzf idpbuilder.tar.gz

./idpbuilder version
# example output
# idpbuilder 0.3.0 go1.21.5 linux/amd64
```

Alternatively, you can download the latest binary from [the latest release page](https://github.com/cnoe-io/idpbuilder/releases/latest).

### Using the idpbuilder

```bash
./idpbuilder create
```

This is the most basic command which creates a Kubernetes Cluster (Kind cluster) with the core packages installed.

<details>
  <summary>What are the core packages?</summary>

  * **ArgoCD** is the GitOps solution to deploy manifests to Kubernetes clusters. In this project, a package is an ArgoCD application. 
  * **Gitea** server is the in-cluster Git server that ArgoCD can be configured to sync resources from. You can sync from local file systems to this.
  * **Ingress-nginx** is used as a method to access in-cluster resources such as ArgoCD UI and Gitea UI.

  See the [Architecture](#Architecture) section for more information on further information on how core packages are installed and configured.

</details>


Once idpbuilder finishes provisioning cluster and packages, you can access GUIs by going to the following addresses in your browser.

* ArgoCD: https://argocd.cnoe.localtest.me:8443/
* Gitea: https://gitea.cnoe.localtest.me:8443/

You can obtain credentials for these packages by running the following commands:

```bash
# argoCD password. username is admin.
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o go-template='{{ range $key, $value := .data }}{{ printf "%s: %s\n" $key ($value | base64decode) }}{{ end }}'

# gitea admin credentials
kubectl get secrets -n gitea gitea-admin-secret \
  -o go-template='{{ range $key, $value := .data }}{{ printf "%s: %s\n" $key ($value | base64decode) }}{{ end }}'
```

####  Example commands

**For more advanced use cases, check out the [examples](./examples) directory.**

You can specify the kubernetes version by using the `--kube-version` flag. Supported versions are available [here](https://github.com/kubernetes-sigs/kind/releases).

```
./idpbuilder create --kube-version v1.27.3
```

If you need to expose more ports between the docker container and the kubernetes host, you can use the `--extra-ports` flag. For example:

```
./idpbuilder create --extra-ports 22:32222
```

If you want to specify your own kind configuration file, use the `--kind-config` flag.

```
./idpbuilder create --build-name local --kind-config ./my-kind.yaml`
```

You can also specify the name of build. This name is used as part of the cluster, namespace, and git repositories.

```
./idpbuilder create --build-name localdev 
```

### Custom Packages

Idpbuilder supports specifying custom packages using the flag `--package-dir` flag. This flag expects a directory containing ArgoCD application files.

Let's take a look at [this example](examples/basic). This example defines two custom package directories to deploy to the cluster.

To deploy these packages, run the following commands from this repository's root.

```
./idpbuilder create --package-dir examples/basic/package1  --package-dir examples/basic/package2
```

Running this command should create three additional ArgoCD applications in your cluster.

```sh
$ kubectl get Applications -n argocd  -l example=basic
NAME         SYNC STATUS   HEALTH STATUS
guestbook    Synced        Healthy
guestbook2   Synced        Healthy
my-app       Synced        Healthy
```

Let's break this down. The [first package directory](examples/basic/package1) defines an application. This corresponds to the `my-app` application above. In this application, we want to deploy manifests from local machine in GitOps way.

The directory contains an [ArgoCD application file](examples/basic/package1/app.yaml). This is a normal ArgoCD application file except for one field.

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
spec:
  source:
    repoURL: cnoe://manifests
```

The `cnoe://` prefix in the `repoURL` field indicates that we want to sync from a local directory.
Values after `cnoe://` is treated as a relative path from this file. In this example, we are instructing idpbuilder to make ArgoCD sync from files in the [manifests directory](examples/basic/package1/manifests).

As a result the following actions were taken by idpbuilder: 
1. Create a Gitea repository.
2. Fill the repository with contents from the manifests directory.
3. Update the Application spec to use the newly created repository.

You can verify this by going to this address in your browser: https://gitea.cnoe.localtest.me:8443/giteaAdmin/idpbuilder-localdev-my-app-manifests

![img.png](docs/images/my-app-repo.png)


This is the repository that corresponds to the [manifests](examples/basic/package1/manifests) folder.
It contains a file called `alpine.yaml`, synced from the `manifests` directory above.

You can also view the updated Application spec by going to this address: https://argocd.cnoe.localtest.me:8443/applications/argocd/my-app

![myapp](docs/images/my-app.png)


The second package directory defines two normal ArgoCD applications referencing a remote repository.
They are applied as-is.

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

If you want to contribute and extend the existing project, make sure that you have installed go (>= 1.21) and cloned this project.
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

### Default manifests installed by idpbuilder

The default manifests for the core packages are available [here](pkg/controllers/localbuild/resources). 
These are generated by scripts. If you want to make changes to them, see below.

#### ArgoCD

ArgoCD manifests are generated using a bash script available [here](./hack/argo-cd/generate-manifests.sh). 
This script runs kustomize to modify the basic installation manifests provided by ArgoCD. Modifications include:

1. Prevent notification and dex pods from running. This is done to keep the number of pods running low by default.
2. Use the annotation tracking instead of the default label tracking. Annotation tracking allows you to avoid [problems caused by the label tracking method](https://argo-cd.readthedocs.io/en/stable/user-guide/resource_tracking/). In addition, this configuration is required when using Crossplane.  


## Extending the IDP builder

We are actively working to include more patterns and examples of extending idpbuilder to get started easily.

