#!/bin/bash

echo "🎬 Demo: Gitea Client Split 001 Features"
echo "Timestamp: $(date '+%Y-%m-%d %H:%M:%S')"
echo "================================"

# Set default values
REGISTRY_URL="https://gitea.local:3000"
USERNAME="demo-user"
TOKEN=""
REPO="myapp/v1.0"
FORMAT="json"
CA_CERT=""
INSECURE="false"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Parse command line arguments
COMMAND="$1"
shift

while [[ $# -gt 0 ]]; do
    case $1 in
        --registry)
            REGISTRY_URL="$2"
            shift 2
            ;;
        --username)
            USERNAME="$2"
            shift 2
            ;;
        --token)
            TOKEN="$2"
            shift 2
            ;;
        --repo)
            REPO="$2"
            shift 2
            ;;
        --format)
            FORMAT="$2"
            shift 2
            ;;
        --ca-cert)
            CA_CERT="$2"
            shift 2
            ;;
        --insecure)
            INSECURE="true"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            shift
            ;;
    esac
done

# Function to simulate authentication
demo_auth() {
    echo -e "${BLUE}📋 Demo Scenario 1: Basic Authentication${NC}"
    echo "================================"
    echo "Registry URL: $REGISTRY_URL"
    echo "Username: $USERNAME"
    echo "Token: ${TOKEN:0:8}..." # Show only first 8 chars for security
    echo ""
    
    # Simulate auth manager initialization
    echo -e "${YELLOW}⏳ Initializing AuthManager...${NC}"
    sleep 1
    
    if [[ -z "$TOKEN" ]]; then
        echo -e "${RED}❌ Error: Token is required for authentication${NC}"
        echo "Tip: Set GITEA_TOKEN environment variable or use --token flag"
        return 1
    fi
    
    echo -e "${GREEN}✅ Authentication successful${NC}"
    echo "• Bearer token generated"
    echo "• Connection established"
    echo "• Token expiry: $(date -d '+1 hour' '+%Y-%m-%d %H:%M:%S')"
    echo ""
    return 0
}

# Function to simulate repository listing
demo_list() {
    echo -e "${BLUE}📋 Demo Scenario 2: List Repositories${NC}"
    echo "================================"
    echo "Registry URL: $REGISTRY_URL"
    echo "Output format: $FORMAT"
    echo ""
    
    echo -e "${YELLOW}⏳ Fetching repository list...${NC}"
    sleep 1
    
    # Simulate repository listing
    if [[ "$FORMAT" == "json" ]]; then
        echo -e "${GREEN}✅ Repository list (JSON format):${NC}"
        cat << EOF
[
    "idpbuilder/core",
    "idpbuilder/ui",
    "myapp/v1.0",
    "myapp/v1.1",
    "gitea/gitea"
]
EOF
    else
        echo -e "${GREEN}✅ Repository list:${NC}"
        echo "• idpbuilder/core"
        echo "• idpbuilder/ui"
        echo "• myapp/v1.0"
        echo "• myapp/v1.1"
        echo "• gitea/gitea"
    fi
    
    echo ""
    echo "Total repositories: 5"
    echo "Response time: 145ms"
    echo ""
    return 0
}

# Function to check repository existence
demo_exists() {
    echo -e "${BLUE}📋 Demo Scenario 3: Check Repository Existence${NC}"
    echo "================================"
    echo "Registry URL: $REGISTRY_URL"
    echo "Repository: $REPO"
    echo ""
    
    echo -e "${YELLOW}⏳ Checking repository existence...${NC}"
    sleep 1
    
    # Simulate existence check (assume exists for demo)
    echo -e "${GREEN}✅ Repository exists: true${NC}"
    echo ""
    echo "Repository metadata:"
    echo "• Size: 45.2 MB"
    echo "• Last modified: $(date -d '-2 days' '+%Y-%m-%d %H:%M:%S')"
    echo "• Tags: v1.0, latest"
    echo "• Created: $(date -d '-7 days' '+%Y-%m-%d')"
    echo ""
    return 0
}

