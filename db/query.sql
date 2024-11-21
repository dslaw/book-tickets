-- name: WriteNewVenue :one
insert into venues (name, description, address, city, subdivision, country_code)
values (@name, @description, @address, @city, @subdivision, @country_code)
returning id;

-- name: UpdateVenue :one
update venues
set
    name = @name,
    description = @description,
    address = @address,
    city = @city,
    subdivision = @subdivision,
    country_code = @country_code
where id = @venue_id
returning id;

-- name: GetVenue :one
select sqlc.embed(venues)
from venues
where id = @venue_id;
