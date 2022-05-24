module github.com/up9inc/mizu/tap/extensions/http

go 1.17

require (
	github.com/beevik/etree v1.1.0
	github.com/mertyildiran/gqlparser/v2 v2.4.6
	github.com/stretchr/testify v1.7.0
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/up9inc/mizu/tap/dbgctl v0.0.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api

replace github.com/up9inc/mizu/tap/dbgctl v0.0.0 => ../../dbgctl
