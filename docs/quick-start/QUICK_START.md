# DeployKubeStack Operator - Quick Start Guide

## Prerequisites

- Kubernetes cluster (kind, minikube, or remote)
- kubectl configured
- Go 1.24.0+
- Docker/Podman for building images

## Set Up a Local Test Cluster

```bash
# Using kind (recommended)
kind create cluster --name deploykubestack
kubectl cluster-info

# Or minikube
minikube start
```

## Build & Deploy the Operator

```bash
# Clone and navigate to repo
cd /home/vnmrk7788/DeployKubeStack

# Install CRD into cluster
make install

# Start operator locally (watches your cluster)
make run

# In another terminal, apply a sample Application
kubectl apply -f config/samples/apps_v1alpha1_application.yaml
```

## Verify It Works

```bash
# Check Application status
kubectl get applications
kubectl describe application application-sample

# See created resources
kubectl get deployment,service -l app=application-sample

# Check operator logs
# (in the terminal where you ran 'make run')
```

## Example: Create Your Own Application

### Minimal (Using Defaults)

```bash
kubectl apply -f - <<EOF
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  image: nginx:latest
  port: 80
EOF

# This creates:
# - Deployment with 1 replica
# - Service on port 80
# - Default resources: 100m CPU, 100Mi memory
```

### With Custom Resources

```bash
kubectl apply -f - <<EOF
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: my-app-custom
spec:
  image: nginx:1.21
  port: 8080
  replicas: 3
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 500m
      memory: 512Mi
EOF
```

## Common Operations

### View Application Details

```bash
kubectl get application my-app
kubectl describe application my-app
kubectl get application my-app -o yaml
```

### Scale the Application

```bash
kubectl patch application my-app -p '{"spec":{"replicas":5}}'
```

### Update the Image

```bash
kubectl patch application my-app --type merge -p '{"spec":{"image":"nginx:alpine"}}'
```

### Delete Application (Automatic Cleanup)

```bash
kubectl delete application my-app
# Deployment and Service are automatically deleted
```

## Verify Deployment & Service

```bash
# See Deployment
kubectl get deployment -l app=my-app
kubectl describe deployment my-app
kubectl logs deployment/my-app

# See Service
kubectl get service -l app=my-app
kubectl describe service my-app

# Port forward to test
kubectl port-forward svc/my-app 8080:8080
# Visit http://localhost:8080
```

## Troubleshooting

### Check operator logs

```bash
# If running locally
# Look at 'make run' terminal output

# If deployed to cluster
kubectl logs -n deploykubestack-system deployment/deploykubestack-controller-manager
```

### Check Application status

```bash
kubectl describe application my-app
# Look for Status.Phase (should be "Ready")
# Look for Message (error details if phase is "Degraded")
```

### Check if Deployment/Service were created

```bash
kubectl get deployment,service -l app=my-app
```

## Run Tests

```bash
# Unit tests (requires envtest binaries)
make test

# Build operator binary
make build

# Run linter
make lint
```

## Stop the Operator

In the terminal where `make run` is executing:
```
Ctrl+C
```

Then uninstall CRDs:
```bash
make uninstall
```

## What Gets Created

When you create an Application, the operator automatically creates:

1. **Deployment** with:
   - Pod replicas (default: 1, or custom)
   - Container image (from spec)
   - Port (from spec)
   - Resources (defaults: 100m CPU, 100Mi memory)
   - Labels: `app=<app-name>`, `managed-by=deploykubestack`

2. **Service** with:
   - Type: ClusterIP (internal, stable DNS name)
   - Port mapping to the Deployment
   - Selector matching the Deployment pods
   - Labels: `app=<app-name>`, `managed-by=deploykubestack`

Both resources are automatically deleted when you delete the Application CR.

## Default Resource Values

When `resources` field is not specified:

```yaml
resources:
  requests:
    cpu: 100m
    memory: 100Mi
  limits:
    cpu: 100m
    memory: 100Mi
```

To override, specify your own:

```yaml
spec:
  image: myapp:1.0
  port: 3000
  resources:
    requests:
      cpu: 250m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

## Next Steps

See [IMPLEMENTATION_SUMMARY.md](IMPLEMENTATION_SUMMARY.md) for:
- Detailed architecture
- Implementation details
- Test descriptions
- Future phases plan
