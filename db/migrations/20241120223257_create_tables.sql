-- migrate:up
create table users (
    id int generated always as identity,
    name varchar(20) not null unique,
    email varchar(100) not null unique,

    unique (name, email),
    primary key (id)
);

create table venues (
    id int generated always as identity,
    name text not null,
    description text,
    address text not null,
    -- Longest city name in the world, as of 2024, has 58 characters.
    city varchar(60) not null,
    subdivision varchar(60) not null,
    -- ISO 3166-1 alpha3
    country_code char(3) not null,
    deleted boolean not null default false,

    unique (name, address),
    primary key (id)
);

create table performers (
    id int generated always as identity,
    name varchar(50) not null check (char_length(name) > 0),

    primary key (id)
);

create table events (
    id int generated always as identity,
    venue_id int not null,
    name varchar(50) not null check (char_length(name) > 0),
    starts_at timestamptz not null,
    ends_at timestamptz not null,
    description text,
    deleted boolean not null default false,

    unique (venue_id, name, starts_at, ends_at),
    foreign key (venue_id) references venues (id),
    primary key (id)
);

create table event_performers (
    id int generated always as identity,
    event_id int not null,
    performer_id int not null,

    unique (event_id, performer_id),
    primary key (id)
);

create table tickets (
    id int generated always as identity,
    event_id int not null,
    purchaser_id int,
    price int not null check (price > 0),
    seat varchar(10) not null check (char_length(seat) > 0),

    foreign key (event_id) references events (id),
    foreign key (purchaser_id) references users (id),
    primary key (id)
);


-- migrate:down
drop table tickets;
drop table event_performers;
drop table events;
drop table performers;
drop table venues;
drop table users;
