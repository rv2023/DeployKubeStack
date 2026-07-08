# Helm Values Examples - DeployKubeStack Operator

This guide provides pre-configured values files for different deployment scenarios: development, staging, and production.

## Overview

Three environment-specific values files are provided in the Helm chart:

- **`values-dev.yaml`** - Development environment (minimal resources, debug logging)
- **`values-staging.yaml`** - Staging environment (moderate resources, info logging)
- **`values-prod.yaml`** - Production environment (HA setup, hardened security, warn logging)

Each can be used to deploy the operator with appropriate configurations for its environment.

## Development Deployment

**Use Case:** Local development, testing, CI/CD test environments

**Characteristics:**
- Single replica (no HA required)
- Minimal CPU (50m) and memory (32Mi) requests
- Debug logging enabled for detailed troubleshooting
- Leader election disabled
- No pod annotations

### Installation

```bash
# Using local development values
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values ./helm/deploykubestack/values-dev.yaml

# Or explicitly
helm install deploykubestack ./helm/deploykubestack \
  -f values-dev.yaml \
  -n deploykubestack-system \
  --create-namespace
```

### Verification

```bash
# Check deployment
kubectl get deployment -n deploykubestack-system

# View debug logs
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager -f

# Check pod resource usage (should be low)
kubectl top pod -n deploykubestack-system
```

### Resource Usage

```
CPU Requests:     50m    (very light)
CPU Limits:       200m   (light)
Memory Requests:  32Mi   (minimal)
Memory Limits:    128Mi  (minimal)
```

### Customization for Development

```bash
# Enable webhook in development if needed
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values values-dev.yaml \
  --set webhook.enabled=true

# Increase verbosity further
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values values-dev.yaml \
  --set logging.level=debug \
  --set logging.development=true
```

## Staging Deployment

**Use Case:** Pre-production testing, QA environment, staging workloads

**Characteristics:**
- Single replica (cost-effective)
- Moderate resources (100m CPU, 64Mi memory)
- Info-level logging (balanced verbosity)
- Preferred pod anti-affinity (spread across nodes if possible)
- Linux node selector

### Installation

```bash
# Using staging values
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values ./helm/deploykubestack/values-staging.yaml

# Or with interactive confirmation
helm install deploykubestack ./helm/deploykubestack \
  -f values-staging.yaml \
  -n deploykubestack-system \
  --create-namespace \
  --dry-run --debug  # Preview before installing
```

### Verification

```bash
# Check deployment status
kubectl get deployment -n deploykubestack-system -o wide

# View application logs
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager -f

# Check pod node placement
kubectl get pods -n deploykubestack-system -o wide

# Verify health probes
kubectl describe pod -n deploykubestack-system \
  -l app.kubernetes.io/name=deploykubestack
```

### Resource Usage

```
CPU Requests:     100m   (moderate)
CPU Limits:       500m   (moderate)
Memory Requests:  64Mi   (modest)
Memory Limits:    256Mi  (moderate)
```

### High Availability in Staging

To test HA capabilities in staging:

```bash
# Scale to 3 replicas with leader election
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values values-staging.yaml \
  --set replicaCount=3 \
  --set leaderElection.enabled=true
```

## Production Deployment

**Use Case:** Production environments, mission-critical applications

**Characteristics:**
- 3 replicas for high availability
- Production-grade resources (200m CPU, 128Mi memory requests)
- Warn-level logging (reduced overhead)
- Required pod anti-affinity (no two replicas on same node)
- Leader election enabled
- Hardened security context
- Prometheus annotations

### Installation

```bash
# Using production values
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values ./helm/deploykubestack/values-prod.yaml

# Preview before install
helm install deploykubestack ./helm/deploykubestack \
  -f values-prod.yaml \
  -n deploykubestack-system \
  --create-namespace \
  --dry-run=server

# Verify all components will be created
helm template deploykubestack ./helm/deploykubestack \
  -f values-prod.yaml \
  -n deploykubestack-system | kubectl apply --dry-run=client -f -
```

### Verification

```bash
# Check all 3 replicas are running
kubectl get deployment -n deploykubestack-system
kubectl get pods -n deploykubestack-system -o wide

# Verify leader election
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager | grep -i leader

# Check resource allocation
kubectl top deployment -n deploykubestack-system
kubectl top pods -n deploykubestack-system

# Verify anti-affinity is working
kubectl get pods -n deploykubestack-system \
  -o custom-columns=NAME:.metadata.name,NODE:.spec.nodeName

# Ensure pods are on different nodes
# Output should show 3 different nodes
```

### Resource Usage

```
CPU Requests:     200m per replica = 600m total (modest HA)
CPU Limits:       1000m per replica = 3000m total (scalable)
Memory Requests:  128Mi per replica = 384Mi total
Memory Limits:    512Mi per replica = 1536Mi total
```

### Production Best Practices

**1. Verify Pod Disruption Budget**

```bash
# Check if PDB should be enabled for your cluster
kubectl get nodes

# If you have 3+ nodes, enable PDB to prevent simultaneous evictions
helm upgrade deploykubestack ./helm/deploykubestack \
  --values values-prod.yaml \
  --set podDisruptionBudget.minAvailable=2 \
  -n deploykubestack-system
```

**2. Configure Monitoring**

```bash
# Prometheus ServiceMonitor (if monitoring is installed)
# The operator exposes metrics on port 8080
# Pod annotations already configured for Prometheus scraping

kubectl port-forward -n deploykubestack-system \
  svc/deploykubestack-controller-manager-metrics-service 8080:8080

curl http://localhost:8080/metrics | grep deploykubestack_
```

