module github.com/up9inc/mizu/tap/extensions/http

go 1.16

require (
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7 // indirect
	github.com/up9inc/mizu/tap/api v0.0.0
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
