.PHONY: all test

all: install test
install:
		@go install

test:
		@go test

release:
	@env GOOS=linux go build -o bin/rv.linux
	@env GOOS=darwin go build -o bin/rv.darwin
