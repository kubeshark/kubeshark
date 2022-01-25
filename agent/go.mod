module mizuserver

go 1.16

require (
	github.com/antelman107/net-wait-go v0.0.0-20210623112055-cf684aebda7b
	github.com/chanced/openapi v0.0.6
	github.com/djherbis/atime v1.0.0
	github.com/getkin/kin-openapi v0.76.0
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.7
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.5.0
	github.com/google/martian v2.1.0+incompatible
	github.com/google/uuid v1.1.2
	github.com/gorilla/websocket v1.4.2
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/orcaman/concurrent-map v0.0.0-20210106121528-16402b402231
	github.com/ory/keto-client-go v0.7.0-alpha.1
	github.com/ory/kratos-client-go v0.8.2-alpha.1
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/stretchr/testify v1.7.0
	github.com/up9inc/basenine/client/go v0.0.0-20220110083745-04fbc6c2068d
	github.com/up9inc/mizu/shared v0.0.0
	github.com/up9inc/mizu/tap v0.0.0
	github.com/up9inc/mizu/tap/api v0.0.0
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
)

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared

replace github.com/up9inc/mizu/tap v0.0.0 => ../tap

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../tap/api
