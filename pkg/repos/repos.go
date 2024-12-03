package repos

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Closable interface {
	Close() error
}

// closeBatch implements `Closable` for closing the result of a batch query.
func closeBatch(br Closable) error {
	return br.Close()
}

type QueryRowable interface {
	QueryRow(func(int, int32, error))
}

type VenuesRepo struct {
	queries db.Querier
}

func NewVenuesRepo(conn db.DBTX) *VenuesRepo {
	return &VenuesRepo{queries: db.New(conn)}
}

// For creating a repo with a mock queries object when testing.
func NewVenuesRepoFromQueries(queries db.Querier) *VenuesRepo {
	return &VenuesRepo{queries: queries}
}

// CreateVenue inserts a new venue into the database of record and returns its
// id, if successful.
func (r *VenuesRepo) CreateVenue(ctx context.Context, venue entities.Venue) (int32, error) {
	params := db.CreateVenueParams{
		Name:        venue.Name,
		Description: MapNullableString(venue.Description),
		Address:     venue.Location.Address,
		City:        venue.Location.City,
		Subdivision: venue.Location.Subdivision,
		CountryCode: venue.Location.CountryCode,
	}
	id, err := r.queries.CreateVenue(ctx, params)
	// TODO: Catch and remap unique constraint error
	return id, err
}

// GetVenue fetches the venue, given by id, from the database of record.
func (r *VenuesRepo) GetVenue(ctx context.Context, id int32) (venue entities.Venue, err error) {
	row, err := r.queries.GetVenue(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return venue, ErrNoSuchEntity
		}
		return venue, err
	}

	venue.ID = row.Venue.ID
	venue.Name = row.Venue.Name
	venue.Description = row.Venue.Description.String
	venue.Location.Address = row.Venue.Address
	venue.Location.City = row.Venue.City
	venue.Location.Subdivision = row.Venue.Subdivision
	venue.Location.CountryCode = row.Venue.CountryCode
	return venue, nil
}

// UpdateVenue updates an existing venue in the database of record.
func (r *VenuesRepo) UpdateVenue(ctx context.Context, venue entities.Venue) error {
	params := db.UpdateVenueParams{
		Name:        venue.Name,
		Description: MapNullableString(venue.Description),
		Address:     venue.Location.Address,
		City:        venue.Location.City,
		Subdivision: venue.Location.Subdivision,
		CountryCode: venue.Location.CountryCode,
		VenueID:     venue.ID,
	}

	if _, err := r.queries.UpdateVenue(ctx, params); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoSuchEntity
		}
		return err
	}

	return nil
}

// DeleteVenue marks a venue and all associated events as deleted in the
// database of record.
func (r *VenuesRepo) DeleteVenue(ctx context.Context, id int32) error {
	countDeleted, err := r.queries.DeleteVenue(ctx, id)
	if err != nil {
		return err
	}
	if countDeleted == 0 {
		return ErrNoSuchEntity
	}
	return nil
}

type EventsRepo struct {
	Conn    *pgxpool.Pool
	queries db.Querier
}

func NewEventsRepo(conn *pgxpool.Pool) *EventsRepo {
	return &EventsRepo{Conn: conn, queries: db.New(conn)}
}

// For creating a repo with a mock queries object when testing.
func NewEventsRepoFromQueries(queries db.Querier) *EventsRepo {
	return &EventsRepo{Conn: nil, queries: queries}
}

func (r *EventsRepo) writePerformers(
	ctx context.Context,
	queries db.Querier,
	performers []entities.Performer,
	closeBatch func(Closable) error,
) ([]string, error) {
	performerNames := make([]string, len(performers))

	if len(performerNames) == 0 {
		return performerNames, nil
	}

	for idx, performer := range performers {
		performerNames[idx] = performer.Name
	}

	br := queries.WritePerformers(ctx, performerNames)
	return performerNames, closeBatch(br)
}

func (r *EventsRepo) ExecCreateEvent(
	ctx context.Context,
	queries db.Querier,
	event entities.Event,
	// Callback to close a batch results object. This allows for ease of
	// testing, as the BatchResults object returned by a batch query doesn't
	// have an interface to mock, and its call to `Close()` forwards the call to
	// a private object.
	closeBatch func(Closable) error,
) (int32, error) {
	// Insert event.
	params := db.CreateEventParams{
		VenueID:     event.Venue.ID,
		Name:        event.Name,
		StartsAt:    MapTime(event.StartsAt),
		EndsAt:      MapTime(event.EndsAt),
		Description: MapNullableString(event.Description),
	}
	id, err := queries.CreateEvent(ctx, params)
	if err != nil {
		return id, err
	}

	// Upsert performers.
	performerNames, err := r.writePerformers(ctx, queries, event.Performers, closeBatch)

	// Add event<->performer associations to the bridge table.
	bridgeParams := make([]db.LinkPerformersParams, len(performerNames))
	for idx, performerName := range performerNames {
		bridgeParams[idx] = db.LinkPerformersParams{EventID: id, Name: performerName}
	}

	lbr := queries.LinkPerformers(ctx, bridgeParams)
	return id, closeBatch(lbr)
}

