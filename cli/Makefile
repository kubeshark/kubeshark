install:
	go build install mizu.go

build:
	go build -o bin/mizu mizu.go

build-cr:
	@echo "Compiling for every OS and Platform"
	@echo ""
	GOOS=darwin GOARCH=amd64 go build -o bin/mizu-darwin-amd64 mizu.go
	GOOS=linux GOARCH=amd64 go build -o bin/mizu-linux-amd64 mizu.go
	@#GOOS=windows GOARCH=amd64 go build -o bin/mizu-windows-amd64.exe mizu.go
	@#GOOS=linux GOARCH=386 go build -o bin/mizu-linux-386 mizu.go
	@#GOOS=windows GOARCH=386 go build -o bin/mizu-windows-386.exe mizu.go
	@#GOOS=darwin GOARCH=arm64 go build -o bin/mizu-darwin-arm64 mizu.go
	@#GOOS=linux GOARCH=arm64 go build -o bin/mizu-linux-arm64 mizu.go
	@#GOOS=windows GOARCH=arm64 go build -o bin/mizu-windows-arm64 mizu.go

clean: 
	#go clean
	rm -f ./bin/*
