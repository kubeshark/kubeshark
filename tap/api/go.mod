module github.com/up9inc/mizu/tap/api

go 1.17

require (
	github.com/google/martian v2.1.0+incompatible
	github.com/up9inc/mizu/tap/dbgctl v0.0.0
)

replace github.com/up9inc/mizu/tap/dbgctl v0.0.0 => ../dbgctl
