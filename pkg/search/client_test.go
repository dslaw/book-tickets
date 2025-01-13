package search_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dslaw/book-tickets/pkg/search"
	_ "github.com/joho/godotenv/autoload"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// WriteSearchDocuments creates test event and venue documents in OpenSearch,
// and returns a teardown function that can be used to delete the created documents.
func WriteSearchDocuments(ctx context.Context, suite *SearchClientTestSuite) func(context.Context) {
	indexDocument := func(index, id, document string) error {
		req := opensearchapi.IndexReq{
			Index:      index,
			DocumentID: id,
			Body:       strings.NewReader(document),
		}
		_, err := suite.OpenSearchClient.Index(ctx, req)
		return err
	}

	// XXX: Search integration tests here use the same test indices as the API
	// tests (see `pkg/api/handlers_test.go`). Test data is crafted such that
	// the search terms and test documents used in different test packages
	// are exclusive (ie a search term in one package should not match test data
	// in another package), but care must be taken to ensure this.
	// A prefix is added to test document ids here in order to disambiguate them
	// and prevent clashes with other test packages that perform setup against
	// the same test indices.
	const idPrefix = "test-search-client"
	t := suite.T()

	// Set up event documents.
	eventDocumentIDs := []string{"1", "2", "3", "4"}
	eventDocuments := []string{
		// Event that should be matched on the `name` field.
		`{"id": 1, "name": "match 1", "starts_at": "2024-06-30T20:00:00.000Z", "deleted": false}`,
		// Event that should be matched on the `name` field, but starts earlier
		// than others, and should be excluded when the starting from datetime
		// is set appropriately.
		`{"id": 2, "name": "match 2", "starts_at": "2024-05-30T20:00:00.000Z", "deleted": false}`,
		// Event that should not be matched due to the `name` field.
		`{"id": 3, "name": "miss", "starts_at": "2024-06-30T20:00:00.000Z", "deleted": false}`,
		// Soft-deleted event.
		`{"id": 4, "name": "match 3", "starts_at": "2024-06-30T20:00:00.000Z", "deleted": true}`,
	}
	for idx, documentID := range eventDocumentIDs {
		id := fmt.Sprintf("%s-%s", idPrefix, documentID)
		err := indexDocument(suite.EventsIndex, id, eventDocuments[idx])
		if err != nil {
			assert.FailNow(t, "Error setting up test data", err)
		}
	}

	// Set up venue documents.
	venueDocumentIDs := []string{"1", "2", "3", "4"}
	venueDocuments := []string{
		// Venue that should be matched on the `name` field.
		`{"id": 1, "name": "match", "deleted": false}`,
		// Venue that should be matched on the `description` field.
		`{"id": 1, "name": "name", "description": "match", "deleted": false}`,
		// Venue that not be matched due to neither `name` nor `description`
		// matching..
		`{"id": 1, "name": "miss", "deleted": false}`,
		// Soft-deleted venue.
		`{"id": 1, "name": "match", "deleted": true}`,
	}
	for idx, documentID := range venueDocumentIDs {
		id := fmt.Sprintf("%s-%s", idPrefix, documentID)
		err := indexDocument(suite.VenuesIndex, id, venueDocuments[idx])
		if err != nil {
			assert.FailNow(t, "Error setting up test data", err)
		}
	}

	// Refresh indices to ensure that test data can be searched for immediately.
	req := opensearchapi.IndicesRefreshReq{Indices: []string{suite.EventsIndex, suite.VenuesIndex}}
	_, err := suite.OpenSearchClient.Indices.Refresh(ctx, &req)
	if err != nil {
		assert.FailNow(t, "Error setting up test data", err)
	}

	// Create and return a closure that can be used to delete the created
	// documents. This provides the teardown function with access to the
	// document ids that should be deleted.
	return func(ctx context.Context) {
		// Delete event documents.
		for _, documentID := range eventDocumentIDs {
			id := fmt.Sprintf("%s-%s", idPrefix, documentID)
			req := opensearchapi.DocumentDeleteReq{Index: suite.EventsIndex, DocumentID: id}
			_, err := suite.OpenSearchClient.Document.Delete(ctx, req)
			if err != nil {
				assert.FailNow(t, "Error tearing down test data", err)
			}
		}

		// Delete venue documents.
		for _, documentID := range venueDocumentIDs {
			id := fmt.Sprintf("%s-%s", idPrefix, documentID)
			req := opensearchapi.DocumentDeleteReq{Index: suite.VenuesIndex, DocumentID: id}
			_, err := suite.OpenSearchClient.Document.Delete(ctx, req)
			if err != nil {
				assert.FailNow(t, "Error tearing down test data", err)
			}
		}
	}
}

