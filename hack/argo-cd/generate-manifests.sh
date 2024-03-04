#!/bin/bash

INSTALL_YAML="pkg/controllers/localbuild/resources/argo/install.yaml"
INGRESS_YAML="pkg/controllers/localbuild/resources/argo/ingress.yaml"

echo "# UCP ARGO INSTALL RESOURCES" > ${INSTALL_YAML}
echo "# This file is auto-generated with 'hack/argo-cd/generate-manifests.sh'" >> ${INSTALL_YAML}
kustomize build ./hack/argo-cd/ >> ${INSTALL_YAML}

cat ./hack/argo-cd/ingress.yaml.tmpl > ${INGRESS_YAML}
