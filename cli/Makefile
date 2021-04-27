install:
	go build -o bin/main main.go
	cp bin/main /usr/local/bin/mizu

build:
	go build -o bin/main main.go

build-cr:
	echo "Compiling for every OS and Platform"
	GOOS=linux GOARCH=386 go build -o bin/mizu-linux-386 main.go
	GOOS=windows GOARCH=386 go build -o bin/mizu-windows-386.exe main.go

	GOOS=darwin GOARCH=amd64 go build -o bin/mizu-darwin-amd64 main.go
	GOOS=linux GOARCH=amd64 go build -o bin/mizu-linux-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o bin/mizu-windows-amd64.exe main.go

	GOOS=darwin GOARCH=arm64 go build -o bin/mizu-darwin-arm64 main.go
	GOOS=linux GOARCH=arm64 go build -o bin/mizu-linux-arm64 main.go
	GOOS=windows GOARCH=arm64 go build -o bin/mizu-windows-arm64 main.go

run:
	go run main.go ".*"
