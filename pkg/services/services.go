package services

import (
	"context"

	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/dslaw/book-tickets/pkg/repos"
)

type VenuesService struct {
	repo *repos.VenuesRepo
}

func NewVenuesService(repo *repos.VenuesRepo) *VenuesService {
	return &VenuesService{repo: repo}
}

// CreateVenue creates a new venue and returns the new entity's id.
func (svc *VenuesService) CreateVenue(ctx context.Context, venue entities.Venue) (int32, error) {
	return svc.repo.CreateVenue(ctx, venue)
}

// GetVenue fetches a venue given by the id.
func (svc *VenuesService) GetVenue(ctx context.Context, id int32) (entities.Venue, error) {
	return svc.repo.GetVenue(ctx, id)
}

// UpdateVenue updates a venue given by the id.
func (svc *VenuesService) UpdateVenue(ctx context.Context, venue entities.Venue) error {
	return svc.repo.UpdateVenue(ctx, venue)
}

// DeleteVenue deletes a venue given by the id. The ids of affected events are
// returned.
func (svc *VenuesService) DeleteVenue(ctx context.Context, id int32) error {
	return svc.repo.DeleteVenue(ctx, id)
}

type EventsService struct {
	repo *repos.EventsRepo
}

func NewEventsService(repo *repos.EventsRepo) *EventsService {
	return &EventsService{repo: repo}
}

// CreateEvent creates a new event and returns the new entity's id.
func (svc *EventsService) CreateEvent(ctx context.Context, event entities.Event) (int32, error) {
	return svc.repo.CreateEvent(ctx, event)
}

// GetEvent fetches an event given by the id.
func (svc *EventsService) GetEvent(ctx context.Context, id int32) (entities.Event, error) {
	return svc.repo.GetEvent(ctx, id)
}

// UpdateEvent updates an event given by the id.
func (svc *EventsService) UpdateEvent(ctx context.Context, event entities.Event) error {
	return svc.repo.UpdateEvent(ctx, event)
}

// DeleteEvent deletes an event given by the id.
func (svc *EventsService) DeleteEvent(ctx context.Context, id int32) error {
	return svc.repo.DeleteEvent(ctx, id)
}

type TicketsService struct {
	repo *repos.TicketsRepo
}

func NewTicketsService(repo *repos.TicketsRepo) *TicketsService {
	return &TicketsService{repo: repo}
}

func (svc *TicketsService) AddTickets(
	ctx context.Context,
	eventID int32,
	tickets []entities.Ticket,
) error {
	return svc.repo.WriteTickets(ctx, tickets)
}

func (svc *TicketsService) AggregateTickets(tickets []entities.Ticket) []entities.AvailableTicketAggregate {
	grouped := make(map[string][]entities.Ticket)
	for _, ticket := range tickets {
		if ticket.IsPurchased {
			continue
		}

		group, ok := grouped[ticket.Seat]
		if !ok {
			grouped[ticket.Seat] = make([]entities.Ticket, 0)
		}
		grouped[ticket.Seat] = append(group, ticket)
	}

	aggregates := make([]entities.AvailableTicketAggregate, len(grouped))
	idx := 0
	for _, group := range grouped {
		if len(group) == 0 {
			// Shouldn't happen.
			continue
		}

		ids := make([]int32, len(group))
		for ticketIdx, ticket := range group {
			ids[ticketIdx] = ticket.ID
		}

		ticket := group[0]
		aggregates[idx] = entities.AvailableTicketAggregate{
			Price: ticket.Price,
			Seat:  ticket.Seat,
			IDs:   ids,
		}
		idx++
	}

	return aggregates
}

func (svc *TicketsService) GetAvailableTickets(ctx context.Context, eventID int32) ([]entities.AvailableTicketAggregate, error) {
	tickets, err := svc.repo.GetAvailableTickets(ctx, eventID)
	if err != nil {
		return []entities.AvailableTicketAggregate{}, err
	}
	return svc.AggregateTickets(tickets), nil
}
