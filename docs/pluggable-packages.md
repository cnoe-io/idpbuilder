# Pluggable packaging

## Background

`idpbuilder` is a tool that aims to:
- Allow developers to stand up a Kubernetes cluster with components that make up a Internal Developer Platform (IDP). 
- Allow for tests to run against IDPs in Continuous Integration systems.
- Standup a working IDP for demo purposes.

It also aims to achieve the above goals while having a single dependency, Docker, at run time. 

When implementing IDPs using open source projects, there is no one set of projects that fit the needs of all organizations because:
1. No organization went through the same path to reach the point of needing a IDP.
1. Past technology choices were made for specific needs of each organization. This results in organizationally unique technology inertia that may not be correctable in the near future.
1. No industry consensus and standards for choosing specific IDP components.

To fit the needs from different organizations, idpbuilder needs to be flexible in the what and how it can deploy different packages. Currently idpbuilder uses ArgoCD to install a [set of application](https://github.com/cnoe-io/idpbuilder/blob/56089e4ae3b27cf90641bfbff2a96c36dd5263e1/pkg/apps/resources.go#L20-L32), and they cannot be changed without modifying the source code.

## Goals

The proposal in this document should:

1. Make the packages installed by `idpbuilder` configurable.
1. Minimize the number of runtime dependencies necessary.
1. Make it easy for end users to configure their packages.
1. Support custom setup scripts and workflows that need to be executed in a defined order before a package becomes ready.

## Proposal

This document proposes the following:
- Make ArgoCD and Argo Workflows hard requirements. 
- Define packages as Argo CD Applications (Helm, Kustomize, and raw manifests)
- Imperative pipelines for configuring packages are handled with ArgoCD resource hooks.

Currently, ArgoCD is a base requirement for both the AWS reference implementation and idpbuilder but not yet officially made a hard requirement. ArgoCD is the CD of choice for CNOE members and it should be the focus over other GitOps solutions. 

Packages then become the formats that ArgoCD supports natively: Helm charts, Kustomize, and raw manifests. We need to ensure idpbuilder support all of them. 

Regardless of how applications are delivered declaratively, custom scripts are often needed before, during, and after application syncing to ensure applications reach the desired state. idpbuilder needs to provide a way to run custom scripts in a defined order. 

We could define a spec to support such cases, but considering the main use cases of idpbuilder center around everything being local, it doesn't require overly complex tasks. For example, in the reference implementation for AWS, the majority of scripting are done to manage authentication mechanisms. Another task that scripts do is domain name configuration for each package such as setting the `baseUrl` field in the Backstage configuration file. In local environments, tasks like these are likely unnecessary.

ArgoCD supports resource hooks which allow users to define tasks to be run during application syncing. While there are some limitations to what it can do, for the majority of simple tasks it should suffice.

### Runtime Git server content generation

The git server uses Go's embed capability to serve contents to ArgoCD.  Because of this, to use different manifests in idpbuilder, it requires compiling the go application. For Go developers this approach is straightforward to work with. For non-Go developers this approach may be frustrating when needing to debug Go errors while compiling the program.

To solve this, Git content should be created at run time by introducing a new flag to idpbuilder. This flag takes a directory and builds an image with the content from the directory.


## Future improvements

- Extend support for using different tools for CD and imperative workflows. For example, we may consider supporting Tekton or Flagger for running imperative workflows.
- Extend support for other GitOps solutions such as Flux CD.
- Add support for authentication. As use cases grow, it will be necessary to support authentication mechanisms. Example use cases: GitHub credentials for ArgoCD, AWS credentials to pull images from a ECR registry.

## Alternatives Considered

#### Use OCI images as applications

Projects such as Sealer and Kapp aim to use OCI images as the artifact to define and deploy multiple Kubernetes resources. This has a few advantages. 

- Immutable single artifact that can be used to deploy to different clusters.
- Simplicity. Application dependencies and supporting resources are defined and confined in the image.
- Can use standard Kubernetes YAML files, Helm Charts, and Kustomize. 
- Native support for signing and verification through industry standard tools like cosign.

In addition, both tools support applying changes in particular order. For example, you can run a Kubernetes Job to migrate database schema before rolling out a new image. 

Since many of the goals are covered by both projects, it is possible to incorporate some of their tools and libraries to implement our goals. 

While this approach addresses most of our goals, there are drawbacks.

Firstly, it introduces a new layer for end users to debug. For end users to debug an issue related to Kubernetes manifest rendering, they now need to:
1. Figure out which OCI image contains the package with the problem. The problem may reside in one of dependent OCI images. 
2. Extract the contents of the OCI image. Determine which one file is responsible for the issue. 
3. Correct the issue.
4. Publish a new image.

Secondly, these tools and using OCI images as packages are not well adopted by CNOE members as evidenced by the tech radar. idpbuilder should be useful and relevant to CNOE members.

#### Define specs for imperative work

To support use cases where more complex steps are needed to orchestrate different services in to the cluster, idpbuilder could define a configuration spec where end users can define their own steps. For example, a spec may look something like the following. 

```yaml
apiVersion: idpbuilder.cnoe.io/v1alpha1
kind: Config
metadata:
  name: test
spec:
  packages:
    - name: crossplane
      preCreation:
        - name: job1
          secretRef:
            name: job1
            path: ./secret1
          spec:
            apiVersion: batch/v1
            kind: Job
            metadata:
              name: pi
      postCreation:
        ...

```

This introduces a few problems. 

1. Complexity.

    This introduces a completely new mechanisms to manage applications. Current idpbuilder design is very simple with no concept of pipelining. It allows end users to define manifests at compile time, and apply them as Argo CD applications. Introducing this feature and maintaining it may involve significant time commitment.


2. Yet another configuration format.

    With the explosion of Kubernetes based projects, a large number of configuration formats were created. Developers and operators are already needing to write and manage configuration files that look similar but slightly different.
  
