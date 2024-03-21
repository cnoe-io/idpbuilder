#!/bin/bash

kind create cluster --name localdev --config kind-config.yaml
./install-kuik.sh
./idpbuilder create --package-dir examples/ref-implementation
./force-delete-ref-impl-argocd.sh
./delete-base-namespaces.sh
./delete-ref-impl-namespaces.sh
./delete-secret-stores.sh
echo "You should now disconnect your network and try to run ./idpbuilder create"
