module github.com/up9inc/mizu/tap/extensions/http

go 1.17

require (
	github.com/beevik/etree v1.1.0
	github.com/stretchr/testify v1.7.0
	github.com/up9inc/mizu/tap/api v0.0.0
	github.com/vektah/gqlparser/v2 v2.3.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd // indirect
	golang.org/x/net v0.0.0-20211209124913-491a49abca63 // indirect
	golang.org/x/sys v0.0.0-20210615035016-665e8c7367d1 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
