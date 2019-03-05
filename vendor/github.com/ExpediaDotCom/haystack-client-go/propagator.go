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
	carrierMap := carrier.(map[string]string)
	carrierMap[p.opts.TraceIDKEY()] = ctx.TraceID
	carrierMap[p.opts.SpanIDKEY()] = ctx.SpanID
	carrierMap[p.opts.ParentSpanIDKEY()] = ctx.ParentID

	ctx.ForeachBaggageItem(func(key, value string) bool {
		carrierMap[fmt.Sprintf("%s%s", p.opts.BaggageKeyPrefix(), key)] = p.codex.Encode(ctx.Baggage[key])
		return true
	})

	return nil
}

/*Extract the span context from the carrier*/
func (p *TextMapPropagator) Extract(carrier interface{}) (*SpanContext, error) {
	baggage := make(map[string](string))
	carrierMap := carrier.(map[string]string)

	baggageKeyLowerCasePrefix := strings.ToLower(p.opts.BaggageKeyPrefix())
	traceIDKeyLowerCase := strings.ToLower(p.opts.TraceIDKEY())
	spanIDKeyLowerCase := strings.ToLower(p.opts.SpanIDKEY())
	parentSpanIDKeyLowerCase := strings.ToLower(p.opts.ParentSpanIDKEY())

	traceID := ""
	spanID := ""
	parentSpanID := ""

	for k, v := range carrierMap {
		lcKey := strings.ToLower(k)
		if strings.HasPrefix(lcKey, baggageKeyLowerCasePrefix) {
			keySansPrefix := k[len(p.opts.BaggageKeyPrefix()):]
			baggage[keySansPrefix] = p.codex.Decode(v)
		} else if lcKey == traceIDKeyLowerCase {
			traceID = carrierMap[k]
		} else if lcKey == spanIDKeyLowerCase {
			spanID = carrierMap[k]
		} else if lcKey == parentSpanIDKeyLowerCase {
			parentSpanID = carrierMap[k]
		}
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
