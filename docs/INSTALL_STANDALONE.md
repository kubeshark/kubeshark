# Mizu install standalone

Mizu can be run detached from the cli using the install command: `mizu install`. This type of mizu instance will run
indefinitely in the cluster.

Please note that install standalone requires you to have RBAC creation permissions, see the [permissions](PERMISSIONS.md)
doc for more details.

```bash
$ mizu install
```

## Stop mizu install

To stop the detached mizu instance and clean all cluster side resources, run `mizu clean`

```bash
$ mizu clean # mizu will continue running in cluster until clean is executed
  Removing mizu resources
```

## Expose mizu web app

Mizu could be exposed at a later stage in any of the following ways:

### Using mizu view command

In a machine that can access both the cluster and a browser, you can run `mizu view` command which creates a proxy.
Besides that, all the regular ways to expose k8s pods are valid.

```bash
$ mizu view
  Establishing connection to k8s cluster...
  Mizu is available at http://localhost:8899
  ^C
  ..
```

### Port Forward

```bash
$ kubectl port-forward -n mizu deployment/mizu-api-server 8899:8899
```

### NodePort

```bash
$ kubectl expose -n mizu deployment mizu-api-server --name mizu-node-port --type NodePort --port 80 --target-port 8899
```

Mizu's IP is the IP of any node (get the IP with `kubectl get nodes -o wide`) and the port is the target port of the new
service (`kubectl get services -n mizu mizu-node-port`). Note that this method will expose Mizu to public access if your
nodes are public.

### LoadBalancer

```bash
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
