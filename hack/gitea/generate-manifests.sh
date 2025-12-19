#!/bin/bash
set -e

INSTALL_YAML="pkg/controllers/localbuild/resources/gitea/k8s/install.yaml"
GITEA_DIR="./hack/gitea"
CHART_VERSION="12.1.2"

echo "# GITEA INSTALL RESOURCES" >${INSTALL_YAML}
echo "# This file is auto-generated with 'hack/gitea/generate-manifests.sh'" >>${INSTALL_YAML}

helm repo add gitea-charts --force-update https://dl.gitea.com/charts/
helm repo update
helm template my-gitea gitea-charts/gitea -f ${GITEA_DIR}/values.yaml --version ${CHART_VERSION} >>${INSTALL_YAML}

# Remove the third line (helm template comment) and replace namespace
sed '3d' ${INSTALL_YAML} | sed 's/namespace: default/namespace: gitea/g' > ${INSTALL_YAML}.tmp

# helm template for pvc uses Release.namespace which doesn't get set
# when running the helm "template" command
# See: https://gitea.com/gitea/helm-chart/issues/630
# and: https://gitea.com/gitea/helm-chart/src/commit/3b2b700441e91a19a535e05de3a9eab2fef0b117/templates/gitea/pvc.yaml#L6
# and: https://github.com/helm/helm/issues/3553#issuecomment-1186518158
# and: https://github.com/splunk/splunk-connect-for-kubernetes/pull/790

cat ${GITEA_DIR}/ingress.yaml.tmpl >>${INSTALL_YAML}.tmp
mv ${INSTALL_YAML}.tmp ${INSTALL_YAML}
