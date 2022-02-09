module github.com/up9inc/mizu/tap/extensions/http

go 1.17

require (
	github.com/beevik/etree v1.1.0
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
)

require (
	github.com/google/martian v2.1.0+incompatible // indirect
	golang.org/x/text v0.3.7 // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
