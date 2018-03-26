package zipkintracer

import (
	"strings"

	"github.com/openzipkin/zipkin-go-opentracing/thrift/gen-go/zipkincore"
)

// Collector represents a Zipkin trace collector, which is probably a set of
// remote endpoints.
type Collector interface {
	Collect(*zipkincore.Span) error
	Close() error
}

// NopCollector implements Collector but performs no work.
type NopCollector struct{}

// Collect implements Collector.
func (NopCollector) Collect(*zipkincore.Span) error { return nil }

// Close implements Collector.
func (NopCollector) Close() error { return nil }

// MultiCollector implements Collector by sending spans to all collectors.
type MultiCollector []Collector

// Collect implements Collector.
func (c MultiCollector) Collect(s *zipkincore.Span) error {
	return c.aggregateErrors(func(coll Collector) error { return coll.Collect(s) })
}

// Close implements Collector.
func (c MultiCollector) Close() error {
	return c.aggregateErrors(func(coll Collector) error { return coll.Close() })
}

func (c MultiCollector) aggregateErrors(f func(c Collector) error) error {
	var e *collectionError
	for i, collector := range c {
		if err := f(collector); err != nil {
			if e == nil {
				e = &collectionError{
					errs: make([]error, len(c)),
				}
			}
			e.errs[i] = err
		}
	}
	return e
}

// CollectionError represents an array of errors returned by one or more
// failed Collector methods.
type CollectionError interface {
	Error() string
	GetErrors() []error
}

type collectionError struct {
	errs []error
}

func (c *collectionError) Error() string {
	errs := []string{}
	for _, err := range c.errs {
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	return strings.Join(errs, "; ")
}

// GetErrors implements CollectionError
func (c *collectionError) GetErrors() []error {
	return c.errs
}
