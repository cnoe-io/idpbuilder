#!/bin/bash

kubectl delete namespace controller-demo-system
kubectl delete customresourcedefinitions.apiextensions.k8s.io mydeployments.apps.demo.cnoe.io
