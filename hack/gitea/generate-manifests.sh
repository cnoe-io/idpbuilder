#!/usr/bin/env bash
set -e

INSTALL_YAML="pkg/controllers/localbuild/resources/gitea/k8s/install.yaml"
GITEA_DIR="./hack/gitea"
CHART_VERSION="9.5.1"

echo "# GITEA INSTALL RESOURCES" >${INSTALL_YAML}
echo "# This file is auto-generated with 'hack/gitea/generate-manifests.sh'" >>${INSTALL_YAML}

helm repo add gitea-charts --force-update https://dl.gitea.com/charts/
helm repo update
helm template my-gitea gitea-charts/gitea -f ${GITEA_DIR}/values.yaml --version ${CHART_VERSION} >>${INSTALL_YAML}
sed -i.bak '3d' ${INSTALL_YAML}

# helm template for pvc uses Release.namespace which doesn't get set
# when running the helm "template" command
# See: https://gitea.com/gitea/helm-chart/issues/630
# and: https://gitea.com/gitea/helm-chart/src/commit/3b2b700441e91a19a535e05de3a9eab2fef0b117/templates/gitea/pvc.yaml#L6
# and: https://github.com/helm/helm/issues/3553#issuecomment-1186518158
# and: https://github.com/splunk/splunk-connect-for-kubernetes/pull/790
sed -i.bak 's/namespace: default/namespace: gitea/g' ${INSTALL_YAML}

cat ${GITEA_DIR}/ingress.yaml.tmpl >>${INSTALL_YAML}

rm -rf "${INSTALL_YAML}.bak"
