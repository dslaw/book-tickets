# book-tickets

Ticketmaster clone.


## Getting Started

Before starting development, `go`, `sqlc`, `docker compose` and `curl` will need
to be installed.

To get started, first build the project:

```bash
$ ln -s .env-dev .env
$ docker compose build
```

run database schema migrations to set up the Postgres database:

```bash
$ docker compose run --rm dbmigrations up
```

then, set up OpenSearch indexes:

```bash
$ ./search/create_indexes.sh
```

finally, run the stack:

```bash
$ docker compose up -d
```


## Testing

Prior to testing, first ensure that the test database is set up:
```bash
$ docker compose run --rm dbmigrations -e TEST_DATABASE_URL up
```

then, create indexes for integration testing against OpenSearch:
```bash
$ ./search/create_indexes.sh test
```

and ensure that the environment variables for testing are exposed to the test
suite:
```bash
$ set -a && source .env-dev
```

then, run tests (the `-short` flag may be used to skip integration tests):
```bash
$ go test ./...
```
