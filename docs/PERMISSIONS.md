![Mizu: The API Traffic Viewer for Kubernetes](../assets/mizu-logo.svg)

# Kubernetes permissions for MIZU  

This document describes in details all permissions required for full and correct operation of Mizu.

## Editting permissions

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

We broke down this list into few categories:

- Required - what is needed for `mizu` to run properly on your k8s cluster
- Optional - permissions needed for proper name resolving for service & pod IPs
  - addition required for policy validation

### Required permissions

Mizu needs following permissions on your Kubernetes cluster to run properly

```yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - create
  - delete
- apiGroups:
  - apps
  resources:
  - daemonsets
  verbs:
  - create
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - services/proxy
  verbs:
  - get
```

#### Permissions required running with install command or (optional) for service / pod name resolving

Mandatory permissions for running with install command.

Optional for service/pod name resolving in non install standalone

```yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - apps
  resources:
  - daemonsets
  verbs:
  - create
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - services/proxy
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterroles
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterrolebindings
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - roles
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - rolebindings
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - apps
  - extensions
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  - extensions
  resources:
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  - apps
  - extensions
  resources:
  - endpoints
  verbs:
  - get
  - list
  - watch
```

#### Permissions for Policy rules validation feature (opt)

Optionally, in order to use the policy rules validation feature, Mizu requires the following additional permissions:

```yaml
- apiGroups:
  - ""
  resources:
  - configmaps
  verbs:
  - get
  - create
  - delete
```

- - -

#### Namespace-Restricted mode

Alternatively, in order to restrict Mizu to one namespace only (by setting `agent.namespace` in the config file), Mizu needs the following permissions in that namespace:

```yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - apps
  resources:
  - daemonsets
  verbs:
  - get
  - create
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - services/proxy
  verbs:
  - get
```

##### Name resolving in Namespace-Restricted mode (opt)

To restrict Mizu to one namespace while also resolving IPs, Mizu needs the following permissions in that namespace:

```yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
  - create
  - delete
- apiGroups:
  - apps
  resources:
  - daemonsets
  verbs:
  - get
  - create
  - patch
  - delete
- apiGroups:
  - ""
  resources:
  - services/proxy
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - roles
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - rolebindings
  verbs:
  - get
  - create
  - delete
- apiGroups:
  - apps
  - extensions
  resources:
  - pods
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - apps
  - extensions
  resources:
  - services
  verbs:
  - get
  - list
  - watch
- apiGroups:
  - ""
  - apps
  - extensions
  resources:
  - endpoints
  verbs:
  - get
  - list
  - watch
```
