package repos_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQuerier struct {
	mock.Mock
}

func (mock *MockQuerier) CreateVenue(ctx context.Context, params db.CreateVenueParams) (int32, error) {
	args := mock.Called(ctx, params)
	return args.Get(0).(int32), args.Error(1)
}

func (mock *MockQuerier) GetVenue(ctx context.Context, venueID int32) (db.GetVenueRow, error) {
	args := mock.Called(ctx, venueID)
	return args.Get(0).(db.GetVenueRow), args.Error(1)
}

func (mock *MockQuerier) UpdateVenue(ctx context.Context, params db.UpdateVenueParams) (int32, error) {
	args := mock.Called(ctx, params)
	return args.Get(0).(int32), args.Error(1)
}

func (mock *MockQuerier) DeleteVenue(ctx context.Context, id int32) (int64, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

const venueID = int32(1)

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

func TestVenuesRepoGetVenueWhenNotFound(t *testing.T) {
	mockQueries := new(MockQuerier)
	mockQueries.On("GetVenue", mock.Anything, venueID).Return(db.GetVenueRow{}, sql.ErrNoRows)

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	actual, err := repo.GetVenue(context.Background(), venueID)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, repos.ErrNoSuchEntity)
}

func TestVenuesRepoGetVenueWhenDeleted(t *testing.T) {
	ret := db.GetVenueRow{Venue: db.Venue{Deleted: true}}

	mockQueries := new(MockQuerier)
	mockQueries.On("GetVenue", mock.Anything, venueID).Return(ret, nil)

	repo := repos.NewVenuesRepoFromQueries(mockQueries)
	actual, err := repo.GetVenue(context.Background(), venueID)

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, repos.ErrEntityDeleted)
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
	mockQueries.On("DeleteVenue", ctx, venueID).Return(int64(venueID), nil)

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
