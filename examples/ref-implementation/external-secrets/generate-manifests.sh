#!/bin/bash
set -e

INSTALL_YAML="manifests/install.yaml"
CHART_VERSION="0.9.11"

echo "# EXTERNAL SECRETS INSTALL RESOURCES" > ${INSTALL_YAML}
echo "# This file is auto-generated with 'examples/ref-impelmentation/external-secrets/generate-manifests.sh'" >> ${INSTALL_YAML}

helm repo add external-secrets --force-update https://charts.external-secrets.io
helm repo update
helm template --namespace external-secrets external-secrets external-secrets/external-secrets -f values.yaml --version ${CHART_VERSION} >> ${INSTALL_YAML}
