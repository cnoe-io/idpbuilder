#!/bin/bash

# Script to install claude-code CLI tool
set -e

echo "🤖 Installing claude-code CLI tool..."

# Install claude-code globally via npm
if command -v npm >/dev/null 2>&1; then
    npm install -g @anthropic-ai/claude-code
    echo "✅ claude-code installation completed"
else
    echo "❌ Error: npm not found. Node.js is required to install claude-code"
    exit 1
fi

# Verify installation
if command -v claude-code >/dev/null 2>&1; then
    echo "🎉 claude-code is now available in PATH"
    claude-code --version
else
    echo "⚠️  Warning: claude-code may not be in PATH yet, but installation completed"
fi
