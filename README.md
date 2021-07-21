# 水 mizu
A simple-yet-powerful API traffic viewer for Kubernetes to help you troubleshoot and debug your microservices. Think TCPDump and Chrome Dev Tools combined.

## Download

Download `mizu` for your platform and operating system

### Latest stable release

* for MacOS - Intel 
```
curl -Lo mizu \
https://github.com/up9inc/mizu/releases/latest/download/mizu_darwin_amd64 \
&& chmod 755 mizu
```
 
* for Linux - Intel 64bit
```
curl -Lo mizu \
https://github.com/up9inc/mizu/releases/latest/download/mizu_linux_amd64 \
&& chmod 755 mizu
``` 

SHA256 checksums are available on the [Releases](https://github.com/up9inc/mizu/releases) page.

### Development (unstable) build
Pick one from the [Releases](https://github.com/up9inc/mizu/releases) page.

## Prerequisites
1. Set `KUBECONFIG` environment variable to your kubernetes configuration. If this is not set, mizu assumes that configuration is at `${HOME}/.kube/config`
2. mizu needs following permissions on your kubernetes cluster to run

```yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - list
  - watch
  - create
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - create
- apiGroups:
  - apps
  resources:
  - daemonsets
  verbs:
  - create
  - patch
- apiGroups:
  - ""
  resources:
  - namespaces
  verbs:
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
3. Optionally, for resolving traffic ip to kubernetes service name, mizu needs below permissions

```yaml
- apiGroups:
  - ""
  resources:
  - pods
  verbs:
  - get
- apiGroups:
  - ""
  resources:
  - services
  verbs:
  - get
  - list
  - watch
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
- apiGroups:
  - ""
  resources:
  - serviceaccounts
  verbs:
  - get
  - create
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterroles
  verbs:
  - list
  - create
  - delete
- apiGroups:
  - rbac.authorization.k8s.io
  resources:
  - clusterrolebindings
  verbs:
  - list
  - create
  - delete
```

## How to run

1. Find pod you'd like to tap to in your Kubernetes cluster
2. Run `mizu tap PODNAME` or `mizu tap REGEX` 
3. Open browser on `http://localhost:8899` as instructed .. 
4. Watch the WebAPI traffic flowing ..
5. Type ^C to stop

## Examples

Run `mizu help` for usage options


To tap specific pod - 
``` 
 $ kubectl get pods 
 NAME                            READY   STATUS    RESTARTS   AGE
 front-end-649fc5fd6-kqbtn       2/2     Running   0          7m
 ..

 $ mizu tap front-end-649fc5fd6-kqbtn
 +front-end-649fc5fd6-kqbtn
 Web interface is now available at http://localhost:8899
 ^C
```

To tap multiple pods using regex - 
``` 
 $ kubectl get pods 
 NAME                            READY   STATUS    RESTARTS   AGE
 carts-66c77f5fbb-fq65r          2/2     Running   0          20m
 catalogue-5f4cb7cf5-7zrmn       2/2     Running   0          20m
 front-end-649fc5fd6-kqbtn       2/2     Running   0          20m
 ..

 $ mizu tap "^ca.*"
 +carts-66c77f5fbb-fq65r
 +catalogue-5f4cb7cf5-7zrmn
 Web interface is now available at http://localhost:8899
 ^C
```

