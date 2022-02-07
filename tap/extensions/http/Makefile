test: test-pull
	@MIZU_TEST=1 go test -v ./... -coverpkg=./... -coverprofile=coverage.out -covermode=atomic

test-update: test-pull
	@MIZU_TEST=1 TEST_UPDATE=1 go test -v ./... -coverpkg=./... -coverprofile=coverage.out -covermode=atomic

test-pull:
	@./pull.sh
