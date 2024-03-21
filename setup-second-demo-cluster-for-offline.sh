#!/bin/bash

kind create cluster --name secondlocaldev --config secondlocaldev-kind-config.yaml
echo "Installing Certmanager"
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.4/cert-manager.yaml
kubectl -n cert-manager wait deployment/cert-manager-webhook --for condition=available --timeout=600s
kubectl -n cert-manager wait deployment/cert-manager --for condition=available --timeout 600s
sleep 10
echo "Installing Kuik"
./install-kuik.sh
kubectl -n kuik-system wait deployment/kube-image-keeper-controllers --for condition=available --timeout 600s
kubectl -n kuik-system rollout status daemonset/kube-image-keeper-proxy --watch --timeout=600s
kubectl -n kuik-system rollout status statefulset/kube-image-keeper-registry --watch --timeout=600s
sleep 10
echo "Running idpbuilder"
./idpbuilder create --build-name secondlocaldev --port 9443 --package-dir examples/ref-implementation
read  -n 1 -p "Check argocd for installation success and then press any key to continue..."
./force-delete-ref-impl-argocd.sh
./delete-base-namespaces.sh
./delete-ref-impl-namespaces.sh
./delete-secret-stores.sh
echo "You should now disconnect your network and try to run ./idpbuilder create"
