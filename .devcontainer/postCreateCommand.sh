#!/usr/bin/env sh

TARGETARCH=amd64

# Install Kind
if ! command -v kubectl &> /dev/null; then
    echo "kind not found in PATH, installing"
    curl -sL -o kind "https://kind.sigs.k8s.io/dl/v0.22.0/kind-linux-${TARGETARCH}" && chmod +x ./kind
    sudo mv ./kind /usr/local/bin
fi

# Install kubectl
if ! command -v kubectl &> /dev/null; then
    echo "kubectl not found in PATH, installing"
    curl -sLO https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/${TARGETARCH}/kubectl && chmod +x ./kubectl
    sudo mv ./kubectl /usr/local/bin
    # Setup kubectl and k autocompletion
    sudo apt update
    sudo apt install bash-completion -y
    printf "
    source <(kubectl completion bash)
    alias k=kubectl
    complete -F __start_kubectl k
    " >> $HOME/.bashrc
fi

# Install helm
if ! command -v helm &> /dev/null; then
    set -e
    echo "helm not found in PATH, installing"
    bash -c "curl -s https://get.helm.sh/helm-v3.14.1-linux-${TARGETARCH}.tar.gz > helm3.tar.gz" && tar -zxvf helm3.tar.gz linux-${TARGETARCH}/helm && chmod +x linux-${TARGETARCH}/helm
    sudo mv linux-${TARGETARCH}/helm /usr/local/bin && rm helm3.tar.gz && rm -R linux-${TARGETARCH}
fi


# Install Github CLI when using ubuntu
if ! command -v gh &> /dev/null; then
    . /etc/os-release
    if [ "$ID" = "ubuntu" ]; then
        echo "Installing Github CLI"
        curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg | sudo dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
        && sudo chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
        && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" | sudo tee /etc/apt/sources.list.d/github-cli.list > /dev/null \
        && sudo apt update \
        && sudo apt install gh -y
    fi
fi

# Make sure go path is owned by vscode
sudo chown -R vscode:vscode /home/vscode/go

# Compile idpbuilder
echo "Compiling idpbuilder"
export PATH=$PATH:/usr/local/go/bin
make build
# Add idpbuilder to PATH
echo "export PATH=\$PATH:/home/vscode/go/src/github.com/cnoe-io/idpbuilder" >> $HOME/.bashrc
