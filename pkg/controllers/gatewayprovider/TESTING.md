# NginxGateway Functional Tests

This directory contains functional and integration tests for the NginxGateway controller.

## Test Files

1. **nginxgateway_functional_test.go**: Unit/functional tests using fake Kubernetes client
   - Tests reconciliation logic
   - Tests status updates
   - Tests deletion handling
   - Can be run without a real cluster

2. **nginxgateway_integration_test.go**: Integration tests requiring a real Kubernetes cluster
   - Tests end-to-end NginxGateway creation and deployment
   - Verifies nginx resources are actually created
   - Validates status aggregation
   - Requires `integration` build tag

## Running Tests

### Functional Tests (No Cluster Required)

```bash
# Run all functional tests
go test -v ./pkg/controllers/gatewayprovider/...

# Run specific test
go test -v ./pkg/controllers/gatewayprovider/... -run TestNginxGatewayStatusUpdate
```

### Integration Tests (Cluster Required)

```bash
# Requires a running Kubernetes cluster with:
# - NginxGateway CRD installed
# - NginxGateway controller running

# Run integration tests
go test -v -tags=integration ./pkg/controllers/gatewayprovider/... -run TestNginxGatewayIntegration

# Run E2E test
go test -v -tags=integration ./pkg/controllers/gatewayprovider/... -run TestNginxGatewayE2E
```

## Test Coverage

### Functional Tests

- **TestNginxGatewayFunctional**: Basic reconciliation flow
  - Creates NginxGateway resource
  - Triggers reconciliation
  - Verifies initial status

- **TestNginxGatewayResourcesCreated**: Resource creation validation
  - Validates nginx resources are created
  - Tests reconciliation completes without errors

- **TestNginxGatewayStatusUpdate**: Status field validation
  - Verifies duck-typed status fields are populated
  - Tests Ready condition
  - Validates controller status (replicas)

- **TestNginxGatewayDeletion**: Finalizer handling
  - Tests deletion flow
  - Verifies finalizer is removed

### Integration Tests

- **TestNginxGatewayIntegration**: Full deployment validation
  - Creates NginxGateway in real cluster
  - Waits for nginx deployment to be ready
  - Verifies all status fields
  - Validates nginx service exists

- **TestNginxGatewayE2E**: End-to-end lifecycle
  - Creates and deletes NginxGateway
  - Tests complete lifecycle

## What Gets Validated

### Status Fields (Duck-Typed)

The following fields match the JSON field names in the API:

- `ingressClassName`: The ingress class name to use
- `loadBalancerEndpoint`: External endpoint for accessing services
- `internalEndpoint`: Cluster-internal API endpoint
- `phase`: Current phase (Installing, Ready, Failed)
- `installed`: Whether nginx is installed
- `version`: Currently installed version
- `controller.replicas`: Desired number of replicas
- `controller.readyReplicas`: Number of ready replicas

### Conditions

- `Ready`: True when nginx is fully operational

### Deployed Resources

- Nginx Ingress Controller Deployment
- Nginx Ingress Controller Service  
- IngressClass resource
- RBAC resources (ServiceAccount, ClusterRole, ClusterRoleBinding)

## Example Test Output

```
=== RUN   TestNginxGatewayIntegration
    nginxgateway_integration_test.go:54: Creating NginxGateway resource...
    nginxgateway_integration_test.go:64: Waiting for NginxGateway to be reconciled...
    nginxgateway_integration_test.go:75: NginxGateway status: Phase=Installing, Installed=false
    nginxgateway_integration_test.go:75: NginxGateway status: Phase=Installing, Installed=false
    nginxgateway_integration_test.go:75: NginxGateway status: Phase=Ready, Installed=true
    nginxgateway_integration_test.go:95: Verifying NginxGateway status...
    nginxgateway_integration_test.go:123: Verifying nginx deployment exists...
    nginxgateway_integration_test.go:130: Deployment status: Replicas=1, ReadyReplicas=1
    nginxgateway_integration_test.go:139: Verifying nginx service exists...
    nginxgateway_integration_test.go:146: Integration test completed successfully!
--- PASS: TestNginxGatewayIntegration (45.23s)
PASS
```

## Troubleshooting

### Integration Tests Fail to Find Resources

- Ensure NginxGateway CRD is installed: `kubectl get crd nginxgateways.idpbuilder.cnoe.io`
- Ensure NginxGateway controller is running: `kubectl get pods -n idpbuilder-system`
- Check controller logs: `kubectl logs -n idpbuilder-system <controller-pod> -f`

### Timeout Waiting for Ready State

- Increase poll timeout in test
- Check nginx deployment status: `kubectl get deployment -n ingress-nginx-test ingress-nginx-controller`
- Check for resource constraints or image pull errors

### Functional Tests Pass but Integration Tests Fail

- Functional tests use fake client and don't deploy real resources
- Integration tests require actual cluster resources
- Verify cluster has sufficient resources for nginx deployment
