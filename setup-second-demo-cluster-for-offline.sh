#!/bin/bash

kind create cluster --name secondlocaldev --config secondlocaldev-kind-config.yaml
./install-kuik.sh
./idpbuilder create --build-name secondlocaldev --port 9443 --package-dir examples/ref-implementation
read  -n 1 -p "Check argocd for installation success and then press any key to continue..."
./force-delete-ref-impl-argocd.sh
./delete-base-namespaces.sh
./delete-ref-impl-namespaces.sh
./delete-secret-stores.sh
echo "You should now disconnect your network and try to run ./idpbuilder create"