**3. Configure Alerting**

Create PrometheusRule for production:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: deploykubestack-alerts
  namespace: deploykubestack-system
spec:
  groups:
  - name: deploykubestack
    interval: 30s
    rules:
    - alert: DeployKubeStackDown
      expr: up{job="deploykubestack"} == 0
      for: 5m
      labels:
        severity: critical
      annotations:
        summary: "DeployKubeStack operator is down"
    - alert: DeployKubeStackHighErrorRate
      expr: rate(controller_runtime_reconcile_errors_total[5m]) > 0.1
      for: 10m
      labels:
        severity: warning
      annotations:
        summary: "DeployKubeStack operator has high error rate"
```

**4. Setup Log Aggregation**

```bash
# Example with ELK Stack (Elasticsearch, Logstash, Kibana)
# Logs are output to stderr and can be collected by log aggregators

# View logs in real-time
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager -f

# Stream logs to a file
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager --tail=-1 \
  > operator-logs.txt

# Use kubetail for multi-pod tailing
kubetail deploykubestack -n deploykubestack-system
```

**5. Backup CRD Definition**

```bash
# Backup the CRD for disaster recovery
kubectl get crd applications.apps.deploykubestack.com -o yaml \
  > applications-crd-backup.yaml

# Backup the Application CR list
kubectl get applications -A -o yaml > applications-backup.yaml

# Store these backups in version control or backup system
```

**6. Upgrade Strategy**

```bash
# In-place rolling upgrade (default)
helm upgrade deploykubestack ./helm/deploykubestack \
  --values values-prod.yaml \
  -n deploykubestack-system
# Kubernetes automatically rolls out updates with leader election

# Blue-green deployment (manual)
# 1. Install new version with different release name
# 2. Switch traffic when ready
# 3. Delete old release

helm install deploykubestack-v2 ./helm/deploykubestack \
  --values values-prod.yaml \
  -n deploykubestack-system \
  --set version=v2

# After verification
helm delete deploykubestack -n deploykubestack-system
```

## Switching Between Environments

### Update from Development to Staging

```bash
# When code is ready for staging
helm upgrade deploykubestack ./helm/deploykubestack \
  --values ./helm/deploykubestack/values-staging.yaml \
  -n deploykubestack-system
```

### Update from Staging to Production

```bash
# After staging validation
helm upgrade deploykubestack ./helm/deploykubestack \
  --values ./helm/deploykubestack/values-prod.yaml \
  -n deploykubestack-system

# Verify rollout
kubectl rollout status deployment/deploykubestack-controller-manager \
  -n deploykubestack-system
```

## Customizing Values

### Common Customizations

**Change Log Level Only**
```bash
helm install deploykubestack ./helm/deploykubestack \
  --values values-prod.yaml \
  --set logging.level=debug
```

**Use Private Registry**
```bash
helm install deploykubestack ./helm/deploykubestack \
  --values values-prod.yaml \
  --set image.registry=private.registry.com \
  --set image.repository=internal/deploykubestack \
  --set imagePullSecrets[0].name=docker-secret
```

**Configure Custom Environment Variables**
```bash
helm install deploykubestack ./helm/deploykubestack \
  --values values-prod.yaml \
  --set env[0].name=CUSTOM_VAR \
  --set env[0].value=custom-value
```

**Create Custom Values File**

Combine multiple files for complex configurations:

```yaml
# values-prod-custom.yaml
# Include from values-prod.yaml
replicaCount: 5  # Override for high-traffic production
resources:
  requests:
    cpu: 500m
    memory: 256Mi
  limits:
    cpu: 2000m
    memory: 1Gi

nodeSelector:
  kubernetes.io/os: linux
  node-type: system
  
tolerations:
- key: dedicated
  operator: Equal
  value: system
  effect: NoSchedule
- key: high-traffic
  operator: Equal
  value: "true"
  effect: NoExecute
```

```bash
helm install deploykubestack ./helm/deploykubestack \
  --namespace deploykubestack-system \
  --create-namespace \
  --values values-prod-custom.yaml
```

## Troubleshooting

### Pods not starting

```bash
# Check events
kubectl describe pod -n deploykubestack-system \
  -l app.kubernetes.io/name=deploykubestack

# Check affinity constraints
kubectl get pods -n deploykubestack-system -o wide
# If pods are pending, check node availability

# Check resource availability
kubectl describe nodes | grep -A 5 "Allocated resources"
```

### Leader election issues

```bash
# View leader election logs
kubectl logs -n deploykubestack-system \
  deployment/deploykubestack-controller-manager | grep -i leader

# Check configmap for leader info
kubectl get configmap -n deploykubestack-system
kubectl describe configmap deploykubestack-leader-election \
  -n deploykubestack-system
```

### High resource usage

```bash
# If actual usage exceeds limits
helm upgrade deploykubestack ./helm/deploykubestack \
  -f values-prod.yaml \
  --set resources.limits.cpu=2000m \
  --set resources.limits.memory=1Gi \
  -n deploykubestack-system
```

## References

- [Helm Values Best Practices](https://helm.sh/docs/chart_best_practices/)
- [Kubernetes Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [Pod Affinity Rules](https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/#affinity-and-anti-affinity)
- [Leader Election in Controllers](https://kubernetes.io/docs/concepts/architecture/leases/)

---

**Last Updated:** July 7, 2026
