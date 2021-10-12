module mizuserver

go 1.16

require (
	github.com/djherbis/atime v1.0.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/getkin/kin-openapi v0.76.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gin-contrib/static v0.0.1
	github.com/gin-gonic/gin v1.7.2
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.5.0
	github.com/google/martian v2.1.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/orcaman/concurrent-map v0.0.0-20210106121528-16402b402231
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7
	github.com/up9inc/mizu/shared v0.0.0
	github.com/up9inc/mizu/tap v0.0.0
	github.com/up9inc/mizu/tap/api v0.0.0
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0
	go.mongodb.org/mongo-driver v1.7.1
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.8
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
)

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared

replace github.com/up9inc/mizu/tap v0.0.0 => ../tap

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../tap/api
