package api

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
}

type GetVenueResponseEnvelope struct {
	Body GetVenueResponse
}
