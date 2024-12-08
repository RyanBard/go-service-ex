# go-service-ex

A microservice written in Go to serve as an example of unit tests, integration tests, interacting with a DB, etc.

## Dev Environment Setup

* Install [direnv](https://direnv.net/) (ex. for Ubuntu: `sudo apt install direnv && echo 'val "$(direnv hook bash)"' >> ~/.bashrc`)
* Configure direnv to load .env files (ex. `mkdir ~/.config/direnv && echo -e '[global]\nload_dotenv = true' > ~/.config/direnv/direnv.toml`)
* Copy the .env.example file to .env and load it with `direnv allow`
* Install [postgres](https://www.postgresql.org/) (ex. for Ubuntu: `sudo apt install postgresql`)
* Install [GVM](https://github.com/moovweb/gvm) and use go 1.17+ (ex. `gvm use go1.23`)

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
* extract common things into their own repo (tx manager, httpx client, etc.)
* branch coverage: https://github.com/junhwi/gobco/
* update golang version
* switch from logrus to slog
* add some example grpc/protobuf code
