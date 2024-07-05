#!/bin/bash

set -e -o pipefail
# get the latest stable release by look for tag name pattern like 'v*.*.*'.  For example, v1.1.1
# GitHub API returns releases in chronological order so we take the first matching tag name.
version=$(curl -s https://api.github.com/repos/cnoe-io/idpbuilder/releases | grep tag_name | grep -o -e '"v[0-9].[0-9].[0-9]"' | head -n1 | sed 's/"//g')

echo "Downloading idpbuilder version ${version}"
curl -L --progress-bar -o ./idpbuilder.tar.gz "https://github.com/cnoe-io/idpbuilder/releases/download/${version}/idpbuilder-$(uname | awk '{print tolower($0)}')-$(uname -m | sed 's/x86_64/amd64/').tar.gz"
tar xzf idpbuilder.tar.gz

echo "Moving idpbuilder binary to /usr/local/bin"
sudo mv ./idpbuilder /usr/local/bin/
idpbuilder version
echo "Successfully installed idpbuilder"
