module github.com/up9inc/mizu/tap/extensions/kafka

go 1.17

require (
	github.com/fatih/camelcase v1.0.0
	github.com/ohler55/ojg v1.12.12
	github.com/segmentio/kafka-go v0.4.27
	github.com/stretchr/testify v1.6.1
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/text v0.3.0
)

require (
	github.com/davecgh/go-spew v1.1.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/klauspost/compress v1.14.2 // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/up9inc/mizu/tap/dbgctl v0.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api

replace github.com/up9inc/mizu/tap/dbgctl v0.0.0 => ../../dbgctl
