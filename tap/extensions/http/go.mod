module github.com/up9inc/mizu/tap/extensions/http

go 1.16

require (
	github.com/google/martian v2.1.0+incompatible
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7
	github.com/up9inc/mizu/shared v0.0.0
	github.com/up9inc/mizu/tap/api v0.0.0
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0
	golang.org/x/net v0.0.0-20210908191846-a5e095526f91
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api

replace github.com/up9inc/mizu/shared v0.0.0 => ../../../shared