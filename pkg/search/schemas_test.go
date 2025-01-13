package search_test

import (
	"testing"
	"time"

	"github.com/dslaw/book-tickets/pkg/search"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
)

func TestUnmarshalDocumentsWithEventDocument(t *testing.T) {
	document1 := `{
    "id": 1,
    "name": "Test Event 1",
    "venue": {"id": 1, "name": "Test Venue 1"},
    "deleted": false,
    "ends_at": "2024-01-01T03:00:00+00:00",
    "starts_at": "2024-01-01T00:00:00+00:00",
    "description": null,
    "_meta": {
        "venues": {"id": [1]},
        "events": {"id": ["1"]}
    }
}`
	document2 := `{
    "id": 2,
    "name": "Test Event 2",
    "venue": {"id": 1, "name": "Test Venue 1"},
    "deleted": false,
    "ends_at": "2024-01-02T03:00:00+00:00",
    "starts_at": "2024-01-02T00:00:00+00:00",
    "description": "Testing",
    "_meta": {
        "venues": {"id": [1]},
        "events": {"id": ["2"]}
    }
}`

	document1StartsAt, _ := time.Parse(time.RFC3339, "2024-01-01T00:00:00+00:00")
	document1EndsAt, _ := time.Parse(time.RFC3339, "2024-01-01T03:00:00+00:00")
	document2StartsAt, _ := time.Parse(time.RFC3339, "2024-01-02T00:00:00+00:00")
	document2EndsAt, _ := time.Parse(time.RFC3339, "2024-01-02T03:00:00+00:00")

	expected := []search.EventDocument{
		{
			ID:          1,
			Name:        "Test Event 1",
			Description: "",
			StartsAt:    document1StartsAt,
			EndsAt:      document1EndsAt,
			Venue:       search.EventVenue{ID: 1, Name: "Test Venue 1"},
			Deleted:     false,
		},
		{
			ID:          2,
			Name:        "Test Event 2",
			Description: "Testing",
			StartsAt:    document2StartsAt,
			EndsAt:      document2EndsAt,
			Venue:       search.EventVenue{ID: 1, Name: "Test Venue 1"},
			Deleted:     false,
		},
	}

	resp := opensearchapi.SearchResp{}
	resp.Hits.Hits = []opensearchapi.SearchHit{
		{Source: []byte(document1)},
		{Source: []byte(document2)},
	}
	actual, err := search.UnmarshalDocuments[search.EventDocument](resp)

	assert.Nil(t, err)
	assert.EqualValues(t, expected, actual)
}

func TestUnmarshalDocumentsWithVenueDocument(t *testing.T) {
	document1 := `{
    "id": 1,
    "city": "San Francisco",
    "name": "Test Venue 1",
    "address": "111 Front St",
    "deleted": false,
    "description": null,
    "subdivision": "CA",
    "country_code": "USA",
    "_meta": {
        "venues": {"id": [1]}
    }
}`
	document2 := `{
    "id": 2,
    "city": "San Francisco",
    "name": "Test Venue 2",
    "address": "222 Front St",
    "deleted": false,
    "description": "A second test venue",
    "subdivision": "CA",
    "country_code": "USA",
    "_meta": {
        "venues": {"id": [2]}
    }
}`

	expected := []search.VenueDocument{
		{
			ID:          1,
			Name:        "Test Venue 1",
			Description: "",
			Address:     "111 Front St",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
			Deleted:     false,
		},
		{
			ID:          2,
			Name:        "Test Venue 2",
			Description: "A second test venue",
			Address:     "222 Front St",
			City:        "San Francisco",
			Subdivision: "CA",
			CountryCode: "USA",
			Deleted:     false,
		},
	}

	resp := opensearchapi.SearchResp{}
	resp.Hits.Hits = []opensearchapi.SearchHit{
		{Source: []byte(document1)},
		{Source: []byte(document2)},
	}
	actual, err := search.UnmarshalDocuments[search.VenueDocument](resp)

	assert.Nil(t, err)
	assert.EqualValues(t, expected, actual)
}
