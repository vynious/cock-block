build:
	go build -o bin/cock-block

run: build
	./bin/cock-block

test: 
	go test -v ./...