package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/danielgtaylor/huma/v2/humatest"
	pkgApi "github.com/dslaw/book-tickets/pkg/api"
	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/dslaw/book-tickets/pkg/services"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/joho/godotenv/autoload"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	readVenueID    = int32(1)
	updateVenueID  = int32(2)
	deletedVenueID = int32(3)
	missingVenueID = int32(999)

	readEventID    = int32(1)
	updateEventID  = int32(2)
	deletedEventID = int32(3)
	missingEventID = int32(999)
)

func ClearTestDatabase(ctx context.Context, conn *pgxpool.Pool) error {
	tableNames := []string{
		"performers",
		"event_performers",
		"events",
		"venues",
	}

	for _, tableName := range tableNames {
		stmt := fmt.Sprintf("delete from %s cascade", tableName)
		_, err := conn.Exec(ctx, stmt)
		if err != nil {
			return err
		}
	}

	return nil
}

func WriteTestData(ctx context.Context, conn *pgxpool.Pool) error {
	insertVenuesStmt := `
insert into venues (id, name, description, address, city, subdivision, country_code, deleted)
overriding system value
values
    ($1, 'Test venue to read', '', '11 Front Street', 'San Francisco', 'CA', 'USA', false),
    ($2, 'Test venue to update', '', '12 Front Street', 'San Francisco', 'CA', 'USA', false),
    ($3, 'Test venue deleted', '', '13 Front Street', 'San Francisco', 'CA', 'USA', true);
`
	_, err := conn.Exec(ctx, insertVenuesStmt, readVenueID, updateVenueID, deletedVenueID)
	if err != nil {
		return err
	}

	insertEventsStmt := `
insert into events (id, venue_id, name, description, starts_at, ends_at, deleted)
overriding system value
values
    ($2, $1, 'Test event to read', '', '2020-01-01:00:00.00Z', '2020-01-01:00:00.00Z', false),
    ($3, $1, 'Test event to update', '', '2020-01-01:00:00.00Z', '2020-01-01:00:00.00Z', false),
    ($4, $1, 'Test event deleted', '', '2020-01-01:00:00.00Z', '2020-01-01:00:00.00Z', true);
`
	_, err = conn.Exec(
		ctx,
		insertEventsStmt,
		readVenueID,
		readEventID,
		updateEventID,
		deletedEventID,
	)

	return err
}

type HandlersTestSuite struct {
	suite.Suite
	Conn *pgxpool.Pool
}

func (suite *HandlersTestSuite) SetupSuite() {
	ctx := context.Background()

	testDatabaseURL := os.Getenv("TEST_DATABASE_URL_LOCAL")
	dbConfig, _ := pgxpool.ParseConfig(testDatabaseURL)
	dbConfig.MaxConns = 1
	conn, err := pgxpool.NewWithConfig(ctx, dbConfig)
	if err != nil {
		assert.FailNow(suite.T(), "Unable to connect to test database")
	}

	// Set up test data.
	if err = ClearTestDatabase(ctx, conn); err != nil {
		assert.FailNow(suite.T(), "Unable to clear test data on setup")
	}
	if err = WriteTestData(ctx, conn); err != nil {
		assert.FailNow(suite.T(), "Unable to write test data on setup")
	}

	suite.Conn = conn
}

func (suite *HandlersTestSuite) TeardownSuite() {
	if err := ClearTestDatabase(context.Background(), suite.Conn); err != nil {
		assert.FailNow(suite.T(), "Unable to clear test data on teardown")
	}

	defer suite.Conn.Close()
}

func CreateAPIForVenues(suite *HandlersTestSuite) humatest.TestAPI {
	t := suite.T()
	service := services.NewVenuesService(repos.NewVenuesRepo(suite.Conn))
	_, api := humatest.New(t)
	pkgApi.RegisterVenuesHandlers(api, service)
	return api
}

func CreateAPIForEvents(suite *HandlersTestSuite) humatest.TestAPI {
	t := suite.T()
	service := services.NewEventsService(repos.NewEventsRepo(suite.Conn))
	_, api := humatest.New(t)
	pkgApi.RegisterEventsHandlers(api, service)
	return api
}

// Test creating a new venue.
func (suite *HandlersTestSuite) TestCreateVenue() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	data := map[string]any{
		"name": "Test creating a new venue",
		"location": map[string]any{
			"address":      "22 Front Street",
			"city":         "San Francisco",
			"subdivision":  "CA",
			"country_code": "USA",
		},
	}

	response := api.Post("/venues", data)
	require.Equal(t, http.StatusOK, response.Code)

	actual := pkgApi.CreateVenueResponse{}
	json.NewDecoder(response.Body).Decode(&actual)
	newVenueID := actual.ID

	require.NotEmpty(t, newVenueID)

	queries := db.New(suite.Conn)
	row, err := queries.GetVenue(context.Background(), newVenueID)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Error reading venue: %s", err))
	}

	assert.EqualValues(t, db.Venue{
		ID:          newVenueID,
		Name:        "Test creating a new venue",
		Description: pgtype.Text{String: "", Valid: false},
		Address:     "22 Front Street",
		City:        "San Francisco",
		Subdivision: "CA",
		CountryCode: "USA",
		Deleted:     false,
	}, row.Venue)
}

