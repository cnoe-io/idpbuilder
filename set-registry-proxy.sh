#!/bin/sh
KIND_NAME=${1-kind}
SETUP_URL=http://docker-registry-proxy:3128/setup/systemd
pids=""
for NODE in $(kind get nodes --name "$KIND_NAME"); do
  docker exec "$NODE" sh -c "\
      curl $SETUP_URL \
      | sed s/docker\.service/containerd\.service/g \
      | sed '/Environment/ s/$/ \"NO_PROXY=127.0.0.0\/8,10.0.0.0\/8,172.16.0.0\/12,192.168.0.0\/16\"/' \
      | bash" & pids="$pids $!" # Configure every node in background
done
wait $pids # Wait for all configurations to end
