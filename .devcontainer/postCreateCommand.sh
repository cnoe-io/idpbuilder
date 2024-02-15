#!/usr/bin/env sh
set -eux

TARGETOS=linux
TARGETARCH=amd64

# Install Kind
curl -L -o kind "https://kind.sigs.k8s.io/dl/v0.22.0/kind-linux-${TARGETARCH}" && chmod +x ./kind
sudo mv ./kind /usr/local/bin

# Install kubectl
curl -LO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/${TARGETARCH}/kubectl && chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin

# Install helm
bash -c "curl -s https://get.helm.sh/helm-v3.14.1-linux-${TARGETARCH}.tar.gz > helm3.tar.gz" && tar -zxvf helm3.tar.gz linux-${TARGETARCH}/helm && chmod +x linux-${TARGETARCH}/helm
sudo mv linux-${TARGETARCH}/helm /usr/local/bin && rm helm3.tar.gz && rm -R linux-${TARGETARCH}

# Install kubebuilder
curl -L -o kubebuilder "https://go.kubebuilder.io/dl/latest/${TARGETOS}/${TARGETARCH}" && chmod +x kubebuilder
sudo mv kubebuilder /usr/local/bin/


# install protocol buffer compiler (protoc)
sudo apt update
sudo apt install -y protobuf-compiler

sudo apt install bash-completion

printf "
source <(kubectl completion bash)
alias k=kubectl
complete -F __start_kubectl k
" >> $HOME/.bashrc


