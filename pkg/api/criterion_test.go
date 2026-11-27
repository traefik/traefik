package api

import (
	"math"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPaginationRejectsOverflow(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/http/routers?per_page=2&page="+strconv.Itoa(math.MaxInt), nil)

	_, err := pagination(request, 2)
	require.Error(t, err)
}

func TestPaginationNextPageDoesNotOverflow(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/http/routers?per_page="+strconv.Itoa(math.MaxInt-1)+"&page=2", nil)

	info, err := pagination(request, math.MaxInt)
	require.NoError(t, err)
	assert.Equal(t, 1, info.nextPage)
}
