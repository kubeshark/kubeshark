module github.com/up9inc/mizu/tap/extensions/http

go 1.16

require (
	github.com/beevik/etree v1.1.0 // indirect
	github.com/google/martian v2.1.0+incompatible
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
	golang.org/x/text v0.3.5 // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