// CreateEvent inserts a new event into the database of record, and creates new
// performers as necessary. The new event's id is returned, if successful.
func (r *EventsRepo) CreateEvent(ctx context.Context, event entities.Event) (int32, error) {
	var id int32

	tx, err := r.Conn.Begin(ctx)
	if err != nil {
		return id, err
	}
	defer tx.Rollback(ctx)

	qtx := db.New(tx)
	id, err = r.ExecCreateEvent(ctx, qtx, event, closeBatch)
	if err != nil {
		return id, err
	}

	err = tx.Commit(ctx)
	return id, err
}

// GetEvent fetches the venue, given by id, from the database of record.
func (r *EventsRepo) GetEvent(ctx context.Context, id int32) (entities.Event, error) {
	rows, err := r.queries.GetEvent(ctx, id)
	if err != nil {
		return entities.Event{}, err
	}
	if len(rows) == 0 {
		return entities.Event{}, ErrNoSuchEntity
	}

	return MapGetEventRows(rows), nil
}

// UpdateEvent updates an existing venue in the database of record.
func (r *EventsRepo) ExecUpdateEvent(
	ctx context.Context,
	queries db.Querier,
	event entities.Event,
	// Callback to close a batch results object. This allows for ease of
	// testing, as the BatchResults object returned by a batch query doesn't
	// have an interface to mock, and its call to `Close()` forwards the call to
	// a private object.
	closeBatch func(Closable) error,
) error {
	params := db.UpdateEventParams{
		EventID:     event.ID,
		Name:        event.Name,
		StartsAt:    MapTime(event.StartsAt),
		EndsAt:      MapTime(event.EndsAt),
		Description: MapNullableString(event.Description),
	}

	if _, err := queries.UpdateEvent(ctx, params); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoSuchEntity
		}
		return err
	}

	if len(event.Performers) == 0 {
		// Remove event<->performer assocations, leaving any dangling performer
		// records intact.
		return queries.TrimUpdatedEventPerformers(ctx, event.ID)
	}

	// Add performer records as necessary, and update the set of
	// event<->associations.
	performerNames, err := r.writePerformers(ctx, queries, event.Performers, closeBatch)
	if err != nil {
		return err
	}

	bridgeParams := db.LinkUpdatedPerformersParams{
		EventID: event.ID,
		Names:   performerNames,
	}
	return queries.LinkUpdatedPerformers(ctx, bridgeParams)
}

// UpdateEvent updates an existing venue in the database of record.
func (r *EventsRepo) UpdateEvent(ctx context.Context, event entities.Event) error {
	tx, err := r.Conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := db.New(tx)
	if err := r.ExecUpdateEvent(ctx, qtx, event, closeBatch); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// DeleteEvent marks an event as deleted in the database of record.
func (r *EventsRepo) DeleteEvent(ctx context.Context, id int32) error {
	countDeleted, err := r.queries.DeleteEvent(ctx, id)
	if err != nil {
		return err
	}
	if countDeleted == 0 {
		return ErrNoSuchEntity
	}
	return nil
}

type TicketsRepo struct {
	queries db.Querier
}

func NewTicketsRepo(conn db.DBTX) *TicketsRepo {
	return &TicketsRepo{queries: db.New(conn)}
}

// For creating a repo with a mock queries object when testing.
func NewTicketsRepoFromQueries(queries db.Querier) *TicketsRepo {
	return &TicketsRepo{queries: queries}
}

func (r *TicketsRepo) ExecWriteTickets(
	ctx context.Context,
	queries db.Querier,
	tickets []entities.Ticket,
	queryRow func(QueryRowable),
) {
	params := make([]db.WriteNewTicketsParams, len(tickets))
	for idx, ticket := range tickets {
		params[idx] = db.WriteNewTicketsParams{
			EventID: ticket.EventID,
			Price:   int32(ticket.Price),
			Seat:    ticket.Seat,
		}
	}

	br := queries.WriteNewTickets(ctx, params)
	queryRow(br)
	return
}

func (r *TicketsRepo) WriteTickets(ctx context.Context, tickets []entities.Ticket) error {
	var err error

	if len(tickets) == 0 {
		return nil
	}

	collectErr := func(_ int, _ int32, batchErr error) {
		err = batchErr
	}

	r.ExecWriteTickets(ctx, r.queries, tickets, func(br QueryRowable) {
		br.QueryRow(collectErr)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNoSuchEntity
		}
		return err
	}

	return nil
}

// GetAvailableTickets fetches tickets that are available for purchase, for the
// given event.
func (r *TicketsRepo) GetAvailableTickets(ctx context.Context, eventID int32) ([]entities.Ticket, error) {
	rows, err := r.queries.GetAvailableTickets(ctx, eventID)
	if err != nil {
		return []entities.Ticket{}, err
	}
	if len(rows) == 0 {
		return []entities.Ticket{}, ErrNoSuchEntity
	}

	return MapGetAvailableTicketRows(rows), nil
}
