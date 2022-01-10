![Mizu: The API Traffic Viewer for Kubernetes](../assets/mizu-logo.svg)

# Kubernetes permissions for MIZU  

This document describes in details all permissions required for full and correct operation of Mizu.

We broke down this list into few categories:

- Required - what is needed for `mizu` to run properly on your k8s cluster
- Optional - permissions needed for proper name resolving for service & pod IPs
  - addition required for policy validation

# Required permissions

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

## Permissions required running with install command or (optional) for service / pod name resolving

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

## Permissions for Policy rules validation feature (opt)

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

## Namespace-Restricted mode

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

### Name resolving in Namespace-Restricted mode (opt)

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
