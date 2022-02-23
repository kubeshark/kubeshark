![Mizu: The API Traffic Viewer for Kubernetes](../assets/mizu-logo.svg)

# Kubernetes permissions for MIZU  

This document describes in details all permissions required for full and correct operation of Mizu.

## Editing permissions

During installation, Mizu creates a `ServiceAccount` and the roles it requires. No further action is required.
However, if there is a need, it is possible to make changes to Mizu permissions.

### Adding permissions on top of Mizu's defaults

Mizu pods use the `ServiceAccount` `mizu-service-account`. Permissions can be added to Mizu by creating `ClusterRoleBindings` and `RoleBindings` that target that `ServiceAccount`.

For example, in order to add a `PodSecurityPolicy` which allows Mizu to run `hostNetwork` and `privileged` pods, create the following resources:

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: my-mizu-psp
spec:
  hostNetwork: true
  privileged: true
  allowedCapabilities:
    - "*"
  fsGroup:
    rule: RunAsAny
  runAsUser:
    rule: RunAsAny
  seLinux:
    rule: RunAsAny
  supplementalGroups:
    rule: RunAsAny
  volumes:
  - "*"
---
kind: ClusterRole
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: my-mizu-clusterrole
rules:
- apiGroups:
  - policy
  resources:
  - podsecuritypolicies
  verbs:
  - use
  resourceNames:
  - my-mizu-psp
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: my-mizu-clusterrolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: my-mizu-clusterrole
subjects:
- kind: ServiceAccount
  name: mizu-service-account # The service account used by Mizu
  namespace: mizu
```

With this setup, when Mizu starts and creates `mizu-service-account`, this account will be subject to `my-mizu-psp` via `my-mizu-clusterrolebinding`.
When Mizu cleans up resources, the above resources will remain available for future executions.

### Replacing Mizu's default permissions with custom permissions

Mizu does not create its `ServiceAccounts`, `ClusterRoles`, `ClusterRoleBindings`, `Roles` or `RoleBindings` if resources by the same name already exist. In order to replace Mizu's defaults, simply create your resources before running Mizu.

For example, creating a `ClusterRole` by the name of `mizu-cluster-role` before running Mizu will cause Mizu to use that `ClusterRole` instead of the default one created by Mizu.

Notes:

1. The resource names must match Mizu's default names.
2. User-managed resources must not have the label `app.kubernetes.io/managed-by=mizu`. Remove the label or set it to another value.

## List of permissions

The permissions that are required to run Mizu depend on the configuration.
By default Mizu requires cluster-wide permissions.
If these are not available to the user, it is possible to run Mizu in namespace-restricted mode which has a reduced set of requirements.
This is done by by setting the `mizu-resources-namespace` config option. See [configuration](CONFIGURATION.md) for instructions.

The different requirements are listed in [the example roles dir](../examples/roles)
