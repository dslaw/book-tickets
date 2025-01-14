package services

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/dslaw/book-tickets/pkg/cache"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/dslaw/book-tickets/pkg/search"
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

// TicketsRepoer provides necessary methods for database operations against
// tickets.
type TicketsRepoer interface {
	GetAvailableTickets(context.Context, int32) ([]entities.Ticket, error)
	GetTicket(context.Context, int32) (entities.Ticket, error)
	SetTicketPurchaser(context.Context, int32, int32) error
	WriteTickets(context.Context, []entities.Ticket) error
}

type TicketsService struct {
	repo               TicketsRepoer
	ticketHoldClient   cache.CacheClienter
	TicketHoldDuration time.Duration
}

func NewTicketsService(
	repo TicketsRepoer,
	ticketHoldClient cache.CacheClienter,
	ticketHoldDuration time.Duration,
) *TicketsService {
	return &TicketsService{
		repo:               repo,
		ticketHoldClient:   ticketHoldClient,
		TicketHoldDuration: ticketHoldDuration,
	}
}

// AddTickets creates new tickets for the given event.
func (svc *TicketsService) AddTickets(
	ctx context.Context,
	eventID int32,
	tickets []entities.Ticket,
) error {
	return svc.repo.WriteTickets(ctx, tickets)
}

// AggregateTickets groups tickets for an event by seat.
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

// GetAvailableTickets fetches available (not purchased) tickets for the vent
// given by the event id.
func (svc *TicketsService) GetAvailableTickets(ctx context.Context, eventID int32) ([]entities.AvailableTicketAggregate, error) {
	tickets, err := svc.repo.GetAvailableTickets(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Check for tickets that have a purchase hold on them and filter them out.
	cacheKeys := make([]string, len(tickets))
	for idx, ticket := range tickets {
		cacheKeys[idx] = svc.ticketHoldClient.MakeKey(ticket.ID)
	}

	ticketHolds, err := svc.ticketHoldClient.GetMany(ctx, cacheKeys...)
	if err != nil {
		return nil, err
	}

	tickets = slices.DeleteFunc(tickets, func(ticket entities.Ticket) bool {
		key := svc.ticketHoldClient.MakeKey(ticket.ID)
		_, hasHold := ticketHolds[key]
		return hasHold
	})

	return svc.AggregateTickets(tickets), nil
}

// SetTicketHold places a time-bounded purchase hold on the ticket given by the
// ticket id.
func (svc *TicketsService) SetTicketHold(ctx context.Context, ticketID int32, holdID string) error {
	if holdID == "" {
		return ErrInvalidHoldID
	}

	// Check that the ticket exists, with lack of an error indicating that it
	// exists.
	_, err := svc.repo.GetTicket(ctx, ticketID)
	if err != nil {
		return err
	}

	key := svc.ticketHoldClient.MakeKey(ticketID)
	return svc.ticketHoldClient.Set(ctx, key, holdID, svc.TicketHoldDuration)
}

// GetHeldTicket fetches the ticket given by `ticketID`, if it is currently held
// and the given hold id matches the current purchase hold.
func (svc *TicketsService) GetHeldTicket(ctx context.Context, ticketID int32, holdID string) (ticket entities.Ticket, err error) {
	if holdID == "" {
		err = ErrInvalidHoldID
		return
	}

	// Check that the ticket is held, and that the hold id matches the given
	// hold id, before fetching the ticket.
	key := svc.ticketHoldClient.MakeKey(ticketID)
	actualHoldID, err := svc.ticketHoldClient.Get(ctx, key)
	if err != nil {
		return
	}
	if actualHoldID != holdID {
		err = ErrHoldIDMismatch
		return
	}

	ticket, err = svc.repo.GetTicket(ctx, ticketID)
	return
}

func (svc *TicketsService) SetTicketPurchaser(ctx context.Context, ticketID int32, purchaserID int32) error {
	return svc.repo.SetTicketPurchaser(ctx, ticketID, purchaserID)
}

type SearchService struct {
	client     search.SearchClienter
	MaxResults int32
}

func NewSearchService(client search.SearchClienter, maxResults int32) (*SearchService, error) {
	if maxResults <= 0 {
		return nil, errors.New("`maxResults` must be positive")
	}
	return &SearchService{client: client, MaxResults: maxResults}, nil
}

func (svc *SearchService) SearchEvents(
	ctx context.Context,
	searchTerm string,
	startTime time.Time,
	limit int32,
) ([]search.EventDocument, error) {
	if limit > svc.MaxResults {
		limit = svc.MaxResults
	}

	return svc.client.SearchEvents(ctx, searchTerm, startTime, limit)
}

func (svc *SearchService) SearchVenues(
	ctx context.Context,
	searchTerm string,
	limit int32,
) ([]search.VenueDocument, error) {
	if limit > svc.MaxResults {
		limit = svc.MaxResults
	}

	return svc.client.SearchVenues(ctx, searchTerm, limit)
}
