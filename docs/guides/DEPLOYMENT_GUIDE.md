# DeployKubeStack Operator - Deployment & Testing Guide

## Architecture Summary

The DeployKubeStack operator is now a working Kubernetes controller that implements Phase 1: Deployment and Service management.

### Components Implemented

```
Application CR (input)
    ↓
    ├→ ApplicationReconciler (main loop)
    ├→ ReconcileDeployment (creates/updates Deployment)
    ├→ ReconcileService (creates/updates Service)
    └→ Status Updates (phase, ready flags)
```

## System Requirements

- **Kubernetes**: 1.20+ (tested with 1.33)
- **Go**: 1.24.0+ (for building)
- **Docker/Podman**: For building container images
- **Kind/Minikube**: For local testing
- **kubectl**: For cluster interaction

## Application CRD Reference

### Minimal Example (All Defaults)

```yaml
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: simple-app
  namespace: default
spec:
  image: nginx:latest
  port: 80
```

**What gets created:**
- Deployment with 1 replica (default)
- Service on port 80
- Pod resources: 100m CPU, 100Mi memory (defaults)

### Full Example (All Options)

```yaml
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: full-app
  namespace: default
  labels:
    team: platform
    env: staging
spec:
  image: myregistry.azurecr.io/myapp:v1.2.3
  port: 3000
  replicas: 5
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

**What gets created:**
- Deployment with 5 replicas
- Service on port 3000
- Pod resources: 200m-1000m CPU, 256Mi-1Gi memory

## Deployment Scenarios

### Scenario 1: Local Development Testing

#### Step 1: Set up test cluster

```bash
# Install kind if not present
curl -Lo ./kind https://kind.sigs.k8s.io/dl/latest/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Create cluster
kind create cluster --name deploykubestack

# Verify
kubectl cluster-info
kubectl get nodes
```

#### Step 2: Build and install operator

```bash
cd /home/vnmrk7788/DeployKubeStack

# Install CRD into cluster
make install

# Build operator binary
make build

# Run operator locally (connects to cluster)
make run
```

Output should show:
```
setup    Starting manager
```

#### Step 3: Deploy an application

In another terminal:

```bash
kubectl apply -f config/samples/apps_v1alpha1_application.yaml
```

#### Step 4: Verify

```bash
# Check Application
kubectl get application application-sample
kubectl describe application application-sample

# Check Deployment
kubectl get deployment -l app=application-sample
kubectl logs deployment/application-sample

# Check Service
kubectl get service -l app=application-sample
kubectl get endpoints -l app=application-sample

# Check Pods
kubectl get pods -l app=application-sample
```

#### Step 5: Test updates

```bash
# Scale replicas
kubectl patch application application-sample -p '{"spec":{"replicas":3}}'
kubectl get deployment application-sample -o jsonpath='{.spec.replicas}'

# Update image
kubectl patch application application-sample --type merge -p '{"spec":{"image":"nginx:alpine"}}'
kubectl get deployment application-sample -o jsonpath='{.spec.template.spec.containers[0].image}'

# Delete
kubectl delete application application-sample
kubectl get deployment application-sample  # Should be gone
```

### Scenario 2: Container Deployment to Real Cluster

#### Step 1: Build and push container image

```bash
# Build
make docker-build IMG=myregistry.azurecr.io/deploykubestack:v1.0

# Push (requires auth to registry)
make docker-push IMG=myregistry.azurecr.io/deploykubestack:v1.0
```

#### Step 2: Deploy to cluster

```bash
# Deploy operator to cluster (creates Deployment, Service, RBAC, etc.)
make deploy IMG=myregistry.azurecr.io/deploykubestack:v1.0

# Verify operator is running
kubectl get deployment -n deploykubestack-system
kubectl get pod -n deploykubestack-system
kubectl logs -n deploykubestack-system -l control-plane=controller-manager -f
```

#### Step 3: Create applications

```bash
# Create sample application
kubectl apply -f - <<EOF
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  image: nginx:latest
  port: 80
  replicas: 3
EOF

# Verify
kubectl get applications
kubectl get deployment,service -l app=my-app
```

#### Step 4: Update and delete

```bash
# Update
kubectl patch application my-app -p '{"spec":{"replicas":5}}'

# Delete (everything cleaned up automatically)
kubectl delete application my-app
```

## Testing Guide

### Unit Tests

```bash
# Run all tests
make test

# Expected output:
# Running Suite: Controller Suite
# Will run 3 of 3 specs
# ✓ Application Controller when creating an Application with minimal spec should create Deployment and Service with defaults
# ✓ Application Controller when creating an Application with custom replicas and resources should create Deployment with custom values
# ✓ Application Controller when Application is deleted should delete Deployment and Service
# 3 Passed | 0 Failed
```

### Manual E2E Testing

```bash
# Create test cluster
kind create cluster --name e2e-test

# Install operator
make install
make run &
OPERATOR_PID=$!

# Test 1: Create with defaults
kubectl apply -f - <<EOF
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: test-defaults
spec:
  image: nginx:latest
  port: 80
EOF

# Verify defaults applied
kubectl get deployment test-defaults -o jsonpath='{.spec.replicas}'  # Should be 1
kubectl get deployment test-defaults -o jsonpath='{.spec.template.spec.containers[0].resources.requests.cpu}'  # Should be 100m

