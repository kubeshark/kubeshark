module github.com/up9inc/mizu/tap

go 1.16

require (
	github.com/google/gopacket v1.1.19
	github.com/google/martian v2.1.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/orcaman/concurrent-map v0.0.0-20210501183033-44dafcb38ecc
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/up9inc/mizu/shared v0.0.0
	golang.org/x/net v0.0.0-20210525063256-abc453219eb5
)

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared
