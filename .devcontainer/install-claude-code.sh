#!/bin/bash

# Script to install claude-code CLI tool
set -e

echo "🤖 Installing claude-code CLI tool..."

# Source Node.js environment from common locations
if [ -f "/usr/local/share/nvm/nvm.sh" ]; then
    echo "🔧 Sourcing NVM environment..."
    export NVM_DIR="/usr/local/share/nvm"
    source "$NVM_DIR/nvm.sh"
fi

# Add common Node.js paths to PATH
export PATH="/usr/local/share/nodejs/bin:/usr/local/bin:$PATH"

# Wait for npm to become available (with timeout)
echo "⏳ Waiting for npm to become available..."
for i in {1..30}; do
    if command -v npm >/dev/null 2>&1; then
        echo "✅ npm found!"
        break
    fi
    echo "⏳ Attempt $i/30: npm not yet available, waiting 2 seconds..."
    sleep 2
done

# Install claude-code globally via npm
if command -v npm >/dev/null 2>&1; then
    echo "📦 Installing claude-code via npm..."
    npm install -g @anthropic-ai/claude-code
    echo "✅ claude-code installation completed"
else
    echo "❌ Error: npm still not found after waiting. Node.js may not be properly installed."
    echo "ℹ️  Available commands:"
    which node || echo "  - node: not found"
    which npm || echo "  - npm: not found"
    echo "ℹ️  Current PATH: $PATH"
    exit 1
fi

# Verify installation
if command -v claude-code >/dev/null 2>&1; then
    echo "🎉 claude-code is now available in PATH"
    claude-code --version
else
    echo "⚠️  Warning: claude-code may not be in PATH yet, but installation completed"
fi