// Test creating a new venue with a missing field and incorrectly formatted
// country code.
func (suite *HandlersTestSuite) TestCreateVenueWhenMalformedData() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	data := map[string]any{
		"name": "Test new venue",
		"location": map[string]any{
			"address": "1 Front Street",
			// city field is omitted.
			"subdivision":  "CA",
			"country_code": "US",
		},
	}

	response := api.Post("/venues", data)
	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

// Test reading an existing venue.
func (suite *HandlersTestSuite) TestGetVenue() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	response := api.Get(fmt.Sprintf("/venues/%d", readVenueID))
	require.Equal(t, http.StatusOK, response.Code)

	expected := pkgApi.GetVenueResponse{ID: readVenueID, Name: "Test venue to read"}
	expected.Location.Address = "11 Front Street"
	expected.Location.City = "San Francisco"
	expected.Location.Subdivision = "CA"
	expected.Location.CountryCode = "USA"

	actual := pkgApi.GetVenueResponse{}
	json.NewDecoder(response.Body).Decode(&actual)

	assert.EqualValues(t, expected, actual)
}

// Test reading a venue that doesn't exist or has been marked deleted returns
// not found.
func (suite *HandlersTestSuite) TestGetVenueWhenDoesntExistOrDeleted() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	for _, id := range []int32{missingVenueID, deletedVenueID} {
		path := fmt.Sprintf("/venues/%d", id)
		response := api.Get(path)
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
}

// Test updating an existing venue.
func (suite *HandlersTestSuite) TestUpdateVenue() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	data := map[string]any{
		"name":        "Test venue to update",
		"description": "Update",
		"location": map[string]any{
			"address":      "12 Front Street",
			"city":         "San Francisco",
			"subdivision":  "CA",
			"country_code": "USA",
		},
	}

	response := api.Put(fmt.Sprintf("/venues/%d", updateVenueID), data)
	require.Equal(t, http.StatusNoContent, response.Code)

	queries := db.New(suite.Conn)
	row, err := queries.GetVenue(context.Background(), updateVenueID)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Error reading venue: %s", err))
	}

	assert.EqualValues(t, row.Venue, db.Venue{
		ID:          updateVenueID,
		Name:        "Test venue to update",
		Description: pgtype.Text{String: "Update", Valid: true},
		Address:     "12 Front Street",
		City:        "San Francisco",
		Subdivision: "CA",
		CountryCode: "USA",
		Deleted:     false,
	})
}

// Test that updating a non-existent or deleted venue returns not found.
func (suite *HandlersTestSuite) TestUpdateVenueWhenDoesntExistOrDeleted() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	data := map[string]any{
		"name": "Test venue to update",
		"location": map[string]any{
			"address":      "999 Front Street",
			"city":         "San Francisco",
			"subdivision":  "CA",
			"country_code": "USA",
		},
	}

	for _, id := range []int32{missingVenueID, deletedVenueID} {
		path := fmt.Sprintf("/venues/%d", id)
		response := api.Put(path, data)
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
}

