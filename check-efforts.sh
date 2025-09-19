#!/bin/bash

for effort in cert-validation-split-001 cert-validation-split-002 cert-validation-split-003 fallback-strategies; do
  echo "=== $effort ==="
  cd /home/vscode/workspaces/this-is-not-the-target-repo-this-is-for-orchestrator-planning-only/efforts/phase1/wave2/$effort
  BRANCH=$(git branch --show-current)
  echo "Branch: $BRANCH"
  LOCAL_SHA=$(git rev-parse HEAD)
  REMOTE_SHA=$(git ls-remote origin refs/heads/phase1/wave2/$effort 2>/dev/null | cut -f1)
  if [ -z "$REMOTE_SHA" ]; then
    echo "Status: NOT ON REMOTE (needs push)"
    echo "Latest commit: $(git log --oneline -1)"
  elif [ "$LOCAL_SHA" = "$REMOTE_SHA" ]; then
    echo "Status: UP TO DATE with remote"
  else
    echo "Status: LOCAL differs from remote"
  fi
  echo ""
done