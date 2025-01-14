// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Event struct {
	ID          int32
	VenueID     int32
	Name        string
	StartsAt    pgtype.Timestamptz
	EndsAt      pgtype.Timestamptz
	Description pgtype.Text
	Deleted     bool
}

type EventPerformer struct {
	ID          int32
	EventID     int32
	PerformerID int32
}

type Performer struct {
	ID   int32
	Name string
}

type Ticket struct {
	ID          int32
	EventID     int32
	PurchaserID pgtype.Int4
	Price       int32
	Seat        string
}

type User struct {
	ID    int32
	Name  string
	Email string
}

type Venue struct {
	ID          int32
	Name        string
	Description pgtype.Text
	Address     string
	City        string
	Subdivision string
	CountryCode string
	Deleted     bool
}
