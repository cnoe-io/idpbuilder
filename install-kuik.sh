#!/bin/bash

helm upgrade --install \                                              
     --create-namespace --namespace kuik-system \
     kube-image-keeper kube-image-keeper \
     --repo https://charts.enix.io/ \
     --set controllers.webhook.ignorePullPolicyAlways=false
