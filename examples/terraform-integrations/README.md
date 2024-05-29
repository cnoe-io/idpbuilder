# Terraform Integrations for Backstage

`idpBuilder` is now experimentally extensible to launch custom terraform patterns using package extensions. This is an experimental effort allowing the users of the `idpBuilder` to run terraform modules using the tooling in place.

Please use the below command to deploy an IDP reference implementation with an Argo application for preparing up the setup for terraform integrations:

```bash
idpbuilder create \
  --use-path-routing \
  --package-dir examples/ref-implementation \
  --package-dir examples/terraform-integrations
```

As you see above, this add-on to `idpbuilder` has a dependency to the [reference implementation](../ref-implementation/). This command primarily does the following:

1. Installs `fluxcd` source repository controller as an `argo` application.
2. Installs `tofu-controller` for managing the lifecycle of terraform deployments from your Kubernetes cluster for operations such as create, delete and update.