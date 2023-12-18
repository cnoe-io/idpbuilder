INSTALL_YAML="pkg/controllers/localbuild/resources/argo/install.yaml"

echo "# UCP ARGO INSTALL RESOURCES" > ${INSTALL_YAML}
echo "# This file is auto-generated with 'hack/argo-cd/generate-manifests.sh'." >> ${INSTALL_YAML}
kustomize build ./hack/argo-cd/ >> ${INSTALL_YAML}
