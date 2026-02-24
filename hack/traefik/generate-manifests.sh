#!/usr/bin/env bash
# Output is written to pkg/controllers/localbuild/resources/traefik/k8s/traefik.yaml
set -euo pipefail

INSTALL_YAML="pkg/controllers/localbuild/resources/traefik/k8s/traefik.yaml"
TRAEFIK_DIR="./hack/traefik"
CHART_VERSION="39.0.2"

echo "" >${INSTALL_YAML}

helm repo add traefik-charts --force-update https://helm.traefik.io/traefik
helm repo update
helm template my-traefik traefik-charts/traefik -n traefik -f ${TRAEFIK_DIR}/values.yaml --version ${CHART_VERSION} >>${INSTALL_YAML}

# Traefik names the https port websecure and the http port web, for our service to work properly we need to rename them
sed -i.bak 's/name: websecure$/name: https/' ${INSTALL_YAML}
sed -i.bak 's/name: web$/name: http/' ${INSTALL_YAML}

cat ${TRAEFIK_DIR}/service.yaml.tmpl >>${INSTALL_YAML}

echo "---" >>${INSTALL_YAML}

cat ${TRAEFIK_DIR}/crds.yaml >>${INSTALL_YAML}

cat ${TRAEFIK_DIR}/tlsstore.yaml >>${INSTALL_YAML}

rm -rf "${INSTALL_YAML}.bak"