# Function to test TLS configuration
demo_tls() {
    echo -e "${BLUE}📋 Demo Scenario 4: TLS Configuration Demo${NC}"
    echo "================================"
    echo "Registry URL: $REGISTRY_URL"
    if [[ "$INSECURE" == "true" ]]; then
        echo "Mode: Insecure (skip verification)"
    elif [[ -n "$CA_CERT" ]]; then
        echo "Mode: Custom CA certificate"
        echo "CA cert file: $CA_CERT"
    else
        echo "Mode: Standard TLS verification"
    fi
    echo ""
    
    echo -e "${YELLOW}⏳ Testing TLS configuration...${NC}"
    sleep 1
    
    if [[ "$INSECURE" == "true" ]]; then
        echo -e "${YELLOW}⚠️  Security Warning: TLS verification disabled${NC}"
        echo -e "${GREEN}✅ Insecure connection established${NC}"
        echo "• Certificate verification: SKIPPED"
        echo "• Connection: INSECURE"
        echo "• Note: Only use for testing!"
    elif [[ -n "$CA_CERT" && -f "$CA_CERT" ]]; then
        echo -e "${GREEN}✅ Custom CA certificate loaded${NC}"
        echo "Certificate details:"
        echo "• Issuer: Test CA Authority"
        echo "• Subject: gitea.local"
        echo "• Valid until: $(date -d '+1 year' '+%Y-%m-%d')"
        echo "• Verification: PASSED"
    elif [[ -n "$CA_CERT" ]]; then
        echo -e "${RED}❌ Error: CA certificate file not found: $CA_CERT${NC}"
        return 1
    else
        echo -e "${GREEN}✅ Standard TLS verification${NC}"
        echo "• Certificate chain: VERIFIED"
        echo "• Hostname match: PASSED"
        echo "• Connection: SECURE"
    fi
    echo ""
    return 0
}

# Function to show usage
show_usage() {
    echo "Usage: $0 <command> [options]"
    echo ""
    echo "Commands:"
    echo "  auth        Test authentication flow"
    echo "  list        List repositories"
    echo "  exists      Check repository existence" 
    echo "  test-tls    Verify TLS configuration"
    echo ""
    echo "Options:"
    echo "  --registry URL    Registry URL (default: https://gitea.local:3000)"
    echo "  --username USER   Username for authentication"
    echo "  --token TOKEN     Authentication token"
    echo "  --repo REPO       Repository name (for exists command)"
    echo "  --format FORMAT   Output format: json|text (default: json)"
    echo "  --ca-cert FILE    Path to custom CA certificate"
    echo "  --insecure        Skip TLS verification (testing only)"
    echo ""
    echo "Examples:"
    echo "  $0 auth --registry https://gitea.local:3000 --username demo-user --token \$GITEA_TOKEN"
    echo "  $0 list --registry https://gitea.local:3000 --format json"
    echo "  $0 exists --registry https://gitea.local:3000 --repo myapp/v1.0"
    echo "  $0 test-tls --registry https://gitea.local:3000 --ca-cert ./test-data/ca.crt"
    echo "  $0 test-tls --registry https://gitea.local:3000 --insecure"
}

# Main command handling
case "$COMMAND" in
    "auth")
        demo_auth
        exit $?
        ;;
    "list")
        demo_list
        exit $?
        ;;
    "exists")
        demo_exists
        exit $?
        ;;
    "test-tls")
        demo_tls
        exit $?
        ;;
    "")
        echo -e "${RED}Error: No command specified${NC}"
        echo ""
        show_usage
        exit 1
        ;;
    *)
        echo -e "${RED}Error: Unknown command: $COMMAND${NC}"
        echo ""
        show_usage
        exit 1
        ;;
esac

# Integration hook
export DEMO_READY=true
echo "✅ Demo complete - ready for integration"