.PHONY: default clean deps update-deps pretty test start coverage coverage-profile coverage-html integration-test

default: clean deps pretty build test

clean:
	rm -rf _coverage

deps:
	go mod download

update-deps:
	go get -u ./...
	go mod tidy

pretty:
	gofmt -l -s -w .

build:
	go build -v ./...

integration-test:
	go test -v --tags=integration --count=1 ./it/...

test:
	go test ./...

test-single:
	go test -v ./... -run $(test_regex)

coverage:
	go test -cover ./...

_coverage:
	mkdir _coverage

coverage-profile: _coverage
	go test -count=1 -coverprofile=_coverage/coverage.out ./...

coverage-html: coverage-profile
	go tool cover -html=_coverage/coverage.out -o _coverage/coverage.html

start:
	go run cmd/server/main.go
