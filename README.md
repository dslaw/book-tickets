# book-tickets

Ticketmaster clone.


## Getting Started

To get started, first build the project:

```bash
$ cp .env-dev .env
$ docker compose build
```

then, run database schema migrations to set up the Postgres database:

```bash
$ docker compose up -d
$ docker compose run --rm dbmigrations up
```


## Testing

Prior to testing, first ensure that the test database is set up:
```bash
$ docker compose up -d
$ docker compose run --rm dbmigrations -e TEST_DATABASE_URL up
```

and ensure that the test database url is exposed to the test suite:
```bash
$ set -a && source .env-dev
```

then, run tests:
```bash
$ go test ./...
```
