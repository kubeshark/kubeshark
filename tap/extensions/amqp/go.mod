module github.com/kubeshark/kubeshark/tap/extensions/amqp

go 1.17

require (
	github.com/kubeshark/kubeshark/logger v0.0.0
	github.com/kubeshark/kubeshark/tap/api v0.0.0
	github.com/stretchr/testify v1.7.0
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/kubeshark/kubeshark/tap/dbgctl v0.0.0 // indirect
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/kubeshark/kubeshark/logger v0.0.0 => ../../../logger

replace github.com/kubeshark/kubeshark/tap/api v0.0.0 => ../../api

replace github.com/kubeshark/kubeshark/tap/dbgctl v0.0.0 => ../../dbgctl
