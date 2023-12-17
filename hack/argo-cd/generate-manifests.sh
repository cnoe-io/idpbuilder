INSTALL_YAML="pkg/controllers/localbuild/resources/argo/install.yaml"

echo "# UCP ARGO INSTALL RESOURCES" > ${INSTALL_YAML}
echo "# This is an auto-generated file. DO NOT EDIT" >> ${INSTALL_YAML}
kustomize build ./hack/argo-cd/ >> ${INSTALL_YAML}