// ClearSearchDocuments deletes documents created by `WriteSearchDocuments` from
// OpenSearch.
func ClearSearchDocuments(
	ctx context.Context,
	client *opensearchapi.Client,
	eventsIndex,
	venuesIndex string,
) error {
	documentIDs := []string{"1", "2", "3", "4"}
	for _, documentID := range documentIDs {
		req := opensearchapi.DocumentDeleteReq{
			Index:      eventsIndex,
			DocumentID: documentID,
		}
		_, err := client.Document.Delete(ctx, req)
		if err != nil {
			return err
		}
	}

	return nil
}

type SearchClientTestSuite struct {
	suite.Suite
	OpenSearchClient *opensearchapi.Client
	EventsIndex      string
	VenuesIndex      string
	teardown         func(context.Context)
}

func (suite *SearchClientTestSuite) SetupSuite() {
	openSearchURL := os.Getenv("TEST_SEARCH_URL_LOCAL")
	openSearchUser := os.Getenv("SEARCH_USER")
	openSearchPassword := os.Getenv("SEARCH_PASSWORD")
	openSearchClient, err := search.NewHTTPClient(openSearchURL, openSearchUser, openSearchPassword)
	if err != nil {
		assert.FailNowf(suite.T(), "Unable to create OpenSearch client", err.Error())
	}

	suite.OpenSearchClient = openSearchClient
	suite.EventsIndex = os.Getenv("TEST_SEARCH_EVENTS_INDEX")
	suite.VenuesIndex = os.Getenv("TEST_SEARCH_VENUES_INDEX")

	suite.teardown = WriteSearchDocuments(context.Background(), suite)
}

func (suite *SearchClientTestSuite) TearDownSuite() {
	suite.teardown(context.Background())
}

// Test that the given search term searches against event names, and that
// soft-deleted events are excluded from the results.
func (suite *SearchClientTestSuite) TestSearchEvents() {
	t := suite.T()
	client := search.NewSearchClientFromHTTPClient(
		suite.OpenSearchClient,
		suite.EventsIndex,
		suite.VenuesIndex,
	)
	actual, err := client.SearchEvents(context.Background(), "match", time.Time{}, 10)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(actual))
}

// Test that the non-matching events are not returned.
func (suite *SearchClientTestSuite) TestSearchEventsWhenTermDoesntMatch() {
	t := suite.T()
	client := search.NewSearchClientFromHTTPClient(
		suite.OpenSearchClient,
		suite.EventsIndex,
		suite.VenuesIndex,
	)
	actual, err := client.SearchEvents(context.Background(), "negative case", time.Time{}, 10)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(actual))
}

// Test that events which start earlier than the given start time are excluded
// from the results.
func (suite *SearchClientTestSuite) TestSearchEventsExcludesEarlierStartingEvents() {
	t := suite.T()
	client := search.NewSearchClientFromHTTPClient(
		suite.OpenSearchClient,
		suite.EventsIndex,
		suite.VenuesIndex,
	)
	startsAt, _ := time.Parse(time.RFC3339, "2024-06-30T00:00:00.000Z")
	actual, err := client.SearchEvents(context.Background(), "match", startsAt, 10)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(actual))
}

// Test that search results are limited by the `limit` argument.
func (suite *SearchClientTestSuite) TestSearchEventsAreLimited() {
	t := suite.T()
	client := search.NewSearchClientFromHTTPClient(
		suite.OpenSearchClient,
		suite.EventsIndex,
		suite.VenuesIndex,
	)
	actual, err := client.SearchEvents(context.Background(), "match", time.Time{}, 1)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(actual))
}

// Test that the given search term searches against venue names and
// descriptions, and that soft-deleted venues are excluded from the results.
func (suite *SearchClientTestSuite) TestSearchVenues() {
	t := suite.T()
	client := search.NewSearchClientFromHTTPClient(
		suite.OpenSearchClient,
		suite.EventsIndex,
		suite.VenuesIndex,
	)
	actual, err := client.SearchVenues(context.Background(), "match", 10)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(actual))
}

// Test that the non-matching venues are not returned.
func (suite *SearchClientTestSuite) TestSearchVenuesWhenTermDoesntMatch() {
	t := suite.T()
	client := search.NewSearchClientFromHTTPClient(
		suite.OpenSearchClient,
		suite.EventsIndex,
		suite.VenuesIndex,
	)
	actual, err := client.SearchVenues(context.Background(), "negative case", 10)

	assert.Nil(t, err)
	assert.Equal(t, 0, len(actual))
}

// Test that search results are limited by the `limit` argument.
func (suite *SearchClientTestSuite) TestSearchVeneusAreLimited() {
	t := suite.T()
	client := search.NewSearchClientFromHTTPClient(
		suite.OpenSearchClient,
		suite.EventsIndex,
		suite.VenuesIndex,
	)
	actual, err := client.SearchVenues(context.Background(), "match", 1)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(actual))
}

func TestSearchClientTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping testing in short mode")
	}
	suite.Run(t, new(SearchClientTestSuite))
}
