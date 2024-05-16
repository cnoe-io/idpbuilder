# Localstack Integration

Please use the below command to deploy an IDP reference implementation with an Argo application that adds Localstack, as well as integrating with Crossplane.

```bash
idpbuilder create \
  --use-path-routing \
  --package-dir examples/ref-implementation \
  --package-dir examples/localstack-integration
```

As you see above, this add-on to `idpbuilder` has a dependency on the [reference implementation](../ref-implementation/). This command primarily does the following:

1. Installs `localstack` helmchart as an `argo` application.
2. Adds localstack crossplane ProviderConfig, targetting localstack

Once the custom package is installed, localstack can be used from the backstage template `app-with-aws-resources`, by changing the `providerConfigName` during the bucket configuration page from `default` to `localstack`.
