#!/bin/bash

docker pull cgr.dev/chainguard/static:latest

docker tag cgr.dev/chainguard/static:latest ko.local/chainguard/static:latest

echo "defaultBaseImage: ko.local/chainguard/static" > .ko.yaml

export KO_DOCKER_REPO=gitea.cnoe.localtest.me:8443/giteaadmin/ && kustomize build config/overlays/local | ko resolve --local --insecure-registry -f - | yq .

echo "Now run the following while offline 'export KO_DOCKER_REPO=gitea.cnoe.localtest.me:8443/giteaadmin/ && kustomize build config/overlays/local | ko resolve --local --insecure-registry -f - | kubectl apply -f -'"
