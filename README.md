# IDP Builder

Internal development platform binary launcher.

## About

Spin up a complete internal developer platform using industry standard technologies like Kubernetes, Argo, and backstage with only Docker required as a dependency.

This is also a completely self-contained binary, meaning you can get up and running simply by downloading a binary release and executing it!

## Build

`make`

## Run

`./idpbuilder -buildName localdev`

## Use

Kubernetes: `kubectl get pods`

Argo: Visit https://argocd.idpbuilder.adskeng.localtest.me:8443/

Backstage: TBD