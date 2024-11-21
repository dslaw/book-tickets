package entities

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
