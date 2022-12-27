.PHONY: default clean deps pretty test start coverage coverage-profile coverage-html integration-test

default: clean deps pretty build test

clean:
	rm -rf _coverage

deps:
	go mod download

pretty:
	gofmt -l -s -w .

build:
	go build -v ./...

integration-test:
	go test -v --tags=integration --count=1 ./it/...

test:
	go test ./...

coverage:
	go test -cover ./...

_coverage:
	mkdir _coverage

coverage-profile: _coverage
	go test -coverprofile=_coverage/coverage.out ./...

coverage-html: coverage-profile
	go tool cover -html=_coverage/coverage.out -o _coverage/coverage.html

start:
	go run cmd/server/main.go
