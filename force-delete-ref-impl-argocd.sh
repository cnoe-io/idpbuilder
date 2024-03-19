#!/bin/bash
kubectl delete namespace argocd --wait=false --timeout=0s --force=true --grace-period=-1
kubectl -n argocd patch applications.argoproj.io spark-operator -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io crossplane -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io crossplane-providers -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io crossplane-compositions -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io external-secrets -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io argo-workflows -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io backstage -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io backstage-templates -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io keycloak -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io gitea -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io argocd -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io ingress-nginx -p '{"metadata":{"finalizers":null}}' --type=merge
kubectl -n argocd patch applications.argoproj.io core-dns -p '{"metadata":{"finalizers":null}}' --type=merge
