/*
 *  Copyright 2018 Expedia Group.
 *
 *     Licensed under the Apache License, Version 2.0 (the "License");
 *     you may not use this file except in compliance with the License.
 *     You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *     Unless required by applicable law or agreed to in writing, software
 *     distributed under the License is distributed on an "AS IS" BASIS,
 *     WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *     See the License for the specific language governing permissions and
 *     limitations under the License.
 *
 */

package haystack

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/opentracing/opentracing-go"
)

/*Propagator defines the interface for injecting and extracing the SpanContext from the carrier*/
type Propagator interface {
	// Inject takes `SpanContext` and injects it into `carrier`
	Inject(ctx *SpanContext, carrier interface{}) error

	// Extract `SpanContext` from the `carrier`
	Extract(carrier interface{}) (*SpanContext, error)
}

/*Codex defines the interface for encoding and decoding the propagated data*/
type Codex interface {
	Encode(value string) string
	Decode(value string) string
}

/*DefaultCodex is a no op*/
type DefaultCodex struct{}

/*Encode a no-op for encoding the value*/
func (c DefaultCodex) Encode(value string) string { return value }

/*Decode a no-op for decoding the value*/
func (c DefaultCodex) Decode(value string) string { return value }

/*URLCodex encodes decodes a url*/
type URLCodex struct{}

/*Encode a no-op for encoding the value*/
func (c URLCodex) Encode(value string) string { return url.QueryEscape(value) }

/*Decode a no-op for decoding the value*/
func (c URLCodex) Decode(value string) string {
	decoded, err := url.QueryUnescape(value)
	if err == nil {
		return decoded
	}
	return ""
}

/*PropagatorOpts defines the options need by a propagator*/
type PropagatorOpts struct {
	TraceIDKEYName       string
	SpanIDKEYName        string
	ParentSpanIDKEYName  string
	BaggagePrefixKEYName string
}

var defaultPropagatorOpts = PropagatorOpts{}
var defaultCodex = DefaultCodex{}

/*TraceIDKEY returns the trace id key in the propagator*/
func (p *PropagatorOpts) TraceIDKEY() string {
	if p.TraceIDKEYName != "" {
		return p.TraceIDKEYName
	}
	return "Trace-ID"
}

/*SpanIDKEY returns the span id key in the propagator*/
func (p *PropagatorOpts) SpanIDKEY() string {
	if p.SpanIDKEYName != "" {
		return p.SpanIDKEYName
	}
	return "Span-ID"
}

/*ParentSpanIDKEY returns the parent span id key in the propagator*/
func (p *PropagatorOpts) ParentSpanIDKEY() string {
	if p.ParentSpanIDKEYName != "" {
		return p.ParentSpanIDKEYName
	}
	return "Parent-ID"
}

/*BaggageKeyPrefix returns the baggage key prefix*/
func (p *PropagatorOpts) BaggageKeyPrefix() string {
	if p.BaggagePrefixKEYName != "" {
		return p.BaggagePrefixKEYName
	}
	return "Baggage-"
}

/*TextMapPropagator implements Propagator interface*/
type TextMapPropagator struct {
	opts  PropagatorOpts
	codex Codex
}

/*Inject injects the span context in the carrier*/
func (p *TextMapPropagator) Inject(ctx *SpanContext, carrier interface{}) error {
	textMapWriter, ok := carrier.(opentracing.TextMapWriter)
	if !ok {
		return opentracing.ErrInvalidCarrier
	}

	textMapWriter.Set(p.opts.TraceIDKEY(), ctx.TraceID)
	textMapWriter.Set(p.opts.SpanIDKEY(), ctx.SpanID)
	textMapWriter.Set(p.opts.ParentSpanIDKEY(), ctx.ParentID)

	ctx.ForeachBaggageItem(func(key, value string) bool {
		textMapWriter.Set(fmt.Sprintf("%s%s", p.opts.BaggageKeyPrefix(), key), p.codex.Encode(ctx.Baggage[key]))
		return true
	})

	return nil
}

/*Extract the span context from the carrier*/
func (p *TextMapPropagator) Extract(carrier interface{}) (*SpanContext, error) {
	textMapReader, ok := carrier.(opentracing.TextMapReader)

	if !ok {
		return nil, opentracing.ErrInvalidCarrier
	}

	baggageKeyLowerCasePrefix := strings.ToLower(p.opts.BaggageKeyPrefix())
	traceIDKeyLowerCase := strings.ToLower(p.opts.TraceIDKEY())
	spanIDKeyLowerCase := strings.ToLower(p.opts.SpanIDKEY())
	parentSpanIDKeyLowerCase := strings.ToLower(p.opts.ParentSpanIDKEY())

	traceID := ""
	spanID := ""
	parentSpanID := ""

	baggage := make(map[string]string)
	err := textMapReader.ForeachKey(func(k, v string) error {
		lcKey := strings.ToLower(k)

		if strings.HasPrefix(lcKey, baggageKeyLowerCasePrefix) {
			keySansPrefix := lcKey[len(p.opts.BaggageKeyPrefix()):]
			baggage[keySansPrefix] = p.codex.Decode(v)
		} else if lcKey == traceIDKeyLowerCase {
			traceID = v
		} else if lcKey == spanIDKeyLowerCase {
			spanID = v
		} else if lcKey == parentSpanIDKeyLowerCase {
			parentSpanID = v
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &SpanContext{
		TraceID:            traceID,
		SpanID:             spanID,
		ParentID:           parentSpanID,
		Baggage:            baggage,
		IsExtractedContext: true,
	}, nil
}

/*NewDefaultTextMapPropagator returns a default text map propagator*/
func NewDefaultTextMapPropagator() *TextMapPropagator {
	return NewTextMapPropagator(defaultPropagatorOpts, defaultCodex)
}

/*NewTextMapPropagator returns a text map propagator*/
func NewTextMapPropagator(opts PropagatorOpts, codex Codex) *TextMapPropagator {
	return &TextMapPropagator{
		opts:  opts,
		codex: codex,
	}
}
