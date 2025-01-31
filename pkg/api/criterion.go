package api

import (
	"fmt"
	"net/http"
	"net/url"
	"slices"
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
	Search         string `url:"search"`
	Status         string `url:"status"`
	ServiceName    string `url:"serviceName"`
	MiddlewareName string `url:"middlewareName"`
}

func newSearchCriterion(query url.Values) *searchCriterion {
	if len(query) == 0 {
		return nil
	}

	search := query.Get("search")
	status := query.Get("status")
	serviceName := query.Get("serviceName")
	middlewareName := query.Get("middlewareName")

	if status == "" && search == "" && serviceName == "" && middlewareName == "" {
		return nil
	}

	return &searchCriterion{
		Search:         search,
		Status:         status,
		ServiceName:    serviceName,
		MiddlewareName: middlewareName,
	}
}

func (c *searchCriterion) withStatus(name string) bool {
	return c.Status == "" || strings.EqualFold(name, c.Status)
}

func (c *searchCriterion) searchIn(values ...string) bool {
	if c.Search == "" {
		return true
	}

	return slices.ContainsFunc(values, func(v string) bool {
		return strings.Contains(strings.ToLower(v), strings.ToLower(c.Search))
	})
}

func (c *searchCriterion) filterService(name string) bool {
	if c.ServiceName == "" {
		return true
	}

	if strings.Contains(name, "@") {
		return c.ServiceName == name
	}

	before, _, _ := strings.Cut(c.ServiceName, "@")

	return before == name
}

func (c *searchCriterion) filterMiddleware(mns []string) bool {
	if c.MiddlewareName == "" {
		return true
	}

	for _, mn := range mns {
		if c.MiddlewareName == mn {
			return true
		}
	}

	return false
}

func pagination(request *http.Request, maximum int) (pageInfo, error) {
	perPage, err := getIntParam(request, "per_page", defaultPerPage)
	if err != nil {
		return pageInfo{}, err
	}

	page, err := getIntParam(request, "page", defaultPage)
	if err != nil {
		return pageInfo{}, err
	}

	startIndex := (page - 1) * perPage
	if startIndex != 0 && startIndex >= maximum {
		return pageInfo{}, fmt.Errorf("invalid request: page: %d, per_page: %d", page, perPage)
	}

	endIndex := startIndex + perPage
	if endIndex >= maximum {
		endIndex = maximum
	}

	nextPage := 1
	if page*perPage < maximum {
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
