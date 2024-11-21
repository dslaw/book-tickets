# book-tickets

Ticketmaster clone.


## Getting Started

To get started, first build the project:

```bash
$ cp .env-dev .env
$ docker-compose build
```

then, run database schema migrations to set up the Postgres database:

```bash
$ docker-compose up -d
$ docker-compose run --rm dbmigrations up
```
