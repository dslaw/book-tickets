package api_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/danielgtaylor/huma/v2/humatest"
	pkgApi "github.com/dslaw/book-tickets/pkg/api"
	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/dslaw/book-tickets/pkg/services"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
)

func ClearTestDatabase(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, "delete from venues")
	return err
}

func WriteTestData(ctx context.Context, conn *pgx.Conn) error {
	insertVenuesStmt := `
insert into venues (id, name, description, address, city, subdivision, country_code, deleted)
overriding system value
values
    ($1, 'Test venue to read', '', '11 Front Street', 'San Francisco', 'CA', 'USA', false),
    ($2, 'Test venue to update', '', '12 Front Street', 'San Francisco', 'CA', 'USA', false),
    ($3, 'Test venue deleted', '', '13 Front Street', 'San Francisco', 'CA', 'USA', true);
`
	_, err := conn.Exec(ctx, insertVenuesStmt, readVenueID, updateVenueID, deletedVenueID)
	return err
}

type HandlersTestSuite struct {
	suite.Suite
	Conn *pgx.Conn
}

func (suite *HandlersTestSuite) SetupSuite() {
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL_LOCAL")
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testDatabaseURL)
	if err != nil {
		assert.FailNow(suite.T(), "Unable to connect to test database")
	}

	// Set up test data.
	if err = ClearTestDatabase(ctx, conn); err != nil {
		assert.FailNow(suite.T(), "Unable to clear test data")
	}
	if err = WriteTestData(ctx, conn); err != nil {
		assert.FailNow(suite.T(), "Unable to write test data")
	}

	suite.Conn = conn
}

func (suite *HandlersTestSuite) TeardownSuite() {
	if err := ClearTestDatabase(context.Background(), suite.Conn); err != nil {
		assert.FailNow(suite.T(), "Unable to clear test data")
	}

	defer suite.Conn.Close(context.Background())
}

func CreateAPIForVenues(suite *HandlersTestSuite) humatest.TestAPI {
	t := suite.T()
	service := services.NewVenuesService(repos.NewVenuesRepo(suite.Conn))
	_, api := humatest.New(t)
	pkgApi.RegisterVenuesHandlers(api, service)
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

	assert.EqualValues(t, row.Venue, db.Venue{
		ID:          newVenueID,
		Name:        "Test creating a new venue",
		Description: pgtype.Text{String: "", Valid: false},
		Address:     "22 Front Street",
		City:        "San Francisco",
		Subdivision: "CA",
		CountryCode: "USA",
		Deleted:     false,
	})
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

// Test reading a venue that doesn't exist.
func (suite *HandlersTestSuite) TestGetVenueWhenDoesntExist() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	response := api.Get(fmt.Sprintf("/venues/%d", missingVenueID))
	require.Equal(t, http.StatusNotFound, response.Code)
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

// Test that updating a deleted venue does not update and returns not found.
func (suite *HandlersTestSuite) TestUpdateVenueWhenDeleted() {
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

	response := api.Put(fmt.Sprintf("/venues/%d", deletedVenueID), data)
	require.Equal(t, http.StatusNotFound, response.Code)
}

// Test that updating a non-existent venue returns not found.
func (suite *HandlersTestSuite) TestUpdateVenueWhenDoesntExist() {
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

	response := api.Put(fmt.Sprintf("/venues/%d", missingVenueID), data)
	require.Equal(t, http.StatusNotFound, response.Code)
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

// Test that deleting a deleted venue returns not found.
func (suite *HandlersTestSuite) TestDeleteVenueWhenDeleted() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	response := api.Delete(fmt.Sprintf("/venues/%d", deletedVenueID))
	require.Equal(t, http.StatusNotFound, response.Code)
}

// Test that deleting a non-existent venue returns not found.
func (suite *HandlersTestSuite) TestDeleteVenueWhenDoesntExist() {
	t := suite.T()
	api := CreateAPIForVenues(suite)

	response := api.Delete(fmt.Sprintf("/venues/%d", missingVenueID))
	require.Equal(t, http.StatusNotFound, response.Code)
}

func TestHandlersTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping testing in short mode")
	}
	suite.Run(t, new(HandlersTestSuite))
}
