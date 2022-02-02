# CHANGELOG
This document summarizes main and fixes changes published in stable (aka `main`) branch of this project.
Ongoing work and development releases are under `develop` branch.

## 0.24.0

### main features
* ARM64 support -- Mizu is now available for ARM 64bit architecture
  * Now you can run Mizu with `minikube` on your Apple M1 laptop or any other ARM-based hosts
* New command helps user verify Mizu deployment
  * Run `mizu check` to verify Mizu was deployed successfully
  * `mizu check` verifies version compatibility, resources and permissions required by Mizu
* EXPERIMENTAL: Service Map - graph of all service interactions
  * Arrow direction show client to server connection
  * Graph edge width reflects volume of traffic captured between the services
  * to enable this experimental feature use `--set service-map=true` flag

### improvements
* Mizu container images are now served from [Docker Hub](https://hub.docker.com/r/up9inc/mizu), as multi-architecture images (arm64, amd64)
* in Mizu GUI the filter query can now be applied by pressing CONTROL/COMMAND + ENTER
* try port-forwarding if http-proxy connection to Mizu API server is not available

### notable bug fixes
* Fixed HTTP/1.0 presentation which was shown as HTTP/1.1
* Fixed handling of long-living TCP connections, improves capturing gRPC and HTTP/2 traffic, and helps in service-mesh setups (istio, linkerd)


## 0.23.0
### notable bug fixes
* fixed errors in Redis protocol parser (better handling of Array and Bulk String message types)



## 0.22.0

### main features
* Service Mesh support -- mizu is now capable to tap mTLS traffic between pods connected by Istio service mesh
  * Use `--service-mesh` option to enable this feature
* New installation option - have the same Mizu functionality as long living pods in your cluster, with password protection
  * To install use `mizu install` command
  * To access use `mizu view` or `kubectl -n mizu port-forward svc/mizu-api-server`
  * To uninstall run `mizu clean`
* At first login
  * Set admin password as prompted, use it to login to mizu later on.
  * After login, user should select cluster namespaces to tap: by default all namespaces in the cluster are selected, user can select/unselect according to their needs. These settings are retained and can be modified at any time via Settings menu (cog icon on the top-right)


### improvements
* improved Mizu permissions/roles logic to support clusters with strict PodSecurityPolicy (PSP) -- see [PERMISSIONS](PERMISSIONS.md) doc for more details
 
### notable bug fixes
* mizu now works properly when API service is exposed via HTTPS url
* mizu now properly displays KAFKA message body 




## 0.21.0

### main features
* New traffic search & stream exprience
* Rich query language with full-text search capabilities on headers & body
* Distinct live-streaming vs paging/browsing modes, all with filter applied

### improvements
* GUI - source and destination IP addresses & service names for each traffic item
* GUI - Mizu health - display warning sign in top bar when not all requested pods are successfully tapped
* GUI - pod tapping status reflected in the list (ok or problem)
* Mizu telemetry - report platform type

### fixes
* Request duration and body size properly shown in GUI (instead of -1)
