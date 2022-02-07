test: test-pull
	MIZU_TEST=1 go test *.go -covermode=atomic -coverprofile=coverage.out

test-update: test-pull
	MIZU_TEST=1 TEST_UPDATE=1 go test *.go -v -covermode=atomic -coverprofile=coverage.out

test-pull:
	./pull.sh
