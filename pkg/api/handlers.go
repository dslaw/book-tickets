package api

// TODO: Log as info level when 404ing?
import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dslaw/book-tickets/pkg/repos"
	"github.com/dslaw/book-tickets/pkg/services"
)

func RegisterVenuesHandlers(api huma.API, service *services.VenuesService) {
	// Create a new venue.
	huma.Post(api, "/venues", func(ctx context.Context, input *struct {
		Body WriteVenueRequest
	}) (*CreateVenueResponseEnvelope, error) {
		venue := MapToVenue(input.Body)
		id, err := service.CreateVenue(ctx, venue)
		if err != nil {
			slog.Error("Issue creating venue", "request_data", input.Body, "error", err)
			return nil, huma.Error500InternalServerError("")
		}

		response := &CreateVenueResponseEnvelope{Body: CreateVenueResponse{ID: id}}
		return response, nil
	})

	// Read an existing venue by id.
	huma.Get(api, "/venues/{id}", func(ctx context.Context, input *struct {
		ID int32 `path:"id"`
	}) (*GetVenueResponseEnvelope, error) {
		venue, err := service.GetVenue(ctx, input.ID)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				return nil, huma.Error404NotFound("")
			}

			slog.Error("Issue fetching venue", "venue_id", input.ID, "error", err)
			return nil, huma.Error500InternalServerError("")
		}

		response := &GetVenueResponseEnvelope{Body: MapToVenueResponse(venue)}
		return response, nil
	})

	// Update an existing venue.
	huma.Put(api, "/venues/{id}", func(ctx context.Context, input *struct {
		ID   int32 `path:"id"`
		Body WriteVenueRequest
	}) (*struct{}, error) {
		venue := MapToVenue(input.Body)
		venue.ID = input.ID
		err := service.UpdateVenue(ctx, venue)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				slog.Error(
					"Attempt to update a non-existent or deleted venue",
					"venue_id", input.ID,
					"request_data", input.Body,
					"error", err,
				)
				return nil, huma.Error404NotFound("")
			}

			slog.Error("Issue updating venue", "venue_id", input.ID, "request_data", input.Body, "error", err)
			return nil, huma.Error500InternalServerError("")
		}
		return nil, nil
	})

	// Delete an existing venue, and all associated events.
	huma.Delete(api, "/venues/{id}", func(ctx context.Context, input *struct {
		ID int32 `path:"id"`
	}) (*struct{}, error) {
		err := service.DeleteVenue(ctx, input.ID)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				return nil, huma.Error404NotFound("")
			}

			slog.Error("Issue deleting venue", "venue_id", input.ID, "error", err)
			return nil, huma.Error500InternalServerError("")
		}
		return nil, nil
	})
}
