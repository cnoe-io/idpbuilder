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
- Define packages as Argo CD Applications.
- Imperative pipelines for configuring packages are handled with ArgoCD resource hooks and Argo Workflows.

Currently, ArgoCD is a base requirement for both the AWS reference implementation and idpbuilder. 


## Future improvements

- Extend support for using different tools for CD and imperative workflows. For example, we may consider supporting Tekton or Flagger for running imperative workflows.

## Alternatives Considered

#### Use OCI images as applications

Projects such as Sealer and Kapp aim to use OCI images as the artifact to define and deploy multiple Kubernetes resources. This has a few advantages. 

- Immutable single artifact that can be used to deploy to different clusters.
- Simplicity. Application dependencies and supporting resources are defined and confined in the image.
- Can use standard Kubernetes YAML files, Helm Charts, and Kustomize. 

In addition, both tools support applying changes in particular order. For example, you can run a Kubernetes Job to migrate database schema before rolling out a new image. 

Since many of the goals are covered by both projects, it is possible to incorporate some of their tools and libraries to implement our goals. 

While this approach addresses most of our goals, there are drawbacks.

Firstly, it introduces a new layer for end users to debug. For end users to debug an issue related to Kubernetes manifest rendering, they now need to:
1. Figure out which OCI image contains the package with the problem. The problem may reside in one of dependent OCI images. 
2. Extract the contents of the OCI image. Determine which one file is responsible for the issue. 
3. Correct the issue.
4. Publish a new image.

Secondly, these tools and using OCI images as packages are not well adopted by CNOE members as evidenced by the tech radar. idpbuilder should be useful and relevant to CNOE members.


