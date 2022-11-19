skipbin := $$(find bin -mindepth 1 -maxdepth 1)
skipexpect := $$(find expect -mindepth 1 -maxdepth 1)

test: test-pull-bin test-pull-expect
	@KUBESHARK_TEST=1 go test -v ./... -coverpkg=./... -race -coverprofile=coverage.out -covermode=atomic

test-update: test-pull-bin
	@KUBESHARK_TEST=1 TEST_UPDATE=1 go test -v ./... -coverpkg=./... -coverprofile=coverage.out -covermode=atomic

test-pull-bin:
	@[ "${skipbin}" ] && echo "Skipping downloading BINs" || ../get-folder-of-tests.sh bin/http bin

test-pull-expect:
	@[ "${skipexpect}" ] && echo "Skipping downloading expected JSONs" || ../get-folder-of-tests.sh expect/http expect
