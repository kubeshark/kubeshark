skipbin := $$(find bin -mindepth 1 -maxdepth 1)
skipexpect := $$(find expect -mindepth 1 -maxdepth 1)

test: test-pull-bin test-pull-expect
	@MIZU_TEST=1 go test -v ./... -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic

test-update: test-pull-bin
	@MIZU_TEST=1 TEST_UPDATE=1 go test -v ./... -coverpkg=./... -coverprofile=coverage.out -covermode=atomic

test-pull-bin:
	@mkdir -p bin
	@[ "${skipbin}" ] && echo "Skipping downloading BINs" || gsutil  -o 'GSUtil:parallel_process_count=5' -o 'GSUtil:parallel_thread_count=5'  -m cp gs://static.up9.io/mizu/test-pcap/bin/kafka/\*.bin bin

test-pull-expect:
	@mkdir -p expect
	@[ "${skipexpect}" ] && echo "Skipping downloading expected JSONs" || gsutil  -o 'GSUtil:parallel_process_count=5' -o 'GSUtil:parallel_thread_count=5'  -m cp -r gs://static.up9.io/mizu/test-pcap/expect9/kafka/\* expect
