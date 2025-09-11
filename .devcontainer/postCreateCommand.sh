#!/usr/bin/env bash

source .devcontainer/setup-ssh.sh

source .devcontainer/install-kubectl.sh 

source .devcontainer/install-kind.sh 

source .devcontainer/install-claude-code.sh

# setup autocomplete for kubectl and alias k
sudo apt-get update -y && sudo apt-get install bash-completion -y
mkdir $HOME/.kube
echo "source <(kubectl completion bash)" >> $HOME/.bashrc
echo "alias k=kubectl" >> $HOME/.bashrc
echo "complete -F __start_kubectl k" >> $HOME/.bashrc

# Configure git if environment variables are set
if [ -n "$GIT_COMMITER_NAME" ]; then
    echo "Configuring git user.name to: $GIT_COMMITER_NAME"
    git config --global user.name "$GIT_COMMITER_NAME"
fi

if [ -n "$GIT_COMMITER_EMAIL" ]; then
    echo "Configuring git user.email to: $GIT_COMMITER_EMAIL"
    git config --global user.email "$GIT_COMMITER_EMAIL"
fi

# 1. Configure GPG agent
mkdir -p ~/.gnupg
echo "pinentry-program /usr/bin/pinentry" > ~/.gnupg/gpg-agent.conf
echo "allow-loopback-pinentry" >> ~/.gnupg/gpg-agent.conf

# 2. Configure GPG client
echo "use-agent" > ~/.gnupg/gpg.conf
echo "pinentry-mode loopback" >> ~/.gnupg/gpg.conf

# 3. Restart GPG agent and set environment
gpgconf --kill gpg-agent
export GPG_TTY=$(tty)
echo 'export GPG_TTY=$(tty)' >> ~/.bashrc

# 4. Configure Git for GPG signing
git config --global commit.gpgsign true
git config --global tag.gpgsign true
git config --global gpg.program gpg
