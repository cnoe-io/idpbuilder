#!/usr/bin/env sh

curl -sLS https://get.arkade.dev | sudo sh

arkade get kind
arkade get kubectl
arkade get helm
arkade get gh
echo "export PATH=\$PATH:/home/vscode/.arkade/bin" >> $HOME/.bashrc

# Make sure go path is owned by vscode
sudo chown -R vscode:vscode /home/vscode/go

# Add idpbuilder to path
echo "export PATH=\$PATH:/home/vscode/go/src/github.com/cnoe-io/idpbuilder" >> $HOME/.bashrc
