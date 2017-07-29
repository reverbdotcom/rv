.PHONY: all test

all: install test
install:
		@go install

test:
		@go test

release:
	@env GOOS=linux go build
	@env GOOS=darwin go build
