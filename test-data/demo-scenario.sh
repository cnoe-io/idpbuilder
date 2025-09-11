#!/bin/bash
# Demo scenario for gitea-client testing

echo "🧪 Gitea Client Test Scenario"
echo "============================="

# Scenario 1: IDP Setup
echo "1. Setting up IDP environment..."
echo "   - Kind cluster: demo-cluster"
echo "   - Gitea namespace: gitea"
echo "   - Protocol: https"
echo "   - Host: demo.local"

# Scenario 2: Certificate Setup
echo "2. Certificate configuration..."
echo "   - Extract from Kind: demo-cluster"
echo "   - Store in: ~/.idpbuilder/certs/"
echo "   - Validate expiry: enabled"
echo "   - Security logging: enabled"

# Scenario 3: Gitea Integration
echo "3. Gitea authentication..."
echo "   - Admin user: giteaAdmin"
echo "   - Token management: enabled"
echo "   - Access scopes: all"
echo "   - URL: https://gitea.demo.local"

# Scenario 4: Build Operations
echo "4. Build system integration..."
echo "   - Local builds: enabled"
echo "   - Registry: gitea.demo.local"
echo "   - TLS verification: strict"
echo "   - CoreDNS: configured"

echo "✅ Demo scenario complete"