package search

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	opensearch "github.com/opensearch-project/opensearch-go/v4"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/opensearch-project/opensearch-go/v4/opensearchutil"
)

const (
	orderAscending = "asc"
	dateFormat     = "yyyy-MM-dd"
)

// NewHTTPClient instantiates an OpenSearch HTTP client. This raw client may be
// used for test setup/teardown.
func NewHTTPClient(address, username, password string) (*opensearchapi.Client, error) {
	return opensearchapi.NewClient(
		opensearchapi.Config{
			Client: opensearch.Config{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				},
				Addresses: []string{address},
				Username:  username,
				Password:  password,
			},
		},
	)
}

type SearchClienter interface {
	SearchEvents(context.Context, string, time.Time, int32) ([]EventDocument, error)
	SearchVenues(context.Context, string, int32) ([]VenueDocument, error)
}

type SearchClient struct {
	conn        *opensearchapi.Client
	EventsIndex string
	VenuesIndex string
}

func NewSearchClient(address, username, password, eventsIndex, venuesIndex string) (*SearchClient, error) {
	client, err := NewHTTPClient(address, username, password)
	if err != nil {
		return nil, err
	}
	return &SearchClient{conn: client, EventsIndex: eventsIndex, VenuesIndex: venuesIndex}, nil
}

func NewSearchClientFromHTTPClient(
	client *opensearchapi.Client,
	eventsIndex,
	venuesIndex string,
) *SearchClient {
	return &SearchClient{conn: client, EventsIndex: eventsIndex, VenuesIndex: venuesIndex}
}

func (client *SearchClient) search(
	ctx context.Context,
	index string,
	payload interface{},
) (*opensearchapi.SearchResp, error) {
	return client.conn.Search(
		ctx,
		&opensearchapi.SearchReq{
			Indices: []string{index},
			Body:    opensearchutil.NewJSONReader(&payload),
		},
	)
}

// DeletedQuery represents a query that checks whether a document's
// `deleted` field is true/false.
type DeletedQuery struct {
	Match struct {
		Deleted bool `json:"deleted"`
	} `json:"match"`
}

func MakeExcludeDeletedQuery() DeletedQuery {
	clause := DeletedQuery{}
	clause.Match.Deleted = true
	return clause
}

// SearchEvents searches for event documents that match the given search term,
// and that begin no earlier than `startTime`.
func (client *SearchClient) SearchEvents(
	ctx context.Context,
	searchTerm string,
	startTime time.Time,
	size int32,
) (events []EventDocument, err error) {
	type DateRangeStartsAtQuery struct {
		Range struct {
			StartsAt struct {
				LT     string `json:"lt"`
				Format string `json:"format"`
			} `json:"starts_at"`
		} `json:"range"`
	}
	type MatchNameQuery struct {
		Match struct {
			Name string `json:"name"`
		} `json:"match"`
	}
	type SortByStartsAt struct {
		StartsAt struct {
			Order string `json:"order"`
		} `json:"starts_at"`
	}

	excludeDeletedQuery := MakeExcludeDeletedQuery()

	matchNameQuery := MatchNameQuery{}
	matchNameQuery.Match.Name = searchTerm

	sortBy := SortByStartsAt{}
	sortBy.StartsAt.Order = orderAscending

	payload := struct {
		Query struct {
			Bool struct {
				MustNot []interface{} `json:"must_not"`
				Should  []interface{} `json:"should"`
			} `json:"bool"`
		} `json:"query"`
		Sort SortByStartsAt `json:"sort"`
		Size int32          `json:"size"`
	}{}
	payload.Query.Bool.MustNot = []interface{}{excludeDeletedQuery}
	payload.Query.Bool.Should = []interface{}{matchNameQuery}
	payload.Sort = sortBy
	payload.Size = size

	if !startTime.IsZero() {
		dateRangeQuery := DateRangeStartsAtQuery{}
		dateRangeQuery.Range.StartsAt.Format = dateFormat
		dateRangeQuery.Range.StartsAt.LT = startTime.Format(time.DateOnly)
		payload.Query.Bool.MustNot = append(payload.Query.Bool.MustNot, dateRangeQuery)
	}

	response, err := client.search(ctx, client.EventsIndex, payload)
	if err != nil {
		return
	}

	return UnmarshalDocuments[EventDocument](*response)
}

// SearchVenues searches for venue documents that contain the given search term.
func (client *SearchClient) SearchVenues(
	ctx context.Context,
	searchTerm string,
	size int32,
) (venues []VenueDocument, err error) {
	const slop = 2
	const maxExpansions = 5

	type MatchPhrasePrefixSubQuery struct {
		Query         string `json:"query"`
		Slop          uint8  `json:"slop"`
		MaxExpansions uint8  `json:"max_expansions"`
	}
	type MatchPhrasePrefixNameQuery struct {
		Name MatchPhrasePrefixSubQuery `json:"name"`
	}
	type MatchPhrasePrefixDescriptionQuery struct {
		Description MatchPhrasePrefixSubQuery `json:"description"`
	}
	type MatchPhrasePrefixQuery struct {
		MatchPhrasePrefix interface{} `json:"match_phrase_prefix"`
	}

	excludeDeletedQuery := MakeExcludeDeletedQuery()

	nameQuery := MatchPhrasePrefixNameQuery{
		Name: MatchPhrasePrefixSubQuery{
			Query:         searchTerm,
			Slop:          slop,
			MaxExpansions: maxExpansions,
		},
	}
	descriptionQuery := MatchPhrasePrefixDescriptionQuery{
		Description: MatchPhrasePrefixSubQuery{
			Query:         searchTerm,
			Slop:          slop,
			MaxExpansions: maxExpansions,
		},
	}

	payload := struct {
		Query struct {
			Bool struct {
				MustNot []interface{}            `json:"must_not"`
				Should  []MatchPhrasePrefixQuery `json:"should"`
			} `json:"bool"`
		} `json:"query"`
		Size int32 `json:"size"`
	}{}
	payload.Query.Bool.MustNot = []interface{}{excludeDeletedQuery}
	payload.Query.Bool.Should = []MatchPhrasePrefixQuery{
		{MatchPhrasePrefix: nameQuery},
		{MatchPhrasePrefix: descriptionQuery},
	}
	payload.Size = size

	response, err := client.search(ctx, client.VenuesIndex, payload)
	if err != nil {
		return
	}

	return UnmarshalDocuments[VenueDocument](*response)
}
