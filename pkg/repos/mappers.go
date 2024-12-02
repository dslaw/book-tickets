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
