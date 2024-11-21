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
