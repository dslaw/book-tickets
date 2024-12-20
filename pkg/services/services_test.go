package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/dslaw/book-tickets/pkg/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCacheClient struct {
	mock.Mock
}

func (mock *MockCacheClient) Close() error {
	args := mock.Called()
	return args.Error(0)
}

func (mock *MockCacheClient) HashExpireAt(ctx context.Context, field string, tm time.Time) error {
	args := mock.Called(ctx, field, tm)
	return args.Error(0)
}

func (mock *MockCacheClient) HashGet(ctx context.Context, field string) (string, error) {
	args := mock.Called(ctx, field)
	return args.String(0), args.Error(1)
}

func (mock *MockCacheClient) HashSet(ctx context.Context, field, value string) error {
	args := mock.Called(ctx, field, value)
	return args.Error(0)
}

func (mock *MockCacheClient) HashMultiGet(ctx context.Context, fields ...string) (map[string]string, error) {
	args := mock.Called(ctx, fields)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (mock *MockCacheClient) MakeField(id int32) string {
	args := mock.Called(id)
	return args.Get(0).(string)
}

type MockTicketsRepo struct {
	mock.Mock
}

func (mock *MockTicketsRepo) GetAvailableTickets(ctx context.Context, id int32) ([]entities.Ticket, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).([]entities.Ticket), args.Error(1)
}

func (mock *MockTicketsRepo) GetTicket(ctx context.Context, id int32) (entities.Ticket, error) {
	args := mock.Called(ctx, id)
	return args.Get(0).(entities.Ticket), args.Error(1)
}

func (mock *MockTicketsRepo) SetTicketPurchaser(ctx context.Context, ticketID int32, purchaserID int32) error {
	args := mock.Called(ctx, ticketID, purchaserID)
	return args.Error(0)
}

func (mock *MockTicketsRepo) WriteTickets(ctx context.Context, tickets []entities.Ticket) error {
	args := mock.Called(ctx, tickets)
	return args.Error(0)
}

type MockTime struct {
	mock.Mock
}

func (mock *MockTime) Now() time.Time {
	args := mock.Called()
	return args.Get(0).(time.Time)
}

func TestTicketsServiceAggregateTickets(t *testing.T) {
	service := &services.TicketsService{}
	tickets := []entities.Ticket{
		{ID: 1, EventID: 1, IsPurchased: false, Price: 10, Seat: "GA"},
		{ID: 2, EventID: 1, IsPurchased: false, Price: 10, Seat: "GA"},
		{ID: 3, EventID: 1, IsPurchased: false, Price: 20, Seat: "Balcony"},
		{ID: 4, EventID: 1, PurchaserID: 1, IsPurchased: true, Price: 10, Seat: "GA"},
	}
	expected := []entities.AvailableTicketAggregate{
		{Price: 10, Seat: "GA", IDs: []int32{1, 2}},
		{Price: 20, Seat: "Balcony", IDs: []int32{3}},
	}

	actual := service.AggregateTickets(tickets)
	assert.ElementsMatch(t, expected, actual)
}

func TestTicketsServiceSetTicketHold(t *testing.T) {
	ticketHoldDuration, _ := time.ParseDuration("1m")
	ticketID := int32(1)
	field := "1"
	holdID := "123"
	now := time.Now()
	expiresAt := now.UTC().Add(ticketHoldDuration)

	mockRepo := new(MockTicketsRepo)
	mockRepo.On("GetTicket", mock.Anything, ticketID).Return(entities.Ticket{}, nil)

	mockClient := new(MockCacheClient)
	mockClient.On("MakeField", ticketID).Return(field)
	mockClient.On("HashSet", mock.Anything, field, holdID).Return(nil)
	mockClient.On("HashExpireAt", mock.Anything, field, expiresAt).Return(nil)

	mockTime := new(MockTime)
	mockTime.On("Now").Return(now)

	service := services.NewTicketsService(mockRepo, mockClient, mockTime, ticketHoldDuration)
	err := service.SetTicketHold(context.Background(), ticketID, holdID)

	assert.Nil(t, err)
	mockClient.AssertCalled(t, "MakeField", ticketID)
	mockClient.AssertCalled(t, "HashSet", mock.Anything, field, holdID)
	mockClient.AssertCalled(t, "HashExpireAt", mock.Anything, field, expiresAt)
	mockTime.AssertCalled(t, "Now")
}

func TestTicketsServiceSetTicketHoldWhenTicketDoesntExist(t *testing.T) {
	ticketHoldDuration, _ := time.ParseDuration("1m")
	ticketID := int32(1)
	holdID := "123"

	mockRepo := new(MockTicketsRepo)
	mockRepo.On("GetTicket", mock.Anything, ticketID).Return(
		entities.Ticket{},
		repos.ErrNoSuchEntity,
	)

	service := services.NewTicketsService(mockRepo, nil, nil, ticketHoldDuration)
	err := service.SetTicketHold(context.Background(), ticketID, holdID)

	assert.ErrorIs(t, repos.ErrNoSuchEntity, err)
}

func TestTicketsServiceGetHeldTicketWhenHoldIDMismatch(t *testing.T) {
	ticketHoldDuration, _ := time.ParseDuration("1m")
	ticketID := int32(1)
	field := "1"
	holdID := "222"
	actualHoldID := "111"

	mockClient := new(MockCacheClient)
	mockClient.On("MakeField", ticketID).Return(field)
	mockClient.On("HashGet", mock.Anything, field).Return(actualHoldID, nil)

	service := services.NewTicketsService(nil, mockClient, nil, ticketHoldDuration)
	ticket, err := service.GetHeldTicket(context.Background(), ticketID, holdID)

	assert.Empty(t, ticket)
	assert.ErrorIs(t, services.ErrHoldIDMismatch, err)
}
