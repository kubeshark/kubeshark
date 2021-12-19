![Mizu: The API Traffic Viewer for Kubernetes](../assets/mizu-logo.svg)
# Configuration options for Mizu

Mizu has many configuration options and flags that affect its behavior. Their values can be modified via command-line interface or via configuration file. 

The list below covers most useful configuration options.

### Config file
Mizu behaviour can be modified via YAML configuration file located at `$HOME/.mizu/config.yaml`. 

Default values for the file can be viewed via `mizu config` command.

### Applying config options via command line
To apply any configuration option via command line, use `--set` following by config option name and value, like in the following example:

```
mizu tap --set tap.dry-run=true
```

Please make sure to use full option name (`tap.dry-run` as opposed to `dry-run` only), incl. section (`tap`, in this example)

## General section

* `agent-image` - full path to Mizu container image, in format `full.path.to/your/image:tag`. Default value is set at compilation time to `gcr.io/up9-docker-hub/mizu/<branch>:<version>`

* `dump-logs` - if set to `true`, saves log files for all Mizu components (tapper, api-server, CLI) in a zip file under `$HOME/.mizu`. Default value is `false`

* `image-pull-policy` - container image pull policy for Kubernetes, default value `Always`. Other accepted values are `Never` or `IfNotExist`. Please mind the implications when changing this.

* `kube-config-path` - path to alternative kubeconfig file to use for all interactions with Kubernetes cluster. By default - `$HOME/.kubeconfig`

* `mizu-resources-namespace` - Kubernetes namespace where all Mizu-related resources are created. Default value `mizu`

* `telemetry` - report anonymous usage statistics. Default value `true`

## section `tap`
* `namespaces` - list of namespace names, in which pods are tapped. Default value is empty, meaning only pods in the current namespace are tapped. Typically supplied as command line options.

* `all-namespaces` - special flag indicating whether Mizu should search and tap pods, matching the regex, in all namespaces. Default is `false`. Please use with caution, tapping too many pods can affect resource consumption.

* `daemon` - instructs Mizu whether to run daemon mode (where CLI command exits after launch, and tapper & api-server pods in Kubernetes continue to run without controlling CLI). Typically supplied as command-line option `--daemon`. Default valie is `false`

* `dry-run` - if true, Mizu will print list of pods matching the supplied (or default) regex and exit without actually tapping the traffic. Default value is `false`. Typically supplied as command-line option `--dry-run`

* `proxy-host` - IP address on which proxy to Mizu API service is launched; should be accessible at `proxy-host:gui-port`. Default value is `127.0.0.1`

* `gui-port` - port on which Mizu GUI is accessible, default value is `8899` (stands for `8899/tcp`)

* `regex` - regular expression used to match pods to tap, when no regex is given in the command line; default value is `.*`, which means `mizu tap` with no additional arguments is runnining as `mizu tap .*` (i.e. tap all pods found in current workspace)

* `no-redact` - instructs Mizu whether to redact certain sensitive fields in the collected traffic. Default value is `false`, i.e. Mizu will replace sentitive data values with *REDACTED* placeholder.

* `ignored-user-agents` - array of strings, describing HTTP *User-Agent* header values to be ignored. Useful to ignore Kubernetes healthcheck and other similar noisy periodic probes. Default value is empty.

* `max-entries-db-size` - maximal size of traffic stored locally in the `mizu-api-server` pod. When this size is reached, older traffic is overwritten with new entries. Default value is `200MB`
 

### section `tap.api-server-resources`
Kubernetes request and limit values for the `mizu-api-server` pod.
Parameters and their default values are same as used natively in Kubernetes pods:

```
        cpu-limit: 750m
        memory-limit: 1Gi
        cpu-requests: 50m
        memory-requests: 50Mi
```

### section `tap.tapper-resources`
Kubernetes request and limit values for the `mizu-tapper` pods (launched via daemonset).
Parameters and their default values are same as used natively in Kubernetes pods:

```
        cpu-limit: 750m
        memory-limit: 1Gi
        cpu-requests: 50m
        memory-requests: 50Mi
```


--

* `analsys` - enables advanced analysis of collected traffic in the UP9 coud platform. Default value is `false`

* `upload-interval` -  in the *analysis* mode, push traffic to UP9 cloud every `upload-interval` seconds. Default value is `10` seconds

* `ask-upload-confirmation` - request user confirmation when uploading tapped data to UP9 cloud


## section `version`
* `debug`- print additional version and build information when `mizu version` command is invoked. Default value is `false`.
