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
