package search

import (
	"encoding/json"
	"time"

	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

type EventVenue struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type EventDocument struct {
	ID          int32      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	StartsAt    time.Time  `json:"starts_at"`
	EndsAt      time.Time  `json:"ends_at"`
	Venue       EventVenue `json:"venue"`
	Deleted     bool       `json:"deleted"`
}

type VenueDocument struct {
	ID          int32  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Subdivision string `json:"subdivision"`
	CountryCode string `json:"country_code"`
	Deleted     bool   `json:"deleted"`
}

func UnmarshalDocuments[D EventDocument | VenueDocument](resp opensearchapi.SearchResp) ([]D, error) {
	documents := make([]D, len(resp.Hits.Hits))
	for idx, hit := range resp.Hits.Hits {
		var document D
		err := json.Unmarshal(hit.Source, &document)
		if err != nil {
			return documents, err
		}
		documents[idx] = document
	}

	return documents, nil
}
