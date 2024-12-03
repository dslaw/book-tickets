// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

const createEvent = `-- name: CreateEvent :one
insert into events (venue_id, name, starts_at, ends_at, description)
values ($1, $2, $3, $4, $5)
returning id
`

type CreateEventParams struct {
	VenueID     int32
	Name        string
	StartsAt    pgtype.Timestamptz
	EndsAt      pgtype.Timestamptz
	Description pgtype.Text
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) (int32, error) {
	row := q.db.QueryRow(ctx, createEvent,
		arg.VenueID,
		arg.Name,
		arg.StartsAt,
		arg.EndsAt,
		arg.Description,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const createVenue = `-- name: CreateVenue :one
insert into venues (name, description, address, city, subdivision, country_code)
values ($1, $2, $3, $4, $5, $6)
returning id
`

type CreateVenueParams struct {
	Name        string
	Description pgtype.Text
	Address     string
	City        string
	Subdivision string
	CountryCode string
}

func (q *Queries) CreateVenue(ctx context.Context, arg CreateVenueParams) (int32, error) {
	row := q.db.QueryRow(ctx, createVenue,
		arg.Name,
		arg.Description,
		arg.Address,
		arg.City,
		arg.Subdivision,
		arg.CountryCode,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const deleteEvent = `-- name: DeleteEvent :one
with delete_event as (
    update events
    set deleted = true
    where
        id = $1
        and deleted = false
    returning id
)
select count(*) from delete_event
`

func (q *Queries) DeleteEvent(ctx context.Context, eventID int32) (int64, error) {
	row := q.db.QueryRow(ctx, deleteEvent, eventID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const deleteVenue = `-- name: DeleteVenue :one
with delete_events as (
    -- Cascade delete to events.
    update events
    set deleted = true
    where venue_id = $1
), delete_venue as (
    update venues
    set deleted = true
    where
        id = $1
        and deleted = false
    returning id
)
select count(*) from delete_venue
`

func (q *Queries) DeleteVenue(ctx context.Context, venueID int32) (int64, error) {
	row := q.db.QueryRow(ctx, deleteVenue, venueID)
	var count int64
	err := row.Scan(&count)
	return count, err
}

const getAvailableTickets = `-- name: GetAvailableTickets :many
select tickets.id, tickets.event_id, tickets.purchaser_id, tickets.price, tickets.seat
from tickets
inner join events on tickets.event_id = events.id
where 
    tickets.purchaser_id is null
    and tickets.event_id = $1
    and events.deleted = false
`

type GetAvailableTicketsRow struct {
	Ticket Ticket
}

func (q *Queries) GetAvailableTickets(ctx context.Context, eventID int32) ([]GetAvailableTicketsRow, error) {
	rows, err := q.db.Query(ctx, getAvailableTickets, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetAvailableTicketsRow
	for rows.Next() {
		var i GetAvailableTicketsRow
		if err := rows.Scan(
			&i.Ticket.ID,
			&i.Ticket.EventID,
			&i.Ticket.PurchaserID,
			&i.Ticket.Price,
			&i.Ticket.Seat,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEvent = `-- name: GetEvent :many
select
    events.id, events.venue_id, events.name, events.starts_at, events.ends_at, events.description, events.deleted,
    venues.name as venue_name,
    performers.id as performer_id,
    performers.name as performer_name
from events
inner join venues on events.venue_id = venues.id
left outer join event_performers on events.id = event_performers.event_id
left outer join performers on event_performers.performer_id = performers.id
where
    events.id = $1
    and events.deleted = false
    and venues.deleted = false
`

type GetEventRow struct {
	Event         Event
	VenueName     string
	PerformerID   pgtype.Int4
	PerformerName pgtype.Text
}

func (q *Queries) GetEvent(ctx context.Context, eventID int32) ([]GetEventRow, error) {
	rows, err := q.db.Query(ctx, getEvent, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetEventRow
	for rows.Next() {
		var i GetEventRow
		if err := rows.Scan(
			&i.Event.ID,
			&i.Event.VenueID,
			&i.Event.Name,
			&i.Event.StartsAt,
			&i.Event.EndsAt,
			&i.Event.Description,
			&i.Event.Deleted,
			&i.VenueName,
			&i.PerformerID,
			&i.PerformerName,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getVenue = `-- name: GetVenue :one
select venues.id, venues.name, venues.description, venues.address, venues.city, venues.subdivision, venues.country_code, venues.deleted
from venues
where
    id = $1
    and deleted = false
`

type GetVenueRow struct {
	Venue Venue
}

func (q *Queries) GetVenue(ctx context.Context, venueID int32) (GetVenueRow, error) {
	row := q.db.QueryRow(ctx, getVenue, venueID)
	var i GetVenueRow
	err := row.Scan(
		&i.Venue.ID,
		&i.Venue.Name,
		&i.Venue.Description,
		&i.Venue.Address,
		&i.Venue.City,
		&i.Venue.Subdivision,
		&i.Venue.CountryCode,
		&i.Venue.Deleted,
	)
	return i, err
}

const linkUpdatedPerformers = `-- name: LinkUpdatedPerformers :exec
with performer_ids as (
    select id
    from performers
    where name = any($2::text[])
), del as (
    delete from event_performers
    where
        event_id = $1
        and not exists (
            select 1
            from performer_ids
            where performer_ids.id = event_performers.performer_id
        )
)
insert into event_performers (event_id, performer_id)
select $1, id
from performer_ids
on conflict (event_id, performer_id) do nothing
`

type LinkUpdatedPerformersParams struct {
	EventID int32
	Names   []string
}

func (q *Queries) LinkUpdatedPerformers(ctx context.Context, arg LinkUpdatedPerformersParams) error {
	_, err := q.db.Exec(ctx, linkUpdatedPerformers, arg.EventID, arg.Names)
	return err
}

const trimUpdatedEventPerformers = `-- name: TrimUpdatedEventPerformers :exec
delete from event_performers
where event_id = $1
`

func (q *Queries) TrimUpdatedEventPerformers(ctx context.Context, eventID int32) error {
	_, err := q.db.Exec(ctx, trimUpdatedEventPerformers, eventID)
	return err
}

const updateEvent = `-- name: UpdateEvent :one
update events
set
    name = $1,
    starts_at = $2,
    ends_at = $3,
    description = $4
where
    id = $5
    and deleted = false
returning id
`

type UpdateEventParams struct {
	Name        string
	StartsAt    pgtype.Timestamptz
	EndsAt      pgtype.Timestamptz
	Description pgtype.Text
	EventID     int32
}

// The updated record's id is returned so that the generated query will return
// an error (`sql.ErrNoRows`) if no record matches the where clause and no
// record is updated.
func (q *Queries) UpdateEvent(ctx context.Context, arg UpdateEventParams) (int32, error) {
	row := q.db.QueryRow(ctx, updateEvent,
		arg.Name,
		arg.StartsAt,
		arg.EndsAt,
		arg.Description,
		arg.EventID,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}

const updateVenue = `-- name: UpdateVenue :one
update venues
set
    name = $1,
    description = $2,
    address = $3,
    city = $4,
    subdivision = $5,
    country_code = $6
where
    id = $7
    and deleted = false
returning id
`

type UpdateVenueParams struct {
	Name        string
	Description pgtype.Text
	Address     string
	City        string
	Subdivision string
	CountryCode string
	VenueID     int32
}

// The updated record's id is returned so that the generated query will return
// an error (`sql.ErrNoRows`) if no record matches the where clause and no
// record is updated.
func (q *Queries) UpdateVenue(ctx context.Context, arg UpdateVenueParams) (int32, error) {
	row := q.db.QueryRow(ctx, updateVenue,
		arg.Name,
		arg.Description,
		arg.Address,
		arg.City,
		arg.Subdivision,
		arg.CountryCode,
		arg.VenueID,
	)
	var id int32
	err := row.Scan(&id)
	return id, err
}
