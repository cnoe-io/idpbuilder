INSTALL_YAML="pkg/controllers/localbuild/resources/nginx/k8s/ingress-nginx.yaml"

echo "# INGRESS-NGINX INSTALL RESOURCES" > ${INSTALL_YAML}
echo "# This file is auto-generated with 'hack/ingress-nginx/generate-manifests.sh'" >> ${INSTALL_YAML}
kustomize build ./hack/ingress-nginx/ >> ${INSTALL_YAML}
