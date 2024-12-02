package api

import "time"

type WriteVenueRequest struct {
	Name        string `json:"name" minLength:"1" maxLength:"100"`
	Description string `json:"description" required:"false" maxLength:"200"`
	Location    struct {
		Address     string `json:"address" minLength:"1" maxLength:"200"`
		City        string `json:"city" minLength:"1" maxLength:"60"`
		Subdivision string `json:"subdivision" minLength:"1" maxLength:"60"`
		CountryCode string `json:"country_code" minLength:"3" maxLength:"3"`
	} `json:"location"`
}

type CreateVenueResponse struct {
	ID int32 `json:"id"`
}

type CreateVenueResponseEnvelope struct {
	Body CreateVenueResponse
}

type GetVenueResponse struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Location    struct {
		Address     string `json:"address"`
		City        string `json:"city"`
		Subdivision string `json:"subdivision"`
		CountryCode string `json:"country_code"`
	} `json:"location"`
	// TODO: Add upcoming events
}

type GetVenueResponseEnvelope struct {
	Body GetVenueResponse
}

type WritePerformerRequest struct {
	Name string `json:"name" minLength:"1" maxLength:"50"`
}

type WriteEventRequest struct {
	VenueID     int32                   `json:"venue_id"`
	Name        string                  `json:"name" minLength:"1" maxLength:"50"`
	Description string                  `json:"description" required:"false" maxLength:"200"`
	StartsAt    time.Time               `json:"starts_at"`
	EndsAt      time.Time               `json:"ends_at"`
	Performers  []WritePerformerRequest `json:"performers"`
}

// TODO: Can abstract these into a single CreateResourceResponse{Envelope}
type CreateEventResponse struct {
	ID int32 `json:"id"`
}

type CreateEventResponseEnvelope struct {
	Body CreateEventResponse
}

type EventVenueResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type EventPerformerResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type GetEventResponse struct {
	ID          int32                    `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	StartsAt    time.Time                `json:"starts_at"`
	EndsAt      time.Time                `json:"ends_at"`
	Venue       EventVenueResponse       `json:"venue"`
	Performers  []EventPerformerResponse `json:"performers"`
}

type GetEventResponseEnvelope struct {
	Body GetEventResponse
}
