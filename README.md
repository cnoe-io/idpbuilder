[![Codespell][codespell-badge]][codespell-link]
[![E2E][e2e-badge]][e2e-link]
[![Go Report Card][report-badge]][report-link]
[![Commit Activity][commit-activity-badge]][commit-activity-link]

# IDP Builder

Internal development platform binary launcher.

> **WORK IN PROGRESS**: This tool is in a pre-release stage and is under active development.

## About

Spin up a complete internal developer platform using industry standard technologies like Kubernetes, Argo, and backstage with only Docker required as a dependency.

This can be useful in several ways:
* Create a single binary which can demonstrate an IDP reference implementation.
* Use within CI to perform integration testing.
* Use as a local development environment for IDP engineers.

## Prerequisites

A container engine is needed locally such as:

| Name                                                  | Supported | Remark                                                                                                                              |
|-------------------------------------------------------|-----------|-------------------------------------------------------------------------------------------------------------------------------------|
| [Docker desktop](https://www.docker.com/get-started/) | Yes       |                                                                                                                                     |
| [Podman desktop](https://podman-desktop.io/)          | No        | idpbuilder can create a cluster using podman [rootful](https://docs.podman.io/en/latest/markdown/podman-machine-set.1.html#rootful) | 


**Note**: Set the `DOCKER_HOST` env var property using `podman` to let idpbuilder to talk with the engine (e.g  export DOCKER_HOST="unix:///var/run/docker.sock")

## Quickstart

If you are interested in running idpbuilder in Codespaces through your browser, check out the [Codespaces](#running-in-codespaces) section.

### Download and install the idpbuilder

Download the latest release with the commands:

```bash
version=$(curl -Ls -o /dev/null -w %{url_effective} https://github.com/cnoe-io/idpbuilder/releases/latest)
version=${version##*/}
curl -L -o ./idpbuilder.tar.gz "https://github.com/cnoe-io/idpbuilder/releases/download/${version}/idpbuilder-$(uname | awk '{print tolower($0)}')-$(uname -m | sed 's/x86_64/amd64/').tar.gz"
tar xzf idpbuilder.tar.gz

./idpbuilder version
# example output
# idpbuilder 0.4.1 go1.21.5 linux/amd64
```

Alternatively, you can download the latest binary from [the latest release page](https://github.com/cnoe-io/idpbuilder/releases/latest).

## Using the idpbuilder

### Basic usage

The most basic command which creates a Kubernetes Cluster (Kind cluster) with the core packages installed.

```bash
./idpbuilder create
```

<details>
  <summary>What are the core packages?</summary>

  * **ArgoCD** is the GitOps solution to deploy manifests to Kubernetes clusters. In this project, a package is an ArgoCD application. 
  * **Gitea** server is the in-cluster Git server that ArgoCD can be configured to sync resources from. You can sync from local file systems to this.
  * **Ingress-nginx** is used as a method to access in-cluster resources such as ArgoCD UI and Gitea UI.

    #### Core package versions
    
    | Name     | Version |
    | -------- | ------- |
    | Argo CD  | v2.10.7 |
    | Gitea    | v9.5.1  |
    | Nginx    | v1.8.1  |

  The default manifests for the core packages are available [here](pkg/controllers/localbuild/resources).
  See the [contribution doc](./CONTRIBUTING.md) for more information on how core packages are installed and configured.

</details>


Once idpbuilder finishes provisioning cluster and packages, you can access GUIs by going to the following addresses in your browser.

* ArgoCD: https://argocd.cnoe.localtest.me:8443/
* Gitea: https://gitea.cnoe.localtest.me:8443/

You can obtain credentials for them by running the following command:

```bash
./idpbuilder get secrets
```

<details>
  <summary> The "get secrets" command </summary>

  The `get secrets` command retrieves the following:
  - ArgoCD initial admin password.
  - Gitea admin user credentials.
  -  Any secrets labeled with `cnoe.io/cli-secret=true`.

  You can think of the command as executing the following kubectl commands:

  ```bash
  kubectl -n argocd get secret argocd-initial-admin-secret
  kubectl get secrets -n gitea gitea-admin-secret
  kubectl get secrets -A -l cnoe.io/cli-secret=true
  ```
  In addition, secrets labeled with `cnoe.io/package-name` can be specified with the `-p` flag. For example, for Gitea:

  ```bash
  ./idpbuilder get secrets -p gitea
  ```

</details>

###  Example commands

**For more advanced use cases, check out the [examples](./examples) directory.**

You can specify the kubernetes version by using the `--kube-version` flag. Supported versions are available [here](https://github.com/kubernetes-sigs/kind/releases).

```
./idpbuilder create --kube-version v1.27.3
```

If you want to specify your own kind configuration file, use the `--kind-config` flag.

```
./idpbuilder create --build-name local --kind-config ./my-kind.yaml`
```

If you want to specify ArgoCD configmap.

```
./idpbuilder create --package-custom-file=argocd:pkg/k8s/test-resources/input/argocd-cm.yaml
```

Run the following commands for available flags and subcommands:

```
./idpbuilder --help
./idpbuilder create --help
```

### Custom Packages

Idpbuilder supports specifying custom packages using the flag `--package-dir` flag. This flag expects a directory containing ArgoCD application files.

Examples of using custom packages are available in the [example](./examples) directory. 
Let's take a look at [this example](examples/basic). This defines two custom package directories to deploy to the cluster.

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


## Contributing

If you'd like to contribute to the project or know the architecture and internals of this project, check out the [contribution doc](./CONTRIBUTING.md).

## Running in Codespaces

1. Create a Codespaces instance. ![img](examples/ref-implementation/images/codespaces-create.png)
2. Wait for it to be ready. It may take several minutes.
3. Get the latest release of idpbuilder:
   ```bash
    version=$(curl -Ls -o /dev/null -w %{url_effective} https://github.com/cnoe-io/idpbuilder/releases/latest)
    version=${version##*/}
    curl -L -o ./idpbuilder.tar.gz "https://github.com/cnoe-io/idpbuilder/releases/download/${version}/idpbuilder-$(uname | awk '{print tolower($0)}')-$(uname -m | sed 's/x86_64/amd64/').tar.gz"
    tar xzf idpbuilder.tar.gz
   ```
4. Run idpbuilder:
   ```
    idpbuilder create --protocol http  \
    --host ${CODESPACE_NAME}-8080.${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN} \
    --port 8080 --use-path-routing
   ```
5. Because Codespaces gives a single externally routable host name for an instance, idpbuilder must deploy with path based routing. 
   This means ArgoCD and Gitea UIs are given with the following commands.
   * ArgoCD: `echo https://${CODESPACE_NAME}-8080.${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN}/argocd`
   * Gitea: `echo https://${CODESPACE_NAME}-8080.${GITHUB_CODESPACES_PORT_FORWARDING_DOMAIN}/gitea`
6. Note that not all examples work with path based routing. 

## Extending the IDP builder

We are actively working to include more patterns and examples of extending idpbuilder to get started easily.

<!-- JUST BADGES & LINKS -->
[codespell-badge]: https://github.com/cnoe-io/idpbuilder/actions/workflows/codespell.yaml/badge.svg
[codespell-link]: https://github.com/cnoe-io/idpbuilder/actions/workflows/codespell.yaml

[e2e-badge]: https://github.com/cnoe-io/idpbuilder/actions/workflows/e2e.yaml/badge.svg
[e2e-link]: https://github.com/cnoe-io/idpbuilder/actions/workflows/e2e.yaml

[report-badge]: https://goreportcard.com/badge/github.com/cnoe-io/idpbuilder
[report-link]: https://goreportcard.com/report/github.com/cnoe-io/idpbuilder

[commit-activity-badge]: https://img.shields.io/github/commit-activity/m/cnoe-io/idpbuilder
[commit-activity-link]: https://github.com/cnoe-io/idpbuilder/pulse