// Test deleting an existing venue.
func (suite *HandlersTestSuite) TestDeleteVenue() {
	toDeleteVenueID := int32(11)
	t := suite.T()

	// Set up a venue to be deleted.
	_, err := suite.Conn.Exec(context.Background(), `
insert into venues (id, name, address, city, subdivision, country_code)
overriding system value
values ($1, 'Test venue to delete', '99 Front Street', 'San Francisco', 'CA', 'USA')
`, toDeleteVenueID)
	if err != nil {
		assert.FailNow(t, "Unable to write test data")
	}

	api := CreateAPIForVenues(suite)

	response := api.Delete(fmt.Sprintf("/venues/%d", toDeleteVenueID))
	require.Equal(t, http.StatusNoContent, response.Code)

	queries := db.New(suite.Conn)
	_, err = queries.GetVenue(context.Background(), toDeleteVenueID)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// Test that deleting a non-existent or deleted venue returns not found.
func (suite *HandlersTestSuite) TestDeleteVenueWhenDoesntExistOrDeleted() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	for _, id := range []int32{missingVenueID, deletedVenueID} {
		path := fmt.Sprintf("/venues/%d", id)
		response := api.Delete(path)
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
}

// Test creating a new event.
func (suite *HandlersTestSuite) TestCreateEvent() {
	t := suite.T()
	ctx := context.Background()

	// Set up a performers record, so we can test the case where the event
	// references an existing performer and a new performer.
	preExistingPerformerName := "Test Performer 1"
	_, err := suite.Conn.Exec(
		ctx,
		"insert into performers (name) values ($1)",
		preExistingPerformerName,
	)
	if err != nil {
		assert.FailNow(t, "Unable to write test data")
	}

	teardown := func() error {
		_, err := suite.Conn.Exec(
			ctx,
			"delete from performers where name = $1",
			preExistingPerformerName,
		)
		if err != nil {
			assert.FailNow(t, "Unable to clean-up test data")
		}
		return nil
	}
	defer teardown()

	api := CreateAPIForEvents(suite)

	startsAt, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	endsAt, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")

	data := map[string]any{
		"name":        "Test new event",
		"venue_id":    readVenueID,
		"description": "Test",
		"starts_at":   "2020-01-01T00:00:00Z",
		"ends_at":     "2020-01-01T08:00:00Z",
		"performers": []map[string]any{
			{"name": preExistingPerformerName}, // Existing performer.
			{"name": "Test Performer 2"},       // New performer.
		},
	}

	response := api.Post("/events", data)
	require.Equal(t, http.StatusOK, response.Code)

	actual := pkgApi.CreateEventResponse{}
	json.NewDecoder(response.Body).Decode(&actual)
	newEventID := actual.ID

	require.NotEmpty(t, newEventID)

	// Check that the database reflects the create operation.
	queries := db.New(suite.Conn)
	rows, err := queries.GetEvent(context.Background(), newEventID)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Error reading event: %s", err))
	}

	require.Equal(t, 2, len(rows)) // One event with two performers yields two returned rows.

	row := rows[0]
	row.Event.StartsAt.Time = row.Event.StartsAt.Time.UTC()
	row.Event.EndsAt.Time = row.Event.EndsAt.Time.UTC()

	assert.EqualValues(t, db.Event{
		ID:          newEventID,
		VenueID:     readVenueID,
		Name:        "Test new event",
		Description: pgtype.Text{String: "Test", Valid: true},
		StartsAt:    pgtype.Timestamptz{Time: startsAt.UTC(), Valid: true},
		EndsAt:      pgtype.Timestamptz{Time: endsAt.UTC(), Valid: true},
	}, row.Event)
	assert.Equal(t, "Test venue to read", row.VenueName)

	assert.Equal(t, []bool{true, true}, []bool{rows[0].PerformerID.Valid, rows[1].PerformerID.Valid})

	actualPerformers := []pgtype.Text{rows[0].PerformerName, rows[1].PerformerName}
	slices.SortFunc(actualPerformers, func(a, b pgtype.Text) int {
		return strings.Compare(a.String, b.String)
	})
	assert.EqualValues(
		t,
		[]pgtype.Text{
			{String: preExistingPerformerName, Valid: true},
			{String: "Test Performer 2", Valid: true},
		},
		actualPerformers,
	)
}

// Test creating a new event with a missing field.
func (suite *HandlersTestSuite) TestCreateEventWhenMalformedData() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	data := map[string]any{
		"name":        "Test new event",
		"venue_id":    1,
		"description": "Test",
		"starts_at":   "2020-01-01T00:00.00Z",
		// ends_at field is omitted.
		"performers": []map[string]any{
			{"name": "Test Performer"},
		},
	}

	response := api.Post("/events", data)
	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

// Test creating a new event with an invalid venue id.
func (suite *HandlersTestSuite) TestCreateEventWhenVenueDoesntExist() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	data := map[string]any{
		"name":        "Test new event",
		"venue_id":    missingVenueID,
		"description": "Test",
		"starts_at":   "2020-01-01T00:00.00Z",
		"ends_at":     "2020-01-01T08:00:00Z",
		"performers":  []map[string]any{},
	}

	response := api.Post("/events", data)
	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

// Test creating a new event when that event already exists.
func (suite *HandlersTestSuite) TestCreateEventWhenEventAlreadyExists() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	data := map[string]any{
		"name":        "Test event to read",
		"venue_id":    readVenueID,
		"description": "Test",
		"starts_at":   "2020-01-01T00:00.00Z",
		"ends_at":     "2020-01-01T08:00:00Z",
		"performers":  []map[string]any{},
	}

	response := api.Post("/events", data)
	require.Equal(t, http.StatusUnprocessableEntity, response.Code)
}

