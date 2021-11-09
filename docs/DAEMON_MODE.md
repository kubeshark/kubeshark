### To run mizu mizu daemon mode (detached from cli)

Please note that daemon mode requires you to have RBAC creation permissions, see the [permissions](PERMISSIONS.md) doc
for more details.

```bash
$ mizu tap "^ca.*" --daemon
  Mizu will store up to 200MB of traffic, old traffic will be cleared once the limit is reached.
  Tapping pods in namespaces "sock-shop"
  Waiting for mizu to be ready... (may take a few minutes)
  +carts-66c77f5fbb-fq65r
  +catalogue-5f4cb7cf5-7zrmn
  ..

$ mizu view
  Establishing connection to k8s cluster...
  Mizu is available at http://localhost:8899
  ^C
  ..

$ mizu clean # mizu will continue running in cluster until clean is executed
  Removing mizu resources
```

### To run mizu daemon mode with LoadBalancer kubernetes service

```bash
$ mizu tap "^ca.*" --daemon
  Mizu will store up to 200MB of traffic, old traffic will be cleared once the limit is reached.
  Tapping pods in namespaces "sock-shop"
  Waiting for mizu to be ready... (may take a few minutes)
  ..

$ kubectl expose deployment -n mizu --port 80 --target-port 8899 mizu-api-server --type=LoadBalancer --name=mizu-lb
  service/mizu-lb exposed
  ..
  
$ kubectl get services -n mizu
  NAME              TYPE           CLUSTER-IP       EXTERNAL-IP     PORT(S)        AGE
  mizu-api-server   ClusterIP      10.107.200.100   <none>          80/TCP         5m5s
  mizu-lb           LoadBalancer   10.107.200.101   34.77.120.116   80:30141/TCP   76s
```

Note that `LoadBalancer` services only work on supported clusters (usually cloud providers) and might incur extra costs

If you changed the `mizu-resources-namespace` value, make sure the `-n mizu` flag of the `kubectl expose` command is
changed to the value of `mizu-resources-namespace`

mizu will now be available both by running `mizu view` or by accessing the `EXTERNAL-IP` of the `mizu-lb` service
through your browser.
