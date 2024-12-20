package repos

import (
	"time"

	"github.com/dslaw/book-tickets/pkg/db"
	"github.com/dslaw/book-tickets/pkg/entities"
	"github.com/jackc/pgx/v5/pgtype"
)

func MapNullableString(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: s != ""}
}

func MapTime(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func MapPurchaserID(id int32) pgtype.Int4 {
	return pgtype.Int4{Int32: id, Valid: true}
}

func MapGetEventRows(rows []db.GetEventRow) entities.Event {
	if len(rows) == 0 {
		return entities.Event{}
	}

	performers := make([]entities.Performer, 0)
	for _, row := range rows {
		if !row.PerformerID.Valid {
			continue
		}

		performers = append(performers, entities.Performer{
			ID:   row.PerformerID.Int32,
			Name: row.PerformerName.String,
		})
	}

	row := rows[0]
	return entities.Event{
		ID:          row.Event.ID,
		Name:        row.Event.Name,
		StartsAt:    row.Event.StartsAt.Time,
		EndsAt:      row.Event.EndsAt.Time,
		Description: row.Event.Description.String,
		Performers:  performers,
		Venue: entities.EventVenue{
			ID:   row.Event.VenueID,
			Name: row.VenueName,
		},
	}
}

func MapTicket(model db.Ticket) entities.Ticket {
	return entities.Ticket{
		ID:          model.ID,
		EventID:     model.EventID,
		PurchaserID: model.PurchaserID.Int32,
		IsPurchased: model.PurchaserID.Valid,
		Price:       uint8(model.Price),
		Seat:        model.Seat,
	}
}

func MapGetAvailableTicketRows(rows []db.GetAvailableTicketsRow) []entities.Ticket {
	tickets := make([]entities.Ticket, len(rows))
	for idx, row := range rows {
		tickets[idx] = MapTicket(row.Ticket)
	}
	return tickets
}
