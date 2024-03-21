#!/bin/bash

kind create cluster --name localdev --config kind-config.yaml
echo "Installing Certmanager"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.yaml
kubectl -n cert-manager wait deployment/cert-manager-webhook --for condition=available
kubectl -n cert-manager wait deployment/cert-manager --for condition=available
echo "Installing Kuik"
./install-kuik.sh
echo "Running idpbuilder"
./idpbuilder create --package-dir examples/ref-implementation
read  -n 1 -p "Check argocd for installation success and then press any key to continue..."
./force-delete-ref-impl-argocd.sh
./delete-base-namespaces.sh
./delete-ref-impl-namespaces.sh
./delete-secret-stores.sh
echo "You should now disconnect your network and try to run ./idpbuilder create"
