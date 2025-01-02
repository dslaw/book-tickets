package cache_test

import (
	"testing"

	"github.com/dslaw/book-tickets/pkg/cache"
	"github.com/stretchr/testify/assert"
)

func TestTicketHoldRepoMakeKey(t *testing.T) {
	id := int32(123)
	repo := cache.TicketHoldClient{}
	actual := repo.MakeKey(id)
	assert.Equal(t, "123", actual)
}

func TestTicketHoldRepoJoinMGetResults(t *testing.T) {
	fields := []string{"a", "b", "c"}
	values := []interface{}{"1", nil, "3"}

	repo := &cache.TicketHoldClient{}
	actual := repo.JoinMGetResults(fields, values)
	assert.Equal(t, map[string]string{"a": "1", "c": "3"}, actual)
}
