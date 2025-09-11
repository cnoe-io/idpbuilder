# Gitea Client Demo Documentation

This document describes the demo features for the idpbuilder gitea-client implementation.

## Overview

The gitea-client is a core component of idpbuilder that provides Git repository functionality through Gitea integration. This demo showcases the key features and capabilities implemented in this effort.

## Demo Script

Run the demo using:

```bash
./demo-features.sh
```

## Demo Features

### 1. IDP Builder CLI Help
- Demonstrates the main idpbuilder command interface
- Shows available commands: create, get, delete
- Provides overview of the tool's purpose

### 2. Gitea Integration Configuration
- Shows key configuration constants
- Demonstrates namespace and credential management
- URL templating for different routing modes

### 3. Certificate Management
- Kind cluster certificate extraction
- Custom CA trust store management
- Registry TLS configuration
- Insecure mode with security logging
- Certificate rotation and reloading

### 4. Package Structure
- Overview of the modular package architecture
- Shows organization of functionality
- Demonstrates separation of concerns

### 5. Build System Integration
- Local build support
- Kubernetes integration
- CoreDNS configuration
- TLS certificate management
- Container registry operations

### 6. Configuration Management
- Feature flag support
- Environment-based configuration
- Path routing vs subdomain routing
- Protocol and port configuration
- Custom host support

### 7. Authentication & Token Management
- Gitea admin token creation/deletion
- Access token management with scopes
- Basic authentication support
- HTTP client configuration
- Password and token patching

### 8. Security & Audit
- Security decision logging
- Audit trail for certificate operations
- File permission management (0600/0700)
- Explicit insecure mode warnings
- Certificate validation and expiry checking

### 9. Test Data Examples
- Sample configuration files
- Test scenarios
- Integration examples

## Technical Architecture

### Core Components

1. **Gitea Client (`pkg/util/gitea.go`)**
   - Manages Gitea API interactions
   - Handles authentication and token management
   - Provides repository operations

2. **Certificate Management (`pkg/certs/`)**
   - Extracts certificates from Kind clusters
   - Manages trust stores and CA pools
   - Configures TLS for registry connections

3. **Build System (`pkg/build/`)**
   - Coordinates build operations
   - Integrates with Kubernetes
   - Manages CoreDNS and TLS setup

4. **Configuration (`pkg/config/`)**
   - Feature flag management
   - Environment configuration
   - Routing and protocol settings

### Key Features

- **Secure by Default**: All certificate operations require explicit configuration
- **Feature Flag Controlled**: Components can be enabled/disabled independently  
- **Audit Logging**: Security decisions are logged for compliance
- **Kubernetes Native**: Built for Kubernetes environments
- **Modular Design**: Components can be used independently

## Prerequisites

- Docker (for container operations)
- Kind cluster (for certificate extraction)
- Kubernetes access (for operations)

## Configuration

The demo shows various configuration options:

```bash
# Feature flags
export IDPBUILDER_CERT_INFRASTRUCTURE_ENABLED=true
export IDPBUILDER_KIND_CERT_EXTRACTION_ENABLED=true
export IDPBUILDER_REGISTRY_TLS_TRUST_ENABLED=true

# Configuration directory
export IDPBUILDER_CONFIG_DIR=~/.idpbuilder

# Security settings
export IDPBUILDER_TLS_INSECURE=false
```

## Demo Validation

The demo script:
- ✅ Exits with code 0 on success
- ✅ Shows all major features
- ✅ Demonstrates security considerations
- ✅ Provides integration hooks
- ✅ Includes timestamp tracking

## Integration Points

The demo includes integration hooks:
- `DEMO_READY=true` environment variable
- Exit code 0 for success validation
- Timestamp logging for coordination
- Structured output for automation

## Security Considerations

The demo emphasizes:
- Explicit security decisions
- Audit trail requirements
- Certificate validation
- Insecure mode warnings
- File permission management

## Next Steps

After running the demo:
1. Examine the test-data directory for examples
2. Review the .demo-config for settings
3. Check security.log for audit entries
4. Verify integration readiness