package services_test

import (
	"testing"

	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/dslaw/book-tickets/pkg/services"
	"github.com/stretchr/testify/assert"
)

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
	assert.EqualValues(t, expected, actual)
}