# Test 2: Create with custom values
kubectl apply -f - <<EOF
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: test-custom
spec:
  image: nginx:latest
  port: 8080
  replicas: 3
  resources:
    requests:
      cpu: 250m
      memory: 512Mi
    limits:
      cpu: 500m
      memory: 1Gi
EOF

# Verify custom values applied
kubectl get deployment test-custom -o jsonpath='{.spec.replicas}'  # Should be 3
kubectl get deployment test-custom -o jsonpath='{.spec.template.spec.containers[0].resources.requests.cpu}'  # Should be 250m

# Test 3: Scale
kubectl patch application test-custom -p '{"spec":{"replicas":5}}'
sleep 5
kubectl get deployment test-custom -o jsonpath='{.spec.replicas}'  # Should be 5

# Test 4: Update image
kubectl patch application test-custom --type merge -p '{"spec":{"image":"nginx:alpine"}}'
sleep 5
kubectl get deployment test-custom -o jsonpath='{.spec.template.spec.containers[0].image}'  # Should be nginx:alpine

# Test 5: Garbage collection
kubectl delete application test-custom
sleep 5
kubectl get deployment test-custom  # Should error (not found)

# Cleanup
kill $OPERATOR_PID
kind delete cluster --name e2e-test
```

## Troubleshooting

### Tests failing with timeout

**Problem**: Tests timeout waiting for Application phase to be "Ready"

**Solution**: Ensure the controller manager is started in the test suite
```bash
# Check suite_test.go has manager setup in BeforeSuite()
cat internal/controller/suite_test.go | grep "NewManager"
```

### Operator not creating Deployment/Service

**Problem**: Application created but Deployment/Service not appearing

**Troubleshooting**:
```bash
# Check operator logs
kubectl logs -n deploykubestack-system -l control-plane=controller-manager

# Check Application status
kubectl describe application my-app

# Check for RBAC issues
kubectl auth can-i create deployments --as=system:serviceaccount:deploykubestack-system:deploykubestack-controller-manager

# Check CRD is installed
kubectl get crd applications.apps.deploykubestack.com
```

### Deployment not using custom resources

**Problem**: Created with custom resources but Deployment shows defaults

**Solution**: Verify Application spec was actually updated
```bash
# Check what operator sees
kubectl get application my-app -o yaml | grep -A 10 "resources:"

# Verify Deployment spec
kubectl get deployment my-app -o yaml | grep -A 20 "resources:"
```

## Performance Considerations

### Resource Overhead

The operator itself requires minimal resources:

```yaml
requests:
  cpu: 100m
  memory: 64Mi
limits:
  cpu: 500m
  memory: 200Mi
```

### Reconciliation Frequency

- Triggered on Application CR changes
- Triggered on owned Deployment/Service changes
- No continuous polling (event-driven)
- Exponential backoff on errors (1s, 2s, 4s, ...)

### Scalability

Tested scenarios:
- Single Application with multi-replica Deployment ✅
- Multiple Applications in same cluster ✅
- Applications across multiple namespaces ✅
- Rapid Application creation/deletion ✅

## Security Considerations

### RBAC

The operator runs with least privilege:
- Can create/update/delete only Deployments and Services
- Cannot access Secrets, ConfigMaps, or other resources
- Per-namespace isolation via namespace-scoped RBAC

### Network Policies

The operator network communication:
- Outbound: Only to Kubernetes API server
- Inbound: Only webhook endpoints (if enabled, currently disabled)
- No egress to external services

### CRD Validation

All inputs validated:
- Image: Required, non-empty string
- Port: 1-65535
- Replicas: 1-1000
- Resources: Must follow Kubernetes format

## Monitoring & Observability

### Logging

The operator logs reconciliation events:

```
INFO reconciled Deployment operation=created
INFO reconciled Service operation=updated
INFO Application reconciliation completed successfully
ERROR failed to reconcile Deployment
```

### Status Monitoring

```bash
# Watch status updates
kubectl get application --watch

# Check status details
kubectl get application -o wide
kubectl get application my-app -o jsonpath='{.status}'
```

### Metrics

Controller-runtime exports Prometheus metrics:
- `controller_runtime_reconcile_total` - reconciliation count
- `controller_runtime_reconcile_errors_total` - error count
- `controller_runtime_reconcile_duration_seconds` - duration histogram

## Upgrade Path

### From Phase 1 to Phase 2+

New resources (HPA, Ingress, etc.) can be added without breaking existing Applications:
- Existing fields unchanged
- New fields optional (non-breaking)
- Automatic reconciliation applies new resources

Example future upgrade:

```yaml
spec:
  image: nginx:latest
  port: 80
  # Existing fields ↑
  
  # New Phase 2 feature
  hpa:
    enabled: true
    minReplicas: 1
    maxReplicas: 10
    targetCPUUtilizationPercentage: 80
```

## Support & Contributing

### Getting Help

1. Check [QUICK_START.md](QUICK_START.md) for common scenarios
2. Review [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for design docs
3. Check operator logs: `make run` output
4. Run tests: `make test`

### Reporting Issues

When reporting issues include:
```bash
# Application manifest
kubectl get application <name> -o yaml

# Application status
kubectl describe application <name>

# Created resources
kubectl get deployment,service -l app=<name>

# Operator logs
# (from 'make run' terminal or kubectl logs if deployed)
```

---

**Status**: Phase 1 (Deployment + Service) complete and tested.
Ready for production use or further development.
