package repos_test

import (
	"context"

	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/stretchr/testify/mock"
)

type MockQuerier struct {
	mock.Mock
}

func (mock *MockQuerier) CreateEvent(ctx context.Context, params db.CreateEventParams) (int32, error) {
	args := mock.Called(ctx, params)
	return args.Get(0).(int32), args.Error(1)
}

func (mock *MockQuerier) CreateVenue(ctx context.Context, params db.CreateVenueParams) (int32, error) {
	args := mock.Called(ctx, params)
	return args.Get(0).(int32), args.Error(1)
}

func (mock *MockQuerier) DeleteEvent(ctx context.Context, id int32) (int64, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (mock *MockQuerier) DeleteVenue(ctx context.Context, id int32) (int64, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).(int64), args.Error(1)
}

func (mock *MockQuerier) GetAvailableTickets(ctx context.Context, eventID int32) ([]db.GetAvailableTicketsRow, error) {
	args := mock.Called(ctx, eventID)
	return args.Get(0).([]db.GetAvailableTicketsRow), args.Error(1)
}

func (mock *MockQuerier) GetEvent(ctx context.Context, id int32) ([]db.GetEventRow, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).([]db.GetEventRow), args.Error(1)
}

func (mock *MockQuerier) GetTicket(ctx context.Context, id int32) (db.GetTicketRow, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).(db.GetTicketRow), args.Error(1)
}

func (mock *MockQuerier) GetVenue(ctx context.Context, venueID int32) (db.GetVenueRow, error) {
	args := mock.Called(ctx, venueID)
	return args.Get(0).(db.GetVenueRow), args.Error(1)
}

func (mock *MockQuerier) LinkPerformers(ctx context.Context, params []db.LinkPerformersParams) *db.LinkPerformersBatchResults {
	args := mock.Called(ctx, params)
	return args.Get(0).(*db.LinkPerformersBatchResults)
}

func (mock *MockQuerier) LinkUpdatedPerformers(ctx context.Context, params db.LinkUpdatedPerformersParams) error {
	args := mock.Called(ctx, params)
	return args.Error(0)
}

func (mock *MockQuerier) SetTicketPurchaser(ctx context.Context, params db.SetTicketPurchaserParams) (int32, error) {
	args := mock.Called(ctx, params)
	return args.Get(0).(int32), args.Error(1)
}

func (mock *MockQuerier) TrimUpdatedEventPerformers(ctx context.Context, id int32) error {
	args := mock.Called(ctx, id)
	return args.Error(0)
}

func (mock *MockQuerier) UpdateEvent(ctx context.Context, params db.UpdateEventParams) (int32, error) {
	args := mock.Called(ctx, params)
	return args.Get(0).(int32), args.Error(1)
}

func (mock *MockQuerier) UpdateVenue(ctx context.Context, params db.UpdateVenueParams) (int32, error) {
	args := mock.Called(ctx, params)
	return args.Get(0).(int32), args.Error(1)
}

func (mock *MockQuerier) WriteNewTickets(ctx context.Context, params []db.WriteNewTicketsParams) *db.WriteNewTicketsBatchResults {
	args := mock.Called(ctx, params)
	return args.Get(0).(*db.WriteNewTicketsBatchResults)
}

func (mock *MockQuerier) WritePerformers(ctx context.Context, params []string) *db.WritePerformersBatchResults {
	args := mock.Called(ctx, params)
	return args.Get(0).(*db.WritePerformersBatchResults)
}
