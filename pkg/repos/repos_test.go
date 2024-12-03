package repos_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const venueID = int32(1)
const eventID = int32(1)

func TestVenuesRepoCreateVenue(t *testing.T) {
	ctx := context.Background()
	params := db.CreateVenueParams{
		Name:        "Test Venue",
		Description: pgtype.Text{String: "Test", Valid: true},
		Address:     "11 Front Street",
		City:        "San Francisco",
		Subdivision: "CA",
		CountryCode: "USA",
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("CreateVenue", ctx, params).Return(venueID, nil)

	// Tests that these values are mapped correctly.
	venue := entities.Venue{
		Name:        "Test Venue",
		Description: "Test",
		Location: entities.VenueLocation{
			Address:     "11 Front Street",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
		},
	}

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	actual, err := repo.CreateVenue(ctx, venue)

	assert.Equal(t, venueID, actual)
	assert.Nil(t, err)
	mockQueries.AssertCalled(t, "CreateVenue", ctx, params)
}

func TestVenuesRepoCreateVenueWhenTableConstraintViolation(t *testing.T) {
	fakeErr := errors.New("Unique constraint violated")

	mockQueries := new(MockQuerier)
	mockQueries.On("CreateVenue", mock.Anything, mock.Anything).Return(venueID, fakeErr)

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	_, err := repo.CreateVenue(context.Background(), entities.Venue{})

	assert.NotNil(t, err) // TODO: Update when error is mapped.
}

func TestVenuesRepoGetVenue(t *testing.T) {
	ret := db.GetVenueRow{
		Venue: db.Venue{
			ID:          venueID,
			Name:        "Test Venue",
			Description: pgtype.Text{String: "Test", Valid: true},
			Address:     "11 Front Street",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
		},
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("GetVenue", mock.Anything, venueID).Return(ret, nil)

	expected := entities.Venue{
		ID:          venueID,
		Name:        "Test Venue",
		Description: "Test",
		Location: entities.VenueLocation{
			Address:     "11 Front Street",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
		},
	}

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	actual, err := repo.GetVenue(context.Background(), venueID)

	assert.EqualValues(t, expected, actual)
	assert.Nil(t, err)
}

func TestVenuesRepoGetVenueWhenNotFoundOrMarkedDeleted(t *testing.T) {
	mockQueries := new(MockQuerier)
	mockQueries.On("GetVenue", mock.Anything, venueID).Return(db.GetVenueRow{}, sql.ErrNoRows)

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	actual, err := repo.GetVenue(context.Background(), venueID)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, repos.ErrNoSuchEntity)
}

func TestVenuesRepoUpdateVenue(t *testing.T) {
	ctx := context.Background()
	params := db.UpdateVenueParams{
		Name:        "Test Venue",
		Description: pgtype.Text{String: "Test", Valid: true},
		Address:     "11 Front Street",
		City:        "San Francisco",
		Subdivision: "CA",
		CountryCode: "USA",
		VenueID:     venueID,
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("UpdateVenue", ctx, params).Return(venueID, nil)

	// Tests that these values are mapped correctly.
	venue := entities.Venue{
		ID:          venueID,
		Name:        "Test Venue",
		Description: "Test",
		Location: entities.VenueLocation{
			Address:     "11 Front Street",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
		},
	}

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	err := repo.UpdateVenue(ctx, venue)

	assert.Nil(t, err)
	mockQueries.AssertCalled(t, "UpdateVenue", ctx, params)
}

func TestVenuesRepoUpdateVenueWhenNoRecord(t *testing.T) {
	mockQueries := new(MockQuerier)
	mockQueries.On("UpdateVenue", mock.Anything, mock.Anything).Return(venueID, sql.ErrNoRows)

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	err := repo.UpdateVenue(context.Background(), entities.Venue{})

	assert.ErrorIs(t, err, repos.ErrNoSuchEntity)
}

func TestVenuesRepoDeleteVenue(t *testing.T) {
	ctx := context.Background()
	mockQueries := new(MockQuerier)
	mockQueries.On("DeleteVenue", ctx, venueID).Return(int64(1), nil)

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	err := repo.DeleteVenue(ctx, venueID)

	assert.Nil(t, err)
	mockQueries.AssertCalled(t, "DeleteVenue", ctx, venueID)
}

func TestVenuesRepoDeleteVenueWhenDoesntExistOrDeleted(t *testing.T) {
	mockQueries := new(MockQuerier)
	mockQueries.On("DeleteVenue", mock.Anything, venueID).Return(int64(0), nil)

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	err := repo.DeleteVenue(context.Background(), venueID)

	assert.ErrorIs(t, err, repos.ErrNoSuchEntity)
}

func TestEventsRepoExecCreateEvent(t *testing.T) {
	ctx := context.Background()
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-01")

	createEventParams := db.CreateEventParams{
		VenueID:     venueID,
		Name:        "Test Event",
		StartsAt:    pgtype.Timestamptz{Time: startsAt, Valid: true},
		EndsAt:      pgtype.Timestamptz{Time: endsAt, Valid: true},
		Description: pgtype.Text{String: "", Valid: false},
	}
	writePerformersParams := []string{"Test Performer"}
	linkPerformersParams := []db.LinkPerformersParams{
		{EventID: eventID, Name: "Test Performer"},
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("CreateEvent", mock.Anything, createEventParams).Return(eventID, nil)
	mockQueries.On("WritePerformers", mock.Anything, writePerformersParams).Return(
		&db.WritePerformersBatchResults{},
	)
	mockQueries.On("LinkPerformers", mock.Anything, linkPerformersParams).Return(
		&db.LinkPerformersBatchResults{},
	)

	event := entities.Event{
		ID:          eventID,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue:       entities.EventVenue{ID: 1},
		Performers:  []entities.Performer{{Name: "Test Performer"}},
	}

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	actual, err := repo.ExecCreateEvent(
		ctx,
		mockQueries,
		event,
		func(br repos.Closable) error { return nil },
	)

	assert.Equal(t, eventID, actual)
	assert.Nil(t, err)
	mockQueries.AssertCalled(t, "CreateEvent", ctx, createEventParams)
	mockQueries.AssertCalled(t, "WritePerformers", ctx, writePerformersParams)
	mockQueries.AssertCalled(t, "LinkPerformers", ctx, linkPerformersParams)
}

func TestEventsRepoGetEvent(t *testing.T) {
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
			PerformerName: pgtype.Text{String: "Test Performer", Valid: true},
		},
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("GetEvent", mock.Anything, eventID).Return(rows, nil)

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
			{ID: 1, Name: "Test Performer"},
		},
	}

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	actual, err := repo.GetEvent(context.Background(), eventID)

	assert.EqualValues(t, expected, actual)
	assert.Nil(t, err)
}

