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
-- The updated record's id is returned so that the generated query will return
-- an error (`sql.ErrNoRows`) if no record matches the where clause and no
-- record is updated.
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
), delete_venue as (
    update venues
    set deleted = true
    where
        id = @venue_id
        and deleted = false
    returning id
)
select count(*) from delete_venue;

-- name: WritePerformers :batchexec
insert into performers (name) values (@name)
on conflict (name) do nothing;

-- name: LinkPerformers :batchexec
insert into event_performers (event_id, performer_id)
select @event_id, performers.id
from performers
where name = @name;

-- name: CreateEvent :one
insert into events (venue_id, name, starts_at, ends_at, description)
values (@venue_id, @name, @starts_at, @ends_at, @description)
returning id;

-- name: GetEvent :many
select
    sqlc.embed(events),
    venues.name as venue_name,
    performers.id as performer_id,
    performers.name as performer_name
from events
inner join venues on events.venue_id = venues.id
left outer join event_performers on events.id = event_performers.event_id
left outer join performers on event_performers.performer_id = performers.id
where
    events.id = @event_id
    and events.deleted = false
    and venues.deleted = false;

-- name: UpdateEvent :one
-- The updated record's id is returned so that the generated query will return
-- an error (`sql.ErrNoRows`) if no record matches the where clause and no
-- record is updated.
update events
set
    name = @name,
    starts_at = @starts_at,
    ends_at = @ends_at,
    description = @description
where
    id = @event_id
    and deleted = false
returning id;

-- name: TrimUpdatedEventPerformers :exec
delete from event_performers
where event_id = @event_id;

-- name: LinkUpdatedPerformers :exec
with performer_ids as (
    select id
    from performers
    where name = any(@names::text[])
), del as (
    delete from event_performers
    where
        event_id = @event_id
        and not exists (
            select 1
            from performer_ids
            where performer_ids.id = event_performers.performer_id
        )
)
insert into event_performers (event_id, performer_id)
select @event_id, id
from performer_ids
on conflict (event_id, performer_id) do nothing;

-- name: DeleteEvent :one
with delete_event as (
    update events
    set deleted = true
    where
        id = @event_id
        and deleted = false
    returning id
)
select count(*) from delete_event;

-- name: GetTicket :one
select sqlc.embed(tickets)
from tickets
inner join events on tickets.event_id = events.id
where 
    tickets.id = @ticket_id
    and events.deleted = false;

-- name: GetAvailableTickets :many
select sqlc.embed(tickets)
from tickets
inner join events on tickets.event_id = events.id
where 
    tickets.purchaser_id is null
    and tickets.event_id = @event_id
    and events.deleted = false;

-- name: WriteNewTickets :batchone
-- The inserted record's id is returned so that the generated query will return
-- an error (`sql.ErrNoRows`) if no record is inserted due to the where clause
-- not finding a matching event.
insert into tickets (event_id, purchaser_id, price, seat)
select events.id, null, @price, @seat
from events
where
    events.id = @event_id
    and events.deleted = false
returning id;

-- name: SetTicketPurchaser :one
update tickets
set purchaser_id = @purchaser_id
where
    id = @ticket_id
    and purchaser_id is null
returning id;
