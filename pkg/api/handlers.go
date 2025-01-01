package api

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/danielgtaylor/huma/v2"
	"github.com/dslaw/book-tickets/pkg/cache"
	"github.com/dslaw/book-tickets/pkg/payment"
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

func RegisterEventsHandlers(api huma.API, service *services.EventsService) {
	// Create a new event.
	huma.Post(api, "/events", func(ctx context.Context, input *struct {
		Body WriteEventRequest
	}) (*CreateEventResponseEnvelope, error) {
		event := MapToEvent(input.Body)
		if !event.IsValid() {
			return nil, huma.Error422UnprocessableEntity("")
		}

		id, err := service.CreateEvent(ctx, event)
		if err != nil {
			slog.Error("Issue creating event", "request_data", input.Body, "error", err)
			return nil, huma.Error500InternalServerError("")
		}

		response := &CreateEventResponseEnvelope{Body: CreateEventResponse{ID: id}}
		return response, nil
	})

	// Read an existing event by id.
	huma.Get(api, "/events/{id}", func(ctx context.Context, input *struct {
		ID int32 `path:"id"`
	}) (*GetEventResponseEnvelope, error) {
		event, err := service.GetEvent(ctx, input.ID)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				return nil, huma.Error404NotFound("")
			}

			slog.Error("Issue fetching event", "event_id", input.ID, "error", err)
			return nil, huma.Error500InternalServerError("")
		}

		response := &GetEventResponseEnvelope{Body: MapToEventResponse(event)}
		return response, nil
	})

	// Update an existing event.
	huma.Put(api, "/events/{id}", func(ctx context.Context, input *struct {
		ID   int32 `path:"id"`
		Body WriteEventRequest
	}) (*struct{}, error) {
		event := MapToEvent(input.Body)
		event.ID = input.ID
		if !event.IsValid() {
			return nil, huma.Error422UnprocessableEntity("")
		}

		err := service.UpdateEvent(ctx, event)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				slog.Error(
					"Attempt to update a non-existent or deleted event",
					"event_id", input.ID,
					"request_data", input.Body,
					"error", err,
				)
				return nil, huma.Error404NotFound("")
			}

			slog.Error("Issue updating event", "request_data", input.Body, "error", err)
			return nil, huma.Error500InternalServerError("")
		}
		return nil, nil
	})

	// Delete an existing event.
	huma.Delete(api, "/events/{id}", func(ctx context.Context, input *struct {
		ID int32 `path:"id"`
	}) (*struct{}, error) {
		err := service.DeleteEvent(ctx, input.ID)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				return nil, huma.Error404NotFound("")
			}

			slog.Error("Issue deleting event", "event_id", input.ID, "error", err)
			return nil, huma.Error500InternalServerError("")
		}
		return nil, nil
	})
}

