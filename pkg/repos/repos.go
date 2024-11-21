package repos

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/jackc/pgx/v5/pgtype"
)

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
// id if successful.
func (r *VenuesRepo) CreateVenue(ctx context.Context, venue entities.Venue) (int32, error) {
	params := db.CreateVenueParams{
		Name:        venue.Name,
		Description: pgtype.Text{String: venue.Description, Valid: venue.Description != ""},
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

	if row.Venue.Deleted {
		return venue, ErrEntityDeleted
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
		Description: pgtype.Text{String: venue.Description, Valid: venue.Description != ""},
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
	slog.Info("Deleted venue", "id", id, "count", countDeleted, "has_error", err != nil)
	if err != nil {
		return err
	}
	if countDeleted == 0 {
		return ErrNoSuchEntity
	}
	return nil
}
