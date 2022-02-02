module github.com/up9inc/mizu/integrationTests

go 1.16

require (
	github.com/up9inc/mizu/shared v0.0.0
)

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../tap/api
