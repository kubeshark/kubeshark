module github.com/up9inc/mizu/tap/extensions/kafka

go 1.16

require (
	github.com/fatih/camelcase v1.0.0
	github.com/klauspost/compress v1.14.1 // indirect
	github.com/ohler55/ojg v1.12.12
	github.com/segmentio/kafka-go v0.4.27
	github.com/up9inc/mizu/tap/api v0.0.0
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
