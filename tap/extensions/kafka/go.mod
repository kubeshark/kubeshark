module github.com/up9inc/mizu/tap/extensions/kafka

go 1.17

require (
	github.com/fatih/camelcase v1.0.0
	github.com/ohler55/ojg v1.12.12
	github.com/segmentio/kafka-go v0.4.27
	github.com/up9inc/mizu/tap/api v0.0.0
)

require (
	github.com/golang/snappy v0.0.1 // indirect
	github.com/google/martian v2.1.0+incompatible // indirect
	github.com/klauspost/compress v1.14.1 // indirect
	github.com/pierrec/lz4 v2.6.0+incompatible // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
