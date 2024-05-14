# Terraform Integrations for Backstage

`idpBuilder` is now experimentally extensible to launch custom terraform patterns using package extensions. This is an experimental effort allowing the users of the `idpBuilder` to run terraform modules using the tooling in place.

Please use the below command to deploy an IDP reference implementation with an Argo application for terraform integrations with few sample patterns we have built:

```bash
idpbuilder create \
  --use-path-routing \
  --package-dir examples/ref-implementation \
  --package-dir examples/terraform-integrations
```

As you see above, this add-on to `idpbuilder` has a dependency to the [reference implementation](../ref-implementation/). This command primarily does the following:

1. Installs `fluxcd` source respository controller as an `argo` application.
2. Installs `tofu-controller` for managing the lifecycle of terraform deployments from your Kubernetes cluster for operations such as create, delete and update.
3. Finally, this stack add-on goes together with the work done under [backstage-terraform-integrations](https://github.com/cnoe-io/backstage-terraform-integrations/). Once the add-on is enabled, the user will need to follow the setup discussed in the [backstage-terraform-integrations](https://github.com/cnoe-io/backstage-terraform-integrations/) repo for the remainder of the configuration, and terraform integrations should work.
