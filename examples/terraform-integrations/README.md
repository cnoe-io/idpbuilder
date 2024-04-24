# Terraform Integrations for Backstage

This is an experimental effort allowing the users of the idpBuilder to run terraform modules using the tooling in place.

This stack add-on goes together with the work done under [backstage-terraform-integrations](https://github.com/cnoe-io/backstage-terraform-integrations/).
Once the add-on is enabled, the user will need to follow the setup discussed in the [backstage-terraform-integrations](https://github.com/cnoe-io/backstage-terraform-integrations/) repo for the remainder of the configuration, and terraform integrations should work.

This add-on has a dependency to the [reference implementation](../ref-implementation/). In order to enable the add-on with 
idpBuilder run the following:

``` 
./idpbuilder create --use-path-routing --package-dir examples/ref-implementation --package-dir examples/terraform-integrations 
```
