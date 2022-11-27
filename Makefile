.PHONY: default clean pretty test start curl

default: clean pretty build test

build:
	go build ./...

clean:
	echo TODO: fix this after fixing build

pretty:
	gofmt -l -s -w .

test:
	go test ./...

start:
	go run cmd/server/main.go

curl:
	curl -v -H 'x-request-id: ryan-test-01' http://localhost:4000/health
	curl -v -H 'x-request-id: ryan-test-02' http://localhost:4000/readiness
	curl -v -H 'x-request-id: ryan-test-03' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-04' -H 'Authorization: ' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-05' -H 'Authorization: Basic foo' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-06' -H 'Authorization: Bearer Bearer foo' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-07' -H 'Authorization: Bearer invalid' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-08' -H 'Authorization: Bearer     foo   ' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-09' -H 'Authorization: Bearer foo' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-10' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-11' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-12' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-13' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-14' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-15' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-16' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-17' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"id": "123", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-18' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{"id": "456", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-19' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X PUT -d '{"id": "456", "name": "foo-n", "desc": "foo-d", "is_archived": false}' http://localhost:4000/api/orgs/123
	curl -v -H 'x-request-id: ryan-test-20' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -d '{' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-21' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' http://localhost:4000/api/orgs
	curl -v -H 'x-request-id: ryan-test-22' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' http://localhost:4000/api/orgs?name=foobar
	curl -v -H 'x-request-id: ryan-test-23' -H 'Authorization: Bearer foo' -H 'Content-Type: application/json' -X DELETE http://localhost:4000/api/orgs/123
