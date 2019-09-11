package api

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultPerPage = 100
	defaultPage    = 1
)

const nextPageHeader = "X-Next-Page"

type pageInfo struct {
	startIndex int
	endIndex   int
	nextPage   int
}

type searchCriterion struct {
	Search string `url:"search"`
	Status string `url:"status"`
}

func newSearchCriterion(query url.Values) *searchCriterion {
	if len(query) == 0 {
		return nil
	}

	search := query.Get("search")
	status := query.Get("status")

	if status == "" && search == "" {
		return nil
	}

	return &searchCriterion{Search: search, Status: status}
}

func (c *searchCriterion) withStatus(name string) bool {
	return c.Status == "" || strings.EqualFold(name, c.Status)
}

func (c *searchCriterion) searchIn(values ...string) bool {
	if c.Search == "" {
		return true
	}

	for _, v := range values {
		if strings.Contains(strings.ToLower(v), strings.ToLower(c.Search)) {
			return true
		}
	}

	return false
}

func pagination(request *http.Request, max int) (pageInfo, error) {
	perPage, err := getIntParam(request, "per_page", defaultPerPage)
	if err != nil {
		return pageInfo{}, err
	}

	page, err := getIntParam(request, "page", defaultPage)
	if err != nil {
		return pageInfo{}, err
	}

	startIndex := (page - 1) * perPage
	if startIndex != 0 && startIndex >= max {
		return pageInfo{}, fmt.Errorf("invalid request: page: %d, per_page: %d", page, perPage)
	}

	endIndex := startIndex + perPage
	if endIndex >= max {
		endIndex = max
	}

	nextPage := 1
	if page*perPage < max {
		nextPage = page + 1
	}

	return pageInfo{startIndex: startIndex, endIndex: endIndex, nextPage: nextPage}, nil
}

func getIntParam(request *http.Request, key string, defaultValue int) (int, error) {
	raw := request.URL.Query().Get(key)
	if raw == "" {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid request: %s: %d", key, value)
	}
	return value, nil
}
