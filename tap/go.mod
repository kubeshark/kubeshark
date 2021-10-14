module github.com/up9inc/mizu/tap

go 1.16

require (
	github.com/bradleyfalzon/tlsx v0.0.0-20170624122154-28fd0e59bac4
	github.com/go-errors/errors v1.4.1
	github.com/google/gopacket v1.1.19
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7
	github.com/up9inc/mizu/tap/api v0.0.0
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7 // indirect
	golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073 // indirect
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ./api
