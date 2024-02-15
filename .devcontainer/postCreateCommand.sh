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

# Install Github CLI
curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
&& sudo chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
&& echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
&& sudo apt update \
&& sudo apt install gh -y

# Make sure go path is owned by vscode
sudo chown -R vscode:vscode /home/vscode/go


# Setup kubectl and k autocompletion
sudo apt install bash-completion
printf "
source <(kubectl completion bash)
alias k=kubectl
complete -F __start_kubectl k
" >> $HOME/.bashrc


