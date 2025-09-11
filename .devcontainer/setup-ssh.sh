#!/bin/bash

# Script to setup SSH for GitHub access
set -e

echo "🔐 Setting up SSH for GitHub access..."

# Add GitHub to known hosts
mkdir -p ~/.ssh
ssh-keyscan -H github.com >> ~/.ssh/known_hosts 2>/dev/null || echo "Warning: Could not add github.com to known hosts"

# Set proper permissions
chmod 700 ~/.ssh 2>/dev/null || true
chmod 600 ~/.ssh/known_hosts 2>/dev/null || true

echo "✅ SSH setup completed"