func TestEventsRepoGetEventWhenNotFoundOrMarkedDeleted(t *testing.T) {
	mockQueries := new(MockQuerier)
	mockQueries.On("GetEvent", mock.Anything, eventID).Return([]db.GetEventRow{}, nil)

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	actual, err := repo.GetEvent(context.Background(), eventID)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, repos.ErrNoSuchEntity)
}

func TestEventsRepoExecUpdateEvent(t *testing.T) {
	ctx := context.Background()
	eventID := int32(1)
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-01")

	updateEventParams := db.UpdateEventParams{
		EventID:     eventID,
		Name:        "Test Event",
		StartsAt:    pgtype.Timestamptz{Time: startsAt, Valid: true},
		EndsAt:      pgtype.Timestamptz{Time: endsAt, Valid: true},
		Description: pgtype.Text{String: "", Valid: false},
	}
	writePerformersParams := []string{"Test Performer"}
	linkPerformersParams := db.LinkUpdatedPerformersParams{
		EventID: eventID,
		Names:   []string{"Test Performer"},
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("UpdateEvent", mock.Anything, updateEventParams).Return(eventID, nil)
	mockQueries.On("WritePerformers", mock.Anything, writePerformersParams).Return(
		&db.WritePerformersBatchResults{},
	)
	mockQueries.On("LinkUpdatedPerformers", mock.Anything, linkPerformersParams).Return(nil)

	event := entities.Event{
		ID:          eventID,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue:       entities.EventVenue{ID: 1},
		Performers:  []entities.Performer{{Name: "Test Performer"}},
	}

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	err := repo.ExecUpdateEvent(
		ctx,
		mockQueries,
		event,
		func(br repos.Closable) error { return nil },
	)

	assert.Nil(t, err)
	mockQueries.AssertCalled(t, "UpdateEvent", ctx, updateEventParams)
	mockQueries.AssertCalled(t, "WritePerformers", ctx, writePerformersParams)
	mockQueries.AssertCalled(t, "LinkUpdatedPerformers", ctx, linkPerformersParams)
}

func TestEventsRepoExecUpdateEventWhenNoPerformers(t *testing.T) {
	ctx := context.Background()
	eventID := int32(1)
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-01")

	updateEventParams := db.UpdateEventParams{
		EventID:     eventID,
		Name:        "Test Event",
		StartsAt:    pgtype.Timestamptz{Time: startsAt, Valid: true},
		EndsAt:      pgtype.Timestamptz{Time: endsAt, Valid: true},
		Description: pgtype.Text{String: "", Valid: false},
	}
	writePerformersParams := []string{}
	linkPerformersParams := db.LinkUpdatedPerformersParams{}

	mockQueries := new(MockQuerier)
	mockQueries.On("UpdateEvent", mock.Anything, updateEventParams).Return(eventID, nil)
	mockQueries.On("TrimUpdatedEventPerformers", mock.Anything, eventID).Return(nil)
	mockQueries.On("WritePerformers", mock.Anything, writePerformersParams).Return(
		&db.WritePerformersBatchResults{},
	)
	mockQueries.On("LinkUpdatedPerformers", mock.Anything, linkPerformersParams).Return(nil)

	event := entities.Event{
		ID:          eventID,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue:       entities.EventVenue{ID: 1},
		Performers:  []entities.Performer{},
	}

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	err := repo.ExecUpdateEvent(
		ctx,
		mockQueries,
		event,
		func(br repos.Closable) error { return nil },
	)

	assert.Nil(t, err)
	mockQueries.AssertCalled(t, "UpdateEvent", ctx, updateEventParams)
	mockQueries.AssertCalled(t, "TrimUpdatedEventPerformers", ctx, eventID)
	mockQueries.AssertNotCalled(t, "WritePerformers")
	mockQueries.AssertNotCalled(t, "LinkUpdatedPerformers")
}

func TestEventsRepoExecUpdateEventWhenDoesntExistOrDeleted(t *testing.T) {
	ctx := context.Background()
	eventID := int32(1)
	startsAt, _ := time.Parse(time.DateOnly, "2020-01-01")
	endsAt, _ := time.Parse(time.DateOnly, "2020-01-01")

	updateEventParams := db.UpdateEventParams{
		EventID:     eventID,
		Name:        "Test Event",
		StartsAt:    pgtype.Timestamptz{Time: startsAt, Valid: true},
		EndsAt:      pgtype.Timestamptz{Time: endsAt, Valid: true},
		Description: pgtype.Text{String: "", Valid: false},
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("UpdateEvent", mock.Anything, updateEventParams).Return(eventID, sql.ErrNoRows)

	event := entities.Event{
		ID:          eventID,
		Name:        "Test Event",
		StartsAt:    startsAt,
		EndsAt:      endsAt,
		Description: "",
		Venue:       entities.EventVenue{ID: 1},
		Performers:  []entities.Performer{{Name: "Test Performer"}},
	}

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	err := repo.ExecUpdateEvent(
		ctx,
		mockQueries,
		event,
		func(br repos.Closable) error { return nil },
	)

	assert.ErrorIs(t, repos.ErrNoSuchEntity, err)
	mockQueries.AssertCalled(t, "UpdateEvent", ctx, updateEventParams)
}

func TestEventsRepoDeleteEvent(t *testing.T) {
	ctx := context.Background()
	mockQueries := new(MockQuerier)
	mockQueries.On("DeleteEvent", ctx, eventID).Return(int64(1), nil)

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	err := repo.DeleteEvent(ctx, eventID)

	assert.Nil(t, err)
	mockQueries.AssertCalled(t, "DeleteEvent", ctx, eventID)
}

func TestEventsRepoDeleteEventWhenDoesntExistOrDeleted(t *testing.T) {
	mockQueries := new(MockQuerier)
	mockQueries.On("DeleteEvent", mock.Anything, eventID).Return(int64(0), nil)

	repo := repos.NewEventsRepoFromQueries(mockQueries)
	err := repo.DeleteEvent(context.Background(), eventID)

	assert.ErrorIs(t, err, repos.ErrNoSuchEntity)
}

func TestTicketsRepoExecWriteTickets(t *testing.T) {
	eventID := int32(1)
	tickets := []entities.Ticket{
		{EventID: eventID, Price: 10, Seat: "GA"},
		{EventID: eventID, Price: 10, Seat: "GA"},
		{EventID: eventID, Price: 20, Seat: "Balcony"},
	}
	params := []db.WriteNewTicketsParams{
		{EventID: eventID, Price: 10, Seat: "GA"},
		{EventID: eventID, Price: 10, Seat: "GA"},
		{EventID: eventID, Price: 20, Seat: "Balcony"},
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("WriteNewTickets", mock.Anything, params).Return(
		&db.WriteNewTicketsBatchResults{},
	)

	repo := repos.NewTicketsRepoFromQueries(mockQueries)
	repo.ExecWriteTickets(
		context.Background(),
		mockQueries,
		tickets,
		func(_ repos.QueryRowable) {},
	)

	mockQueries.AssertCalled(t, "WriteNewTickets", mock.Anything, params)
}

func TestTicketsRepoGetAvailableTickets(t *testing.T) {
	eventID := int32(1)
	ctx := context.Background()
	rows := []db.GetAvailableTicketsRow{
		{Ticket: db.Ticket{ID: 1, EventID: eventID, PurchaserID: pgtype.Int4{Valid: false}, Price: 10, Seat: "GA"}},
		{Ticket: db.Ticket{ID: 2, EventID: eventID, PurchaserID: pgtype.Int4{Valid: false}, Price: 10, Seat: "GA"}},
		{Ticket: db.Ticket{ID: 3, EventID: eventID, PurchaserID: pgtype.Int4{Valid: false}, Price: 20, Seat: "Balcony"}},
	}
	expected := []entities.Ticket{
		{ID: 1, EventID: eventID, IsPurchased: false, Price: 10, Seat: "GA"},
		{ID: 2, EventID: eventID, IsPurchased: false, Price: 10, Seat: "GA"},
		{ID: 3, EventID: eventID, IsPurchased: false, Price: 20, Seat: "Balcony"},
	}

	mockQueries := new(MockQuerier)
	mockQueries.On("GetAvailableTickets", ctx, eventID).Return(rows, nil)

	repo := repos.NewTicketsRepoFromQueries(mockQueries)
	actual, err := repo.GetAvailableTickets(ctx, eventID)

	assert.Nil(t, err)
	assert.ElementsMatch(t, expected, actual)
	mockQueries.AssertCalled(t, "GetAvailableTickets", ctx, eventID)
}

func TestTicketsRepoGetAvailableTicketsWhenEventDoesntExistOrDeleted(t *testing.T) {
	eventID := int32(1)
	ctx := context.Background()
	rows := []db.GetAvailableTicketsRow{}

	mockQueries := new(MockQuerier)
	mockQueries.On("GetAvailableTickets", ctx, eventID).Return(rows, nil)

	repo := repos.NewTicketsRepoFromQueries(mockQueries)
	_, err := repo.GetAvailableTickets(ctx, eventID)

	assert.ErrorIs(t, repos.ErrNoSuchEntity, err)
	mockQueries.AssertCalled(t, "GetAvailableTickets", ctx, eventID)
}
