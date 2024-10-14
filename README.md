[![Codespell][codespell-badge]][codespell-link]
[![E2E][e2e-badge]][e2e-link]
[![Go Report Card][report-badge]][report-link]
[![Commit Activity][commit-activity-badge]][commit-activity-link]

# IDP Builder

Internal development platform binary launcher.

## About

Spin up a complete internal developer platform using industry standard technologies like Kubernetes, Argo, and backstage with only Docker required as a dependency.

This can be useful in several ways:
* Create a single binary which can demonstrate an IDP reference implementation.
* Use within CI to perform integration testing.
* Use as a local development environment for platform engineers.

## Getting Started

The easiest way to get started is to grab the idpbuilder binary for your platform and run it. You can visit our [nightly releases](https://github.com/cnoe-io/idpbuilder/releases/latest) page to download the version for your system, or run the following commands:

```bash
arch=$(if [[ "$(uname -m)" == "x86_64" ]]; then echo "amd64"; else uname -m; fi)
os=$(uname -s | tr '[:upper:]' '[:lower:]')


idpbuilder_latest_tag=$(curl --silent "https://api.github.com/repos/cnoe-io/idpbuilder/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
curl -LO  https://github.com/cnoe-io/idpbuilder/releases/download/$idpbuilder_latest_tag/idpbuilder-$os-$arch.tar.gz
tar xvzf idpbuilder-$os-$arch.tar.gz
```

You can then run idpbuilder with the create argument to spin up your CNOE IDP:

```bash
./idpbuilder create
```

For more detailed information, checkout our [documentation website](https://cnoe.io/docs/reference-implementation/installations/idpbuilder) on getting started with idpbuilder.

## Community

- If you have questions or concerns about this tool, please feel free to reach out to us on the [CNCF Slack Channel](https://cloud-native.slack.com/archives/C05TN9WFN5S).
- You can also join our community meetings to meet the team and ask any questions. Checkout [this calendar](https://calendar.google.com/calendar/embed?src=064a2adfce866ccb02e61663a09f99147f22f06374e7a8994066bdc81e066986%40group.calendar.google.com&ctz=America%2FLos_Angeles) for more information.

## Contribution

Checkout the [contribution doc](./CONTRIBUTING.md) for contribution guidelines and more information on how to set up your local environment.


<!-- JUST BADGES & LINKS -->
[codespell-badge]: https://github.com/cnoe-io/idpbuilder/actions/workflows/codespell.yaml/badge.svg
[codespell-link]: https://github.com/cnoe-io/idpbuilder/actions/workflows/codespell.yaml

[e2e-badge]: https://github.com/cnoe-io/idpbuilder/actions/workflows/e2e.yaml/badge.svg
[e2e-link]: https://github.com/cnoe-io/idpbuilder/actions/workflows/e2e.yaml

[report-badge]: https://goreportcard.com/badge/github.com/cnoe-io/idpbuilder
[report-link]: https://goreportcard.com/report/github.com/cnoe-io/idpbuilder

[commit-activity-badge]: https://img.shields.io/github/commit-activity/m/cnoe-io/idpbuilder
[commit-activity-link]: https://github.com/cnoe-io/idpbuilder/pulse
