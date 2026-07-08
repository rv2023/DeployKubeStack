# Helm Deployment Guide - DeployKubeStack Operator

## Overview

The DeployKubeStack operator can be deployed to Kubernetes using Helm charts. The Helm chart includes:

- ✅ Custom Resource Definition (CRD) for Application
- ✅ Controller Deployment
- ✅ Service Account with proper RBAC
- ✅ ClusterRole and ClusterRoleBinding
- ✅ Metrics service
- ✅ Health probes (liveness and readiness)

## Prerequisites

- **Kubernetes:** 1.20 or later
- **Helm:** 3.0 or later
- **Container Registry Access:** GHCR (default) or configure alternative registry

## Quick Start

### 1. Install the Operator

```bash
# Create namespace
kubectl create namespace deploykubestack-system

# Install from local chart
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system

# Or install with custom values
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --values custom-values.yaml
```

### 2. Verify Installation

```bash
# Check deployment
kubectl get deployment -n deploykubestack-system

# View logs
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager

# Check if CRD is installed
kubectl get crd | grep deploykubestack

# Verify RBAC
kubectl get clusterrole | grep deploykubestack
```

### 3. Create an Application

```yaml
apiVersion: apps.deploykubestack.com/v1alpha1
kind: Application
metadata:
  name: my-app
spec:
  image: nginx:latest
  port: 80
  replicas: 2
  resources:
    requests:
      cpu: 100m
      memory: 100Mi
    limits:
      cpu: 500m
      memory: 512Mi
```

```bash
# Apply the Application
kubectl apply -f application.yaml

# Monitor status
kubectl get applications
kubectl describe application my-app
```

## Configuration

### Common Values

**Image Configuration**
```yaml
image:
  registry: ghcr.io                    # Container registry
  repository: rv2023/DeployKubeStack   # Image name
  tag: ""                              # Auto-detects Chart.appVersion
  pullPolicy: IfNotPresent             # Always, IfNotPresent, Never
```

**Replica and Resource Configuration**
```yaml
replicaCount: 1                        # Number of operator instances

resources:
  requests:
    cpu: 100m                          # CPU request
    memory: 64Mi                       # Memory request
  limits:
    cpu: 500m                          # CPU limit
    memory: 128Mi                      # Memory limit
```

**CRD Configuration**
```yaml
crds:
  install: true                        # Whether to install CRD
  keep: true                           # Keep CRD on chart deletion
```

**Logging Configuration**
```yaml
logging:
  level: info                          # debug, info, warn, error
  development: false                   # Enable development mode
```

**Feature Flags**
```yaml
metricsServer:
  enabled: true                        # Expose Prometheus metrics
  port: 8080                           # Metrics server port

healthProbe:
  enabled: true                        # Enable liveness/readiness probes
  port: 8081                           # Health probe port

webhook:
  enabled: false                       # Enable webhooks (future)
```

### Custom Values Examples

**Production Deployment**
```yaml
# values-prod.yaml
replicaCount: 2

image:
  registry: private.registry.com
  pullPolicy: Always

resources:
  requests:
    cpu: 200m
    memory: 128Mi
  limits:
    cpu: 1000m
    memory: 512Mi

logging:
  level: warn

nodeSelector:
  kubernetes.io/os: linux

affinity:
  podAntiAffinity:
    preferredDuringSchedulingIgnoredDuringExecution:
    - weight: 100
      podAffinityTerm:
        labelSelector:
          matchExpressions:
          - key: app.kubernetes.io/name
            operator: In
            values:
            - deploykubestack
        topologyKey: kubernetes.io/hostname
```

**Development Deployment**
```yaml
# values-dev.yaml
replicaCount: 1

logging:
  level: debug
  development: true

resources:
  requests:
    cpu: 50m
    memory: 32Mi
  limits:
    cpu: 200m
    memory: 128Mi
```

## Deployment Commands

### Helm Install

```bash
# Basic install
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace

# Install with custom values
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values values-prod.yaml

# Install from OCI registry (when available)
helm install deploykubestack oci://ghcr.io/rv2023/deploykubestack-charts \
  --version 1.0.0 \
  --namespace deploykubestack-system \
  --create-namespace

# Install with individual value overrides
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --set replicaCount=2 \
  --set image.tag=v1.0.0 \
  --set logging.level=debug
```

### Helm Upgrade

```bash
# Upgrade to latest
helm upgrade deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system

# Upgrade with new values
helm upgrade deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --values new-values.yaml

# Upgrade specific value
helm upgrade deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --set image.tag=v1.1.0
```

### Helm Uninstall

