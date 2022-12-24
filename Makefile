.PHONY: default clean deps pretty test start curl coverage coverage-profile coverage-html integration-test

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

curl:
	curl -v -H 'x-request-id: ryan-test-health' http://localhost:4000/health
	curl -v -H 'x-request-id: ryan-test-readiness' http://localhost:4000/readiness
	curl -v -H 'x-request-id: ryan-test-get-by-id-no-auth' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-get-by-id-empty-auth' -H 'Authorization: ' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-get-by-id-basic-auth' -H 'Authorization: Basic foo' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-get-by-id-bearer-too-many-parts' -H 'Authorization: Bearer Bearer foo' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-get-by-id-invalid-token' -H 'Authorization: Bearer invalid' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-get-by-id-token-with-leading-trailing-spaces' -H 'Authorization: Bearer     foo   ' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-get-by-id-valid-token' -H 'Authorization: Bearer foo' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-save-post-no-id' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-save-put-no-id' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-save-post-id-in-body-only' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-save-put-id-in-body-only' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-save-post-id-in-path-only' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-save-put-id-in-path-only' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-save-post-id-in-both-match' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-save-put-id-in-both-match' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-save-post-id-in-both-mismatch' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"id": "456", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-save-put-id-in-both-mismatch' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"id": "456", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-save-invalid-json' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-save-missing-name' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"desc": "foo-d"}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-save-missing-desc' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"name": "foo-n"}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-save-missing-name-and-desc' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-find-all' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-find-all-matching-name' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' http://localhost:4000/api/orgs?name=foobar
	curl -v -H 'x-request-id: ryan-test-delete' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X DELETE -d '{"id": "123", "name": "foo", "desc": "bar", "is_archived": false, "version": 1}' http://localhost:4000/api/orgs/123
