module github.com/kubeshark/kubeshark/tap/extensions/http

go 1.17

require (
	github.com/google/martian v2.1.0+incompatible
	github.com/kubeshark/kubeshark/tap/api v0.0.0
	github.com/mertyildiran/gqlparser/v2 v2.4.6
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kubeshark/kubeshark/tap/dbgctl v0.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/kubeshark/kubeshark/tap/api v0.0.0 => ../../api

replace github.com/kubeshark/kubeshark/tap/dbgctl v0.0.0 => ../../dbgctl
