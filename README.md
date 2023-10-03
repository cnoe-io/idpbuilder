# IDP Builder

Internal development platform binary launcher.

## About

Spin up a complete internal developer platform using industry standard technologies like Kubernetes, Argo, and backstage with only Docker required as a dependency.

This is also a completely self-contained binary, meaning you can get up and running simply by downloading a binary release and executing it!

## Quickstart:

### Build

`make`

### Run

`./idpbuilder -buildName localdev`

### Use

Kubernetes: `kubectl get pods`

Argo: Visit https://argocd.idpbuilder.cnoe.io.local:8443/

Backstage: http://backstage.idpbuilder.cnoe.io.local/

## Architecture

The IDP builder binary is primarily composed of a wrapper around a [Kubebuilder](https://kubebuilder.io) based operator and associated type called localbuild. See: [api/v1alpha1/localbuild_types.go](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/api/v1alpha1/localbuild_types.go#L28-L66) and [pkg/controllers/localbuild/controller.go](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/localbuild/controller.go#L54-L84)

However it starts out by creating a Kind cluster to register the CRD and controller for localbuild and to host the resources created by it which in turn create the basis for our IDP. You can see the steps taken to install the dependencies and stand up the localbuild controller in the CLI codebase here: [pkg/build/build.go](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/build/build.go#L95-L131)

### Kind Cluster
The Kind cluster is created using the Kind library and only requires Docker be installed on the host. See: [ReconcileKindCluster](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/build/build.go#L39-L59)

### Localbuild

Localbuild's reconciler steps through a number of subreconcilers to create all of the IDP components. See: [Subreconcilers](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/localbuild/controller.go#L69-L74)

The subreconcilers currently include:

* [ReconcileProjectNamespace:](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/localbuild/controller.go#L102C32-L102C57) Creates a namespace for the Localbuild objects
* [ReconcileArgo:](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/localbuild/argo.go#L51) Installs ArgoCD
* [ReconcileEmbeddedGitServer:](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/localbuild/controller.go#L125) Installs a "gitserver" which is another Kubebuilder Operate in this project See: [api/v1alpha1/gitserver_types.go](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/api/v1alpha1/gitserver_types.go)
* [ReconcileArgoApps](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/localbuild/controller.go#L172) Steps through all the "Embedded" Argo Apps and installs them. See: [Embedded Apps](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/resources.go#L20-L32)

### GitServer

Gitserver is essentially a fileserver for the Argo apps that are packaged within this IDP builder. As you might expect, it serves the files used by the ArgoCD apps using the git protocol. You can see the container image that contains these files get built here in the [GitServer Reconciler](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/gitserver/image.go#L44-L60)

### Embedded Argo Apps

The embedded Argo apps are created by the Localbuild reconciler See: [Argo Apps](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/controllers/localbuild/controller.go#L210-L243) Then they are picked up by the ArgoCD operator which in turn connects to the GitServer to perform their gitops deployment.

The resources served by the GitServer are contained within this binary here: [pkg/apps/srv](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/srv/)

They include:
* [ArgoCD Ingress:](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/srv/argocd/ingress.yaml) which makes the ArgoCD interface available.
* [Backstage:](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/srv/backstage/install.yaml) which intalls the Backstage Resources.
* [Crossplane:](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/srv/crossplane/crossplane.yaml) which uses the Crossplane Helm chart to install Crossplane.
* [Nginx Ingress Controller:](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/srv/nginx-ingress/ingress-nginx.yaml) which makes nginx ingresses available on the cluster.

As you can imagine each of these apps are are deployed by ArgoCD to the Kind cluster created by CLI. The Argo apps can be inspected with kubectl in the `argocd` namespace and the resources they create can be seen in their corresponding namespaces (`backstage` and `crossplane`)

## Extending the IDP builder
In the future we hope to allow for a pluggable interface to allow for extending the IDP builder with additional argo apps. For now you simply need to add additional apps in the [Embedded Apps](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/resources.go#L20-L32) and also add the resources they will deploy in the `srv` folder: [pkg/apps/srv](https://github.com/cnoe-io/idpbuilder/blob/4b0f8ecdd7266083373da51d5add1bca73e05a33/pkg/apps/srv/)
