-- name: CreateVenue :one
insert into venues (name, description, address, city, subdivision, country_code)
values (@name, @description, @address, @city, @subdivision, @country_code)
returning id;

-- name: GetVenue :one
select sqlc.embed(venues)
from venues
where
    id = @venue_id
    and deleted = false;

-- name: UpdateVenue :one
update venues
set
    name = @name,
    description = @description,
    address = @address,
    city = @city,
    subdivision = @subdivision,
    country_code = @country_code
where
    id = @venue_id
    and deleted = false
returning id;

-- name: DeleteVenue :one
with delete_events as (
    -- Cascade delete to events.
    update events
    set deleted = true
    where venue_id = @venue_id
), delete_venues as (
    update venues
    set deleted = true
    where
        id = @venue_id
        and deleted = false
    returning id
)
select count(*) from delete_venues;
