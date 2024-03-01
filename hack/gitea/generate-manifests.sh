#!/bin/bash
set -e

INSTALL_YAML="pkg/controllers/localbuild/resources/gitea/k8s/install.yaml"
GITEA_DIR="./hack/gitea"
CHART_VERSION="9.5.1"

echo "# GITEA INSTALL RESOURCES" > ${INSTALL_YAML}
echo "# This file is auto-generated with 'hack/gitea/generate-manifests.sh'" >> ${INSTALL_YAML}

helm repo add gitea-charts --force-update https://dl.gitea.com/charts/
helm repo update
helm template my-gitea gitea-charts/gitea -f ${GITEA_DIR}/values.yaml --version ${CHART_VERSION} >> ${INSTALL_YAML}
sed -i '3d' ${INSTALL_YAML}

cat ${GITEA_DIR}/ingress.yaml.tmpl >> ${INSTALL_YAML}
cat ${GITEA_DIR}/gitea-creds.yaml >> ${INSTALL_YAML}
