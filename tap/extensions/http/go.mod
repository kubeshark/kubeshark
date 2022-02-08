module github.com/up9inc/mizu/tap/extensions/http

go 1.17

require (
	github.com/beevik/etree v1.1.0
	github.com/stretchr/testify v1.7.0
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.3.5 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
