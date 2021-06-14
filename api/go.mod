module mizuserver

go 1.16

require (
	github.com/antoniodipinto/ikisocket v0.0.0-20210417133349-f1502512d69a
	github.com/beevik/etree v1.1.0
	github.com/djherbis/atime v1.0.0
	github.com/fasthttp/websocket v1.4.3-beta.1 // indirect
	github.com/go-playground/locales v0.13.0
	github.com/go-playground/universal-translator v0.17.0
	github.com/go-playground/validator/v10 v10.5.0
	github.com/gofiber/fiber/v2 v2.8.0
	github.com/google/gopacket v1.1.19
	github.com/google/martian v2.1.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20210106121528-16402b402231
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/up9inc/mizu/shared v0.0.0
	github.com/up9inc/mizu/tap v0.0.0
	go.mongodb.org/mongo-driver v1.5.1
	golang.org/x/net v0.0.0-20210421230115-4e50805a0758
	gorm.io/driver/sqlite v1.1.4
	gorm.io/gorm v1.21.8
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
)

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared
replace github.com/up9inc/mizu/tap v0.0.0 => ../tap
