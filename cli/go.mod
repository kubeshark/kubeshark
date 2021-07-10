module github.com/up9inc/mizu/cli

go 1.16

require (
	github.com/gorilla/websocket v1.4.2
	github.com/spf13/cobra v1.1.3
	github.com/up9inc/mizu/shared v0.0.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/kubectl v0.21.0
)

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared
