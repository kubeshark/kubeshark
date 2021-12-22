Release Notes 

V0.21.0
Main features
* New Traffic search & Stream experience 
  * Rich query language with full-text search capabilities on headers & body
  * Distinct live-streaming vs paging/browsing modes, all with the filter applied

* Display Improvements
  Display source and destination IP address & service names for traffic item

* Mizu Health
  * Display warning when not all requested pods are up and running properly 
  * Pod tapping status is reflected in the pod's list in the top bar

* Mizue Telemetry 
  Report platform type

Notable bug fixes
* mizu tap daemonset prints duplicate pods and errors
* mizu: HTTP2 upgraded requests not shown
* Mizu tappers wasn't deployed because of security issue - add informative message to user
* mizu - shows several duplicates http entries
