#!/bin/bash

INSTALL_YAML="pkg/controllers/localbuild/resources/nginx/k8s/ingress-nginx.yaml"
NGINX_DIR="./hack/ingress-nginx"


echo "# INGRESS-NGINX INSTALL RESOURCES" > ${INSTALL_YAML}
echo "# This file is auto-generated with 'hack/ingress-nginx/generate-manifests.sh'" >> ${INSTALL_YAML}
kustomize build ${NGINX_DIR} >> ${INSTALL_YAML}

cat ${NGINX_DIR}/service-ingress-nginx.yaml.tmpl >> ${INSTALL_YAML}