func RegisterTicketsHandlers(api huma.API, service *services.TicketsService) {
	// Release tickets for an event.
	huma.Post(api, "/events/{id}/tickets", func(ctx context.Context, input *struct {
		EventID int32 `path:"id"`
		Body    WriteTicketReleaseRequest
	}) (*struct{}, error) {
		tickets := MapToTickets(input.Body, input.EventID)
		err := service.AddTickets(ctx, input.EventID, tickets)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				return nil, huma.Error404NotFound("")
			}

			slog.Error(
				"Issue releasing tickets",
				"event_id", input.EventID,
				"request_data", input.Body,
				"error", err,
			)
			return nil, huma.Error500InternalServerError("")
		}
		return nil, nil
	})

	// Read tickets for an event.
	huma.Get(api, "/events/{id}/tickets", func(ctx context.Context, input *struct {
		EventID int32 `path:"id"`
	}) (*GetAvailableTicketsAggregateResponseEnvelope, error) {
		ticketAggregates, err := service.GetAvailableTickets(ctx, input.EventID)
		if err != nil {
			if errors.Is(err, repos.ErrNoSuchEntity) {
				return nil, huma.Error404NotFound("")
			}

			slog.Error(
				"Issue fetching available tickets",
				"event_id", input.EventID,
				"error", err,
			)
			return nil, huma.Error500InternalServerError("")
		}

		response := &GetAvailableTicketsAggregateResponseEnvelope{}
		response.Body = MapToAvailableTicketsAggregateResponse(ticketAggregates)
		return response, nil
	})

	// Set a purchase hold on a ticket.
	huma.Post(api, "/tickets/{id}/hold", func(ctx context.Context, input *struct {
		ID     int32  `path:"id"`
		UserID string `header:"x-user-id"`
	}) (*struct{}, error) {
		ticketID := input.ID
		holdID := input.UserID
		err := service.SetTicketHold(ctx, ticketID, holdID)
		if err != nil {
			if errors.Is(err, services.ErrInvalidHoldID) {
				slog.Error("Invalid hold id", "ticket_id", ticketID, "hold_id", holdID)
				return nil, huma.Error422UnprocessableEntity("")
			}

			if errors.Is(err, repos.ErrNoSuchEntity) {
				return nil, huma.Error404NotFound("")
			}

			if errors.Is(err, cache.ErrAlreadyHasHold) {
				slog.Error(
					"Attempt to place a hold on an already held ticket",
					"ticket_id", ticketID,
					"hold_id", holdID,
				)
				return nil, huma.Error422UnprocessableEntity("")
			}

			slog.Error(
				"Issue setting a ticket hold",
				"ticket_id", ticketID,
				"hold_id", holdID,
				"error", err,
			)
			return nil, huma.Error500InternalServerError("")
		}
		return nil, nil
	})

	// Purchase a ticket.
	huma.Post(api, "/tickets/{id}/purchase", func(ctx context.Context, input *struct {
		ID int32 `path:"id"`
		// TODO: Should probably model user id as an int and convert to a string
		// for hold id.
		UserID string `header:"x-user-id"`
		Card   Card
	}) (*PaymentResponseEnvelope, error) {
		holdID := input.UserID
		userID, err := strconv.Atoi(input.UserID)
		if err != nil {
			slog.Error("Unable to parse user id", "user_id", input.UserID, "error", err)
			return nil, huma.Error500InternalServerError("")
		}

		card := payment.Card{
			Name:            input.Card.Name,
			Address:         input.Card.Address,
			Number:          input.Card.Number,
			ExpirationMonth: input.Card.ExpirationMonth,
			ExpirationYear:  input.Card.ExpirationYear,
			CVC:             input.Card.CVC,
		}

		ticket, err := service.GetHeldTicket(ctx, input.ID, holdID)
		if err != nil {
			if errors.Is(err, services.ErrInvalidHoldID) {
				slog.Error("Invalid hold id", "ticket_id", input.ID, "hold_id", holdID)
				return nil, huma.Error422UnprocessableEntity("")
			}

			if errors.Is(err, cache.ErrNotFound) {
				slog.Error(
					"Attempt to purchase a ticket without a purchase hold",
					"ticket_id", input.ID,
					"hold_id", holdID,
					"error", err,
				)
				return nil, huma.Error422UnprocessableEntity("")
			}

			if errors.Is(err, services.ErrHoldIDMismatch) {
				slog.Error(
					"Attempt to purchase a held ticket with incorrect hold id",
					"ticket_id", input.ID,
					"hold_id", holdID,
					"error", err,
				)
				return nil, huma.Error422UnprocessableEntity("")
			}

			slog.Error(
				"Issue purchasing a ticket",
				"ticket_id", input.ID,
				"hold_id", holdID,
				"error", err,
			)
			return nil, huma.Error500InternalServerError("")
		}

		paymentClient := &payment.PaymentClient{}
		paymentSuccessful, err := paymentClient.SubmitPayment(ticket, card)
		if err != nil {
			slog.Error(
				"Issue on ticket payment submission",
				"ticket_id", input.ID,
				"hold_id", holdID,
				"error", err,
			)
			return nil, huma.Error500InternalServerError("")
		}

		response := PaymentResponse{Success: paymentSuccessful}

		if !paymentSuccessful {
			return &PaymentResponseEnvelope{Body: response}, nil
		}

		err = service.SetTicketPurchaser(ctx, input.ID, int32(userID))
		if err != nil {
			slog.Error(
				"Issue setting ticket purchaser",
				"ticket_id", input.ID,
				"purchaser_id", userID,
				"error", err,
			)
			return nil, huma.Error500InternalServerError("")
		}

		// NB: Purchase hold is left intact - as it only serves to filter out
		// unavailable tickets, and the ticket is no longer available once
		// purchased, there shouldn't be any impact to functionality. The
		// purchase hold should also be removed from Redis once its expiration
		// time is hit.
		return &PaymentResponseEnvelope{Body: response}, nil
	})
}
