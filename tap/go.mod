module github.com/up9inc/mizu/tap

go 1.16

require (
	github.com/google/gopacket v1.1.19
	github.com/romana/rlog v0.0.0-20171115192701-f018bc92e7d7
	github.com/up9inc/mizu/tap/api v0.0.0
    golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
    golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073
    golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d
    golang.org/x/text v0.3.5
    golang.org/x/tools v0.0.0-20210106214847-113979e3529a
)

replace github.com/up9inc/mizu/tap/api v0.0.0 => ./api
