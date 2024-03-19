#!/bin/bash
./force-delete-ref-impl-argocd.sh
./delete-base-namespaces.sh
./delete-ref-impl-namespaces.sh
