module github.com/up9inc/mizu/tap

go 1.16

require (
	github.com/bradleyfalzon/tlsx v0.0.0-20170624122154-28fd0e59bac4
	github.com/go-errors/errors v1.4.1
	github.com/google/gopacket v1.1.19
	github.com/op/go-logging v0.0.0-20160315200505-970db520ece7
	github.com/up9inc/mizu/shared v0.0.0
	github.com/up9inc/mizu/tap/api v0.0.0
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ./api

replace github.com/up9inc/mizu/shared v0.0.0 => ../shared
