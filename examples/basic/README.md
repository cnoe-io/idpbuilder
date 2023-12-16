
## Basic Example

This directory contains basic examples of using the custom package feature.

### Local manifests

The [package1 directory](./package1) is an example of a custom package that you have developed locally, and you want test.

This configuration instructs idpbuilder to:

1. Create a Gitea repository.
2. Sync the contents of the [manifests](./package1/manifests) directory to the repostiory.
3. Replace the `spec.Source(s).repoURL` field with the Gitea repository URL.

### Remote manifests

The [package2 directory](./package2) is an example for packages available remotely. This is applied directly to the cluster.
