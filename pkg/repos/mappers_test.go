package repos_test

import (
	"testing"
	"time"

	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestMapNullableString(t *testing.T) {
	type TestInput struct {
		String        string
		ExpectedValid bool
	}
	for _, testInput := range []TestInput{
		{String: "", ExpectedValid: false},
		{String: "text", ExpectedValid: true},
	} {
		actual := repos.MapNullableString(testInput.String)
		assert.Equal(t, testInput.String, actual.String)
		assert.Equal(t, testInput.ExpectedValid, actual.Valid)
	}
}

func TestMapTime(t *testing.T) {
	val := time.Now()
	actual := repos.MapTime(val)
	assert.Equal(t, val, actual.Time)
	assert.Equal(t, true, actual.Valid)
}

func TestMapPurchaserID(t *testing.T) {
	purchaserID := int32(1)
	actual := repos.MapPurchaserID(purchaserID)
	assert.Equal(t, pgtype.Int4{Int32: purchaserID, Valid: true}, actual)
}

func TestMapGetEventRows(t *testing.T) {
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-01")

	rows := []db.GetEventRow{
		{
			Event: db.Event{
				ID:          eventID,
				VenueID:     1,
				Name:        "Test Event",
				StartsAt:    pgtype.Timestamptz{Time: startsAt, Valid: true},
				EndsAt:      pgtype.Timestamptz{Time: endsAt, Valid: true},
				Description: pgtype.Text{String: "", Valid: false},
				Deleted:     false,
			},
			VenueName:     "Test Venue",
			PerformerID:   pgtype.Int4{Int32: 1, Valid: true},
			PerformerName: pgtype.Text{String: "Test Performer 1", Valid: true},
		},
		{
			Event: db.Event{
				ID:          eventID,
				VenueID:     1,
				Name:        "Test Event",
				StartsAt:    pgtype.Timestamptz{Time: startsAt, Valid: true},
				EndsAt:      pgtype.Timestamptz{Time: endsAt, Valid: true},
				Description: pgtype.Text{String: "", Valid: false},
				Deleted:     false,
			},
			VenueName:     "Test Venue",
			PerformerID:   pgtype.Int4{Int32: 2, Valid: true},
			PerformerName: pgtype.Text{String: "Test Performer 2", Valid: true},
		},
	}
	expected := entities.Event{
		ID:          eventID,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue: entities.EventVenue{
			ID:   1,
			Name: "Test Venue",
		},
		Performers: []entities.Performer{
			{ID: 1, Name: "Test Performer 1"},
			{ID: 2, Name: "Test Performer 2"},
		},
	}

	actual := repos.MapGetEventRows(rows)
	assert.EqualValues(t, expected, actual)
}

func TestMapGetEventRowsWhenNoPerformers(t *testing.T) {
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-01")

	rows := []db.GetEventRow{
		{
			Event: db.Event{
				ID:          eventID,
				VenueID:     1,
				Name:        "Test Event",
				StartsAt:    pgtype.Timestamptz{Time: startsAt, Valid: true},
				EndsAt:      pgtype.Timestamptz{Time: endsAt, Valid: true},
				Description: pgtype.Text{String: "", Valid: false},
				Deleted:     false,
			},
			VenueName:     "Test Venue",
			PerformerID:   pgtype.Int4{Int32: 0, Valid: false},
			PerformerName: pgtype.Text{String: "", Valid: false},
		},
	}
	expected := entities.Event{
		ID:          eventID,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue: entities.EventVenue{
			ID:   1,
			Name: "Test Venue",
		},
		Performers: []entities.Performer{},
	}

	actual := repos.MapGetEventRows(rows)
	assert.EqualValues(t, expected, actual)
}

func TestMapGetEventRowsWhenEmptyResultSet(t *testing.T) {
	rows := []db.GetEventRow{}
	actual := repos.MapGetEventRows(rows)
	assert.Empty(t, actual)
}

func TestMapTicket(t *testing.T) {
	model := db.Ticket{
		ID:          int32(1),
		EventID:     int32(11),
		PurchaserID: pgtype.Int4{Int32: 0, Valid: false},
		Price:       int32(25),
		Seat:        "GA",
	}
	expected := entities.Ticket{
		ID:          int32(1),
		EventID:     int32(11),
		PurchaserID: int32(0),
		IsPurchased: false,
		Price:       uint8(25),
		Seat:        "GA",
	}

	actual := repos.MapTicket(model)
	assert.EqualValues(t, expected, actual)
}

func TestMapGetAvailableTicketRows(t *testing.T) {
	rows := []db.GetAvailableTicketsRow{
		{Ticket: db.Ticket{ID: 1, EventID: 1, PurchaserID: pgtype.Int4{Int32: 1, Valid: true}, Price: 10, Seat: "GA"}},
		{Ticket: db.Ticket{ID: 2, EventID: 1, PurchaserID: pgtype.Int4{Int32: 0, Valid: false}, Price: 10, Seat: "GA"}},
		{Ticket: db.Ticket{ID: 3, EventID: 1, PurchaserID: pgtype.Int4{Int32: 0, Valid: false}, Price: 20, Seat: "Balcony"}},
	}
	expected := []entities.Ticket{
		{ID: 1, EventID: 1, PurchaserID: 1, IsPurchased: true, Price: uint8(10), Seat: "GA"},
		{ID: 2, EventID: 1, PurchaserID: 0, IsPurchased: false, Price: uint8(10), Seat: "GA"},
		{ID: 3, EventID: 1, PurchaserID: 0, IsPurchased: false, Price: uint8(20), Seat: "Balcony"},
	}

	actual := repos.MapGetAvailableTicketRows(rows)
	assert.EqualValues(t, expected, actual)
}

func TestMapGetAvailableTicketRowsWhenEmptyResultSet(t *testing.T) {
	rows := []db.GetAvailableTicketsRow{}
	actual := repos.MapGetAvailableTicketRows(rows)
	assert.Empty(t, actual)
}
