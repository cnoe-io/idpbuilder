# Contributing guide

Welcome to the project, and thanks for considering contributing to this project. 

If you have any questions or need clarifications on topics covered here, please feel free to reach out to us on the [#cnoe-interest](https://cloud-native.slack.com/archives/C05TN9WFN5S) channel on CNCF Slack.

## Setting up a development environment

To get started with the project on your machine, you need to install the following tools:
1. Go 1.21+. See [this official guide](https://go.dev/doc/install) from Go authors.
2. Make. You can install it through a package manager on your system. E.g. Install `build-essential` for Ubuntu systems.
3. Docker. Similar to Make, you can install it through your package manager or [Docker Desktop](https://www.docker.com/products/docker-desktop/).

Once required tools are installed, clone this repository. `git clone https://github.com/cnoe-io/idpbuilder.git`.

Then change your current working directory to the repository root. e.g. `cd idpbuilder`.

All subsequent commands described in this document assumes they are executed from the repository root.
Ensure your docker daemon is running and available. e.g. `docker images` command should not error out.

## Building from the main branch

1. Checkout the main branch. `git checkout main`
2. Build the binary. `make build`. This compiles the project. It will take several minutes for the first time. Example output shown below:
    ```
    ~/idpbuilder$ make build
    test -s /home/ubuntu/idpbuilder/bin/controller-gen && /home/ubuntu/idpbuilder/bin/controller-gen --version | grep -q v0.12.0 || \
    GOBIN=/home/ubuntu/idpbuilder/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.12.0
    /home/ubuntu/idpbuilder/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./api/..." output:crd:artifacts:config=pkg/controllers/resources
    /home/ubuntu/idpbuilder/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
    go fmt ./...
    go vet ./...
    go build -o idpbuilder main.go  
    ```
3. Once build finishes, you should have an executable file called `idpbuilder` in the root of the repository.
4. The file is ready to use. Execute this command to confirm: `./idpbuilder --help`


### Testing basic functionalities

To test the very basic functionality of idpbuilder, Run the following command: `./idpbuilder create`

This command creates a kind cluster, expose associated endpoints to your local machine using an ingress controller and deploy the following packages:

1. [Kind](https://kind.sigs.k8s.io/) cluster.
2. [ArgoCD](https://argo-cd.readthedocs.io/en/stable/) resources.
3. [Gitea](https://about.gitea.com/) resources.
4. [Backstage](https://backstage.io/) resources.

They are deployed as ArgoCD Applications with the Gitea repositories set as their sources. 

UIs for Backstage, Gitea, and ArgoCD are accessible on the machine:
* Gitea: http://gitea.cnoe.localtest.me:8443/explore/repos
* Backstage: http://backstage.cnoe.localtest.me:8880/
* ArgoCD: https://argocd.cnoe.localtest.me:8443/applications

ArgoCD username is `admin` and the password can be obtained with 
```
kubectl -n argocd get secret argocd-initial-admin-secret -o go-template='{{ range $key, $value := .data }}{{ printf "%s: %s\n" $key ($value | base64decode) }}{{ end }}'
```

Gitea admin credentials can be obtained with 
```
kubectl get secrets -n gitea gitea-admin-secret -o go-template='{{ range $key, $value := .data }}{{ printf "%s: %s\n" $key ($value | base64decode) }}{{ end }}'
```

All ArgoCD applications should be synced and healthy. You can check them in the UI or 
```
kubectl get application -n argocd
```
