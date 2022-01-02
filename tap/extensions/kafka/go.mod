module github.com/up9inc/mizu/tap/extensions/kafka

go 1.16

require (
	github.com/klauspost/compress v1.13.6 // indirect, see https://github.com/klauspost/compress/issues/428
	github.com/segmentio/kafka-go v0.4.17
	github.com/up9inc/mizu/tap/api v0.0.0
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
