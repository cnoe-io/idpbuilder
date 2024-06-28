# Reference implementation

This example creates a local version of the CNOE reference implementation.

## Prerequisites

Ensure you have the following tools installed on your computer.

**Required**

- [idpbuilder](https://github.com/cnoe-io/idpbuilder/releases/latest): version `0.3.0` or later
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl): version `1.27` or later
- Your computer should have at least 6 GB RAM allocated to Docker. If you are on Docker Desktop, see [this guide](https://docs.docker.com/desktop/settings/mac/).

**Optional**

- AWS credentials: Access Key and secret Key. If you want to create AWS resources in one of examples below.

## Installation

**_NOTE:_**
- If you'd like to run this in your web browser through Codespaces, please follow [the instructions here](./codespaces.md) to install instead.

- _This example assumes that you run the reference implementation with the default port configguration of 8443 for the idpBuilder.
If you happen to configure a different host or port for the idpBuilder, the manifests in the reference example need to be updated
and be configured with the new host and port. you can use the [replace.sh](replace.sh) to change the port as desired prior to applying the manifest as instructed in the command above._

Run the following command from the root of this repository.

```bash
idpbuilder create --use-path-routing --package-dir examples/ref-implementation-appset
```
