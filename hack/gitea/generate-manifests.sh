#!/bin/bash
set -e

INSTALL_YAML="pkg/controllers/localbuild/resources/gitea/k8s/install.yaml"
GITEA_DIR="./hack/gitea"
CHART_VERSION="12.1.2"

# By default, skip generation to avoid download failures behind firewalls
# Set SKIP_GITEA_MANIFEST_GENERATION=false to force regeneration
SKIP_GITEA_MANIFEST_GENERATION="${SKIP_GITEA_MANIFEST_GENERATION:-true}"

if [ "$SKIP_GITEA_MANIFEST_GENERATION" = "true" ]; then
  if [ -f "$INSTALL_YAML" ]; then
    echo "Skipping gitea manifest generation (SKIP_GITEA_MANIFEST_GENERATION=true)"
    echo "To regenerate, run: SKIP_GITEA_MANIFEST_GENERATION=false $0"
    exit 0
  else
    echo "Warning: $INSTALL_YAML does not exist and SKIP_GITEA_MANIFEST_GENERATION=true"
    echo "Attempting to generate manifests..."
  fi
fi

# Use a temporary file to avoid corrupting the original if generation fails
TEMP_YAML="${INSTALL_YAML}.tmp"

echo "# GITEA INSTALL RESOURCES" >${TEMP_YAML}
echo "# This file is auto-generated with 'hack/gitea/generate-manifests.sh'" >>${TEMP_YAML}

# Add retry logic for helm repo operations
MAX_RETRIES=3
RETRY_COUNT=0
until helm repo add gitea-charts --force-update https://dl.gitea.com/charts/; do
  RETRY_COUNT=$((RETRY_COUNT+1))
  if [ $RETRY_COUNT -ge $MAX_RETRIES ]; then
    echo "Failed to add helm repo after $MAX_RETRIES attempts"
    echo "Download may be blocked by firewall. Keeping existing $INSTALL_YAML"
    rm -f "${TEMP_YAML}"
    exit 1
  fi
  echo "Retrying helm repo add... (attempt $((RETRY_COUNT+1))/$MAX_RETRIES)"
  sleep 2
done

helm repo update

# Use --kube-context="" to ensure helm doesn't try to use any k8s cluster context
# This prevents helm from detecting the CI runner's namespace (arc-runners)
helm template my-gitea gitea-charts/gitea \
  -f ${GITEA_DIR}/values.yaml \
  --version ${CHART_VERSION} \
  --kube-context="" \
  --namespace=default >>${TEMP_YAML}

# Remove the third line (helm template comment) and replace namespace
sed '3d' ${TEMP_YAML} | sed 's/namespace: default/namespace: gitea/g' > ${TEMP_YAML}.2

# helm template for pvc uses Release.namespace which doesn't get set
# when running the helm "template" command
# See: https://gitea.com/gitea/helm-chart/issues/630
# and: https://gitea.com/gitea/helm-chart/src/commit/3b2b700441e91a19a535e05de3a9eab2fef0b117/templates/gitea/pvc.yaml#L6
# and: https://github.com/helm/helm/issues/3553#issuecomment-1186518158
# and: https://github.com/splunk/splunk-connect-for-kubernetes/pull/790

cat ${GITEA_DIR}/ingress.yaml.tmpl >>${TEMP_YAML}.2

# Move temp file to final location only if everything succeeded
mv ${TEMP_YAML}.2 ${INSTALL_YAML}
rm -f ${TEMP_YAML}

echo "Successfully generated $INSTALL_YAML"
