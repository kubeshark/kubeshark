module github.com/up9inc/mizu/tap/extensions/kafka

go 1.16

require (
	github.com/segmentio/kafka-go v0.4.17
	github.com/up9inc/mizu/tap/api v0.0.0
	github.com/up9inc/mizu/shared v0.0.0
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api

replace github.com/up9inc/mizu/shared v0.0.0 => ../../../shared