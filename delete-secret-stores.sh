#!/bin/bash

kubectl delete clustersecretstore argocd
kubectl delete clustersecretstore gitea
kubectl delete clustersecretstore keycloak

