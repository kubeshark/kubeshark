module github.com/up9inc/mizu/tap/api

go 1.17

require (
	github.com/google/gopacket v1.1.19
	github.com/google/martian v2.1.0+incompatible
	github.com/up9inc/mizu/shared v0.0.0
)

require github.com/op/go-logging v0.0.0-20160315200505-970db520ece7 // indirect

replace github.com/up9inc/mizu/shared v0.0.0 => ../../shared
