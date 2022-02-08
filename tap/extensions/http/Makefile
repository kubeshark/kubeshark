test: test-pull-expect
	@MIZU_TEST=1 go test -v ./... -coverpkg=./... -coverprofile=coverage.out -covermode=atomic

test-update: test-pull-bin
	@MIZU_TEST=1 TEST_UPDATE=1 go test -v ./... -coverpkg=./... -coverprofile=coverage.out -covermode=atomic

test-pull-bin:
	@[ "$(ls -A --ignore='.??*' bin)" ] && echo "Skipping downloading BINs" || gsutil -m cp gs://static.up9.io/mizu/test-pcap/bin/http/\*.bin bin

test-pull-expect:
	@[ "$(ls -A --ignore='.??*' bin)" ] && echo "Skipping downloading BINs" || gsutil -m cp gs://static.up9.io/mizu/test-pcap/bin/http/\*.bin bin