// Test reading an existing event.
func (suite *HandlersTestSuite) TestGetEvent() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	response := api.Get(fmt.Sprintf("/events/%d", readEventID))
	require.Equal(t, http.StatusOK, response.Code)

	startsAt, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	endsAt, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	expected := pkgApi.GetEventResponse{
		ID:          readEventID,
		Name:        "Test event to read",
		Description: "",
		StartsAt:    startsAt.UTC(),
		EndsAt:      endsAt.UTC(),
		Venue:       pkgApi.EventVenueResponse{ID: readVenueID, Name: "Test venue to read"},
		Performers:  []pkgApi.EventPerformerResponse{},
	}

	actual := pkgApi.GetEventResponse{}
	json.NewDecoder(response.Body).Decode(&actual)
	actual.StartsAt = actual.StartsAt.UTC()
	actual.EndsAt = actual.EndsAt.UTC()

	assert.EqualValues(t, expected, actual)
}

// Test reading an event that doesn't exist or has been marked deleted returns
// not found.
func (suite *HandlersTestSuite) TestGetEventWhenDoesntExistOrDeleted() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	for _, id := range []int32{missingEventID, deletedEventID} {
		path := fmt.Sprintf("/events/%d", id)
		response := api.Get(path)
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
}

// Test updating an existing event.
func (suite *HandlersTestSuite) TestUpdateEvent() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	startsAt, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	endsAt, _ := time.Parse(time.RFC3339, "2020-01-01T08:00:00Z")

	data := map[string]any{
		"name":        "Test event to update",
		"venue_id":    readVenueID,
		"description": "Update",
		"starts_at":   "2020-01-01T00:00:00Z",
		"ends_at":     "2020-01-01T08:00:00Z",
		"performers": []map[string]any{
			{"name": "Test Performer 1"},
		},
	}

	response := api.Put(fmt.Sprintf("/events/%d", updateEventID), data)
	require.Equal(t, http.StatusNoContent, response.Code)

	// Check that the database reflects the update operation.
	queries := db.New(suite.Conn)
	rows, err := queries.GetEvent(context.Background(), updateEventID)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("Error reading event: %s", err))
	}

	require.Equal(t, 1, len(rows)) // One event with one performer yields one returned row.

	row := rows[0]
	row.Event.StartsAt.Time = row.Event.StartsAt.Time.UTC()
	row.Event.EndsAt.Time = row.Event.EndsAt.Time.UTC()

	assert.EqualValues(t, db.Event{
		ID:          updateEventID,
		VenueID:     readVenueID,
		Name:        "Test event to update",
		Description: pgtype.Text{String: "Update", Valid: true},
		StartsAt:    pgtype.Timestamptz{Time: startsAt.UTC(), Valid: true},
		EndsAt:      pgtype.Timestamptz{Time: endsAt.UTC(), Valid: true},
	}, row.Event)
	assert.Equal(t, "Test venue to read", row.VenueName)
	assert.Equal(t, true, row.PerformerID.Valid)
	assert.Equal(t, row.PerformerName, pgtype.Text{String: "Test Performer 1", Valid: true})
}

// Test that updating a non-existent or deleted event returns not found.
func (suite *HandlersTestSuite) TestUpdateEventWhenDoesntExistOrDeleted() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	data := map[string]any{
		"name":        "Test update event",
		"venue_id":    readVenueID,
		"description": "Test",
		"starts_at":   "2020-01-01T00:00:00Z",
		"ends_at":     "2020-01-01T08:00:00Z",
		"performers": []map[string]any{
			{"name": "Test Performer 1"},
		},
	}

	for _, id := range []int32{missingEventID, deletedEventID} {
		path := fmt.Sprintf("/events/%d", id)
		response := api.Put(path, data)
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
}

// Test deleting an existing event.
func (suite *HandlersTestSuite) TestDeleteEvent() {
	toDeleteEventID := int32(11)
	t := suite.T()

	// Set up an event to be deleted.
	_, err := suite.Conn.Exec(context.Background(), `
insert into events (id, venue_id, name, starts_at, ends_at)
overriding system value
values ($1, $2, 'Test event to delete', '2020-01-01:00:00.00Z', '2020-01-01:00:00.00Z')
`, toDeleteEventID, readVenueID)
	if err != nil {
		assert.FailNow(t, "Unable to write test data")
	}

	api := CreateAPIForEvents(suite)

	response := api.Delete(fmt.Sprintf("/events/%d", toDeleteEventID))
	require.Equal(t, http.StatusNoContent, response.Code)

	queries := db.New(suite.Conn)
	ret, _ := queries.GetEvent(context.Background(), toDeleteEventID)
	assert.Empty(t, ret)
}

// Test that deleting a non-existent or deleted event returns not found.
func (suite *HandlersTestSuite) TestDeleteEventWhenDoesntExistOrDeleted() {
	t := suite.T()
	api := CreateAPIForEvents(suite)

	for _, id := range []int32{missingEventID, deletedEventID} {
		path := fmt.Sprintf("/events/%d", id)
		response := api.Delete(path)
		assert.Equal(t, http.StatusNotFound, response.Code)
	}
}

func TestHandlersTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping testing in short mode")
	}
	suite.Run(t, new(HandlersTestSuite))
}