package entities

import "time"

type VenueLocation struct {
	Address     string
	City        string
	Subdivision string
	CountryCode string
}

type Venue struct {
	ID          int32
	Name        string
	Description string
	Location    VenueLocation
}

type Performer struct {
	ID   int32
	Name string
}

type EventVenue struct {
	ID   int32
	Name string
}

type Event struct {
	ID          int32
	Name        string
	StartsAt    time.Time
	EndsAt      time.Time
	Description string
	Venue       EventVenue
	Performers  []Performer
}

func (e *Event) IsValid() bool {
	if e.EndsAt.Before(e.StartsAt) {
		return false
	}
	return true
}
