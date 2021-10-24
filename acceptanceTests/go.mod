module github.com/up9inc/mizu/tests

go 1.16

require (
	github.com/up9inc/mizu/shared v0.0.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared
