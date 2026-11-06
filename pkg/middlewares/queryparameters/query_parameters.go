package queryparameters

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/middlewares"
)

const typeName = "QueryParameters"

// queryParameters is a middleware used to modify the query parameters of the request URL.
type queryParameters struct {
	next      http.Handler
	name      string
	set       map[string]string
	setNames  []string
	add       map[string]string
	addNames  []string
	deletions map[string]struct{}
}

// New creates a new query parameters middleware.
func New(ctx context.Context, next http.Handler, config dynamic.QueryParameters, name string) (http.Handler, error) {
	middlewares.GetLogger(ctx, name, typeName).Debug().Msg("Creating middleware")

	deletions := make(map[string]struct{}, len(config.Delete))
	for _, param := range config.Delete {
		if param == "" {
			return nil, errors.New("empty query parameter name in delete")
		}
		deletions[param] = struct{}{}
	}

	for param := range config.Set {
		if err := validateName(param, "set", deletions); err != nil {
			return nil, err
		}
	}

	for param := range config.Add {
		if err := validateName(param, "add", deletions); err != nil {
			return nil, err
		}
	}

	// Sorted so that the parameters appended to the query come out in a stable order.
	return &queryParameters{
		next:      next,
		name:      name,
		set:       config.Set,
		setNames:  slices.Sorted(maps.Keys(config.Set)),
		add:       config.Add,
		addNames:  slices.Sorted(maps.Keys(config.Add)),
		deletions: deletions,
	}, nil
}

func validateName(param, option string, deletions map[string]struct{}) error {
	if param == "" {
		return fmt.Errorf("empty query parameter name in %s", option)
	}

	if _, ok := deletions[param]; ok {
		return fmt.Errorf("query parameter %q is in both %s and delete", param, option)
	}

	return nil
}

func (q *queryParameters) GetTracingInformation() (string, string) {
	return q.name, typeName
}

func (q *queryParameters) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.RawQuery = q.modify(req.URL.RawQuery)
	req.RequestURI = req.URL.RequestURI()

	q.next.ServeHTTP(rw, req)
}

// modify works on the raw query rather than on url.Values, so that the parameters it does not
// touch keep their order and their original encoding: url.Values.Encode sorts and re-encodes
// everything, which can break backends that sign or otherwise depend on the exact query.
func (q *queryParameters) modify(rawQuery string) string {
	var pairs []string
	replaced := make(map[string]bool, len(q.set))

	if rawQuery != "" {
		for pair := range strings.SplitSeq(rawQuery, "&") {
			param := paramName(pair)

			if _, ok := q.deletions[param]; ok {
				continue
			}

			if value, ok := q.set[param]; ok {
				if !replaced[param] {
					pairs = append(pairs, encode(param, value))
					replaced[param] = true
				}
				continue
			}

			pairs = append(pairs, pair)
		}
	}

	for _, param := range q.setNames {
		if !replaced[param] {
			pairs = append(pairs, encode(param, q.set[param]))
		}
	}

	for _, param := range q.addNames {
		pairs = append(pairs, encode(param, q.add[param]))
	}

	return strings.Join(pairs, "&")
}

func paramName(pair string) string {
	name, _, _ := strings.Cut(pair, "=")

	unescaped, err := url.QueryUnescape(name)
	if err != nil {
		return name
	}

	return unescaped
}

func encode(param, value string) string {
	return url.QueryEscape(param) + "=" + url.QueryEscape(value)
}
