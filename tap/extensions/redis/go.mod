module github.com/up9inc/mizu/tap/extensions/redis

go 1.17

require github.com/up9inc/mizu/tap/api v0.0.0

require github.com/google/martian v2.1.0+incompatible // indirect

replace github.com/up9inc/mizu/tap/api v0.0.0 => ../../api
