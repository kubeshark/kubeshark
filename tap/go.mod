module github.com/up9inc/mizu/tap

go 1.17

require (
	github.com/cilium/ebpf v0.8.0
	github.com/go-errors/errors v1.4.2
	github.com/google/gopacket v1.1.19
	github.com/hashicorp/golang-lru v0.5.4
	github.com/shirou/gopsutil v3.21.11+incompatible
	github.com/struCoder/pidusage v0.2.1
	github.com/up9inc/mizu/logger v0.0.0
	github.com/up9inc/mizu/tap/api v0.0.0
	github.com/up9inc/mizu/tap/dbgctl v0.0.0
	github.com/vishvananda/netns v0.0.0-20211101163701-50045581ed74
	k8s.io/api v0.23.3
)

require (
	github.com/go-logr/logr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20220207234003-57398862261d // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.23.3 // indirect
	k8s.io/klog/v2 v2.40.1 // indirect
	k8s.io/utils v0.0.0-20220127004650-9b3446523e65 // indirect
	sigs.k8s.io/json v0.0.0-20211208200746-9f7c6b3444d2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
	sigs.k8s.io/yaml v1.3.0 // indirect
)

replace github.com/up9inc/mizu/logger v0.0.0 => ../logger

replace github.com/up9inc/mizu/tap/api v0.0.0 => ./api

replace github.com/up9inc/mizu/tap/dbgctl v0.0.0 => ./dbgctl
