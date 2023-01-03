# go-service-ex

A microservice written in Go to serve as an example of unit tests, integration tests, interacting with a DB, etc.

## DB Setup

```
./db/local-setup.sh
```

## Formatting, Building, Testing

```
make
```

## Generating Coverage Report

```
make coverate-html
open _coverage/coverage.html
```

## Running

```
# load any necessary environment variables: ex. JWT_SECRET, DB_USER, DB_PASSWORD, DB_SSL_MODE, etc.
make start
```

## Integration Tests

```
# load any necessary environment variables: ex. JWT_SECRET, etc.
make integration-test
```

## TODO
* add a tx example (setup user and org in 1 tx)
* add prometheus metrics
