module github.com/up9inc/mizu/tap/extensions/kafka

go 1.17

require (
	github.com/fatih/camelcase v1.0.0
	github.com/ohler55/ojg v1.12.12
	github.com/segmentio/kafka-go v0.4.27
	github.com/stretchr/testify v1.7.0
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/text v0.3.7
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/frankban/quicktest v1.14.0 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/go-cmp v0.5.7 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/klauspost/compress v1.14.2 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pierrec/lz4 v2.6.1+incompatible // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api

replace github.com/up9inc/mizu/logger v0.0.0 => ../../../logger
