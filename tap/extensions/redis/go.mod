module github.com/up9inc/mizu/tap/extensions/redis

go 1.17

require (
	github.com/stretchr/testify v1.7.0
	github.com/up9inc/mizu/tap/api v0.0.0
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/up9inc/mizu/tap/dbgctl v0.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api

replace github.com/up9inc/mizu/tap/dbgctl v0.0.0 => ../../dbgctl
