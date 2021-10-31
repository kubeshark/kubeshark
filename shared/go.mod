module github.com/up9inc/mizu/shared

go 1.16

require (
	github.com/docker/go-units v0.4.0
	github.com/golang-jwt/jwt/v4 v4.1.0
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/up9inc/mizu/tap/api v0.0.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/kubectl v0.21.0
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../tap/api