```bash
# Uninstall (keeps CRD by default)
helm uninstall deploykubestack \
  --namespace deploykubestack-system

# Uninstall and remove CRD
helm uninstall deploykubestack \
  --namespace deploykubestack-system
  # Then manually delete CRD if needed:
  # kubectl delete crd applications.apps.deploykubestack.com
```

### Helm Status & Debug

```bash
# Check release status
helm status deploykubestack -n deploykubestack-system

# View values for current release
helm get values deploykubestack -n deploykubestack-system

# View manifest
helm get manifest deploykubestack -n deploykubestack-system

# Get release history
helm history deploykubestack -n deploykubestack-system

# Rollback to previous release
helm rollback deploykubestack 1 -n deploykubestack-system
```

## Monitoring

### Health Checks

```bash
# Port-forward to access health probes
kubectl port-forward -n deploykubestack-system \
  svc/deploykubestack-controller-manager 8081:8081

# Check liveness
curl http://localhost:8081/healthz

# Check readiness
curl http://localhost:8081/readyz
```

### Metrics

```bash
# Port-forward to metrics server
kubectl port-forward -n deploykubestack-system \
  svc/deploykubestack-controller-manager-metrics-service 8080:8080

# Scrape metrics
curl http://localhost:8080/metrics | grep deploykubestack

# Configure Prometheus scrape job:
# - job_name: 'deploykubestack'
#   kubernetes_sd_configs:
#   - role: service
#     namespaces:
#       names:
#       - deploykubestack-system
#   relabel_configs:
#   - source_labels: [__meta_kubernetes_service_port_name]
#     regex: metrics
#     action: keep
```

## Troubleshooting

### Pod Not Running

```bash
# Check pod status
kubectl get pods -n deploykubestack-system

# View events
kubectl describe pod <pod-name> -n deploykubestack-system

# Check logs
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager

# Check resource availability
kubectl describe nodes | grep -A 5 "Allocated resources"
```

### CRD Issues

```bash
# Verify CRD exists
kubectl get crd applications.apps.deploykubestack.com -o yaml

# Check CRD status
kubectl describe crd applications.apps.deploykubestack.com

# If missing, reinstall chart with crds.install=true
helm upgrade deploykubestack ./helm/deploykubestack \
  --set crds.install=true
```

### RBAC Issues

```bash
# Check service account
kubectl get sa deploykubestack-controller-manager -n deploykubestack-system

# Check cluster role
kubectl get clusterrole deploykubestack-controller-manager-role -o yaml

# Check role binding
kubectl get clusterrolebinding deploykubestack-controller-manager-rolebinding -o yaml

# Test RBAC with kubectl auth can-i
kubectl auth can-i create deployments \
  --as=system:serviceaccount:deploykubestack-system:deploykubestack-controller-manager
```

### Applications Not Creating Resources

```bash
# Check operator logs for errors
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager -f

# Check Application status
kubectl describe application my-app

# Verify child resources exist
kubectl get deployment,service -l app=my-app
```

## Advanced Configuration

### Private Registry

```bash
# Create image pull secret
kubectl create secret docker-registry private-registry-secret \
  --docker-server=private.registry.com \
  --docker-username=username \
  --docker-password=password \
  -n deploykubestack-system

# Use in Helm chart
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --set image.registry=private.registry.com \
  --set imagePullSecrets[0].name=private-registry-secret
```

### High Availability

```yaml
# values-ha.yaml
replicaCount: 3

leaderElection:
  enabled: true
  namespace: deploykubestack-system

affinity:
  podAntiAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
    - labelSelector:
        matchExpressions:
        - key: app.kubernetes.io/name
          operator: In
          values:
          - deploykubestack
      topologyKey: kubernetes.io/hostname
```

```bash
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --values values-ha.yaml
```

### Namespace-Scoped Operator

By default, the operator is cluster-scoped (manages Applications in all namespaces). For namespace-scoped operator, you would need to modify the RBAC to use Role/RoleBinding instead of ClusterRole/ClusterRoleBinding (requires chart modifications).

## Chart File Structure

```
helm/deploykubestack/
├── Chart.yaml                         # Chart metadata & version
├── values.yaml                        # Default values
├── README.md                          # Chart README
└── templates/
    ├── _helpers.tpl                   # Template helper functions
    ├── crd-application.yaml            # Application CRD definition
    ├── serviceaccount.yaml             # Service account
    ├── clusterrole.yaml                # RBAC cluster role
    ├── clusterrolebinding.yaml         # RBAC cluster role binding
    ├── deployment.yaml                 # Controller deployment
    └── service-metrics.yaml            # Metrics service
```

## References

- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [CRD Documentation](https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/)
- [DeployKubeStack GitHub](https://github.com/rv2023/DeployKubeStack)

---

**For issues or improvements:** https://github.com/rv2023/DeployKubeStack/issues
