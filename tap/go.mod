module github.com/up9inc/mizu/tap

go 1.16

require (
	github.com/google/gopacket v1.1.19
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/net v0.0.0-20210421230115-4e50805a0758 // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ./api
