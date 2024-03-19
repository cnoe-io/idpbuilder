#!/bin/bash

kind create cluster --name localdev --config kind-config.yaml
./install-kuik.sh
./idpbuilder create
./force-delete-ref-impl-argocd.sh
./delete-base-namespaces.sh
./delete-ref-impl-namespaces.sh
echo "You should now disconnect your network and try to run ./idpbuilder create"
