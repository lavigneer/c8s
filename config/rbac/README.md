# RBAC Configuration for C8S Controller

This directory contains Role-Based Access Control (RBAC) manifests for the C8S controller.

## Files

### service_account.yaml
Creates the ServiceAccount that the controller runs as.

### role.yaml
Defines the ClusterRole with all permissions the controller needs:

**C8S CRD Permissions**:
- Full CRUD on PipelineConfigs, PipelineRuns, RepositoryConnections
- Status subresource updates
- Finalizer management

**Kubernetes Resource Permissions**:
- **Jobs**: Full CRUD (controller creates Jobs for each step)
- **Pods**: Read-only (monitor job execution)
- **Secrets**: Read-only (inject into step containers)
- **ConfigMaps**: CRUD (store pipeline metadata)
- **Events**: Create/patch (record pipeline events)
- **ResourceQuotas**: Read-only (enforce namespace limits)
- **Leases**: Full CRUD (leader election for HA)

### role_binding.yaml
Binds the ClusterRole to the ServiceAccount.

## Installation

```bash
# Install RBAC resources
kubectl apply -f config/rbac/

# Verify ServiceAccount
kubectl get serviceaccount c8s-controller -n c8s-system

# Verify ClusterRole
kubectl get clusterrole c8s-controller-role

# Verify ClusterRoleBinding
kubectl get clusterrolebinding c8s-controller-rolebinding
```

## Security Considerations

- **Least Privilege**: Controller only has permissions it needs
- **ClusterRole**: Required because controller watches resources across all namespaces
- **Secret Access**: Read-only, secrets are not persisted by controller
- **Job Creation**: Controller creates Jobs in the same namespace as PipelineRun
- **Namespace Isolation**: Teams can use separate namespaces with ResourceQuotas

## Multi-Tenancy

For multi-tenant deployments, consider:

1. **Namespace-scoped deployment**: Use Role instead of ClusterRole
   ```bash
   # Replace ClusterRole with Role
   # Replace ClusterRoleBinding with RoleBinding
   ```

2. **Per-namespace controller**: Deploy controller in each team namespace

3. **ResourceQuotas**: Limit resources per namespace
   ```yaml
   apiVersion: v1
   kind: ResourceQuota
   metadata:
     name: team-quota
     namespace: team-a
   spec:
     hard:
       requests.cpu: "100"
       requests.memory: 200Gi
       pods: "50"
   ```

## Troubleshooting

### Controller can't create Jobs
Check RBAC permissions:
```bash
kubectl auth can-i create jobs --as=system:serviceaccount:c8s-system:c8s-controller -n default
```

### Controller can't read Secrets
Check RBAC permissions:
```bash
kubectl auth can-i get secrets --as=system:serviceaccount:c8s-system:c8s-controller -n default
```

### Leader election failures
Check lease permissions:
```bash
kubectl auth can-i create leases --as=system:serviceaccount:c8s-system:c8s-controller -n c8s-system
```
