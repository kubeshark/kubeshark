module github.com/up9inc/mizu/tap

go 1.17

require (
	github.com/bradleyfalzon/tlsx v0.0.0-20170624122154-28fd0e59bac4
	github.com/google/gopacket v1.1.19
	github.com/up9inc/mizu/shared v0.0.0
	github.com/up9inc/mizu/tap/api v0.0.0
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f
	k8s.io/api v0.21.2
)

require (
	github.com/go-logr/logr v0.4.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.5.4 // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/sys v0.0.0-20210426230700-d19ff857e887 // indirect
	golang.org/x/text v0.3.4 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/apimachinery v0.21.2 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.1.0 // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ./api

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared
