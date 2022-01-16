# CHANGELOG
This document summarizes main and fixes changes published in stable (aka `main`) branch of this project.
Ongoing work and development releases are under `develop` branch.

## 0.22.0

### main features
* Service Mesh support -- mizu is now capable to tap mTLS traffic between pods connected by Istio service mesh
  * Use `--service-mesh` option to enable this feature
* New installation option - have the same Mizu functionality as long living pods in your cluster, with password protection.
  * To install use `mizu install` command
  * To access use `mizu view` or `kubectl -n mizu port-forward svc/mizu-api-server`
  * To uninstall run `mizu clean`
* At first login
  * set admin password as prompted, use it to login to mizu later on.
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
