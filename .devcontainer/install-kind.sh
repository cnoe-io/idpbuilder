#!/usr/bin/env bash
set -e

# Install KIND (Kubernetes IN Docker)
KIND_VERSION="v0.23.0" # Change as needed

echo "Installing KIND ${KIND_VERSION}..."

# Try to install to /usr/local/bin/kind with sudo, fallback to user local bin
if sudo curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64 2>/dev/null && sudo chmod +x /usr/local/bin/kind 2>/dev/null; then
    echo "KIND installed to /usr/local/bin/kind"
else
    echo "Installing KIND to user local bin..."
    mkdir -p /home/vscode/.local/bin
    curl -Lo /home/vscode/.local/bin/kind https://kind.sigs.k8s.io/dl/${KIND_VERSION}/kind-linux-amd64
    chmod +x /home/vscode/.local/bin/kind
    echo "KIND installed to /home/vscode/.local/bin/kind"
fi

echo "KIND installed!"
