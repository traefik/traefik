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
)

/*SpanContext implements opentracing.spanContext*/
type SpanContext struct {
	// traceID represents globally unique ID of the trace.
	TraceID string

	// spanID represents span ID that must be unique within its trace
	SpanID string

	// parentID refers to the ID of the parent span.
	// Should be empty if the current span is a root span.
	ParentID string

	//Context baggage. The is a snapshot in time.
	Baggage map[string]string

	// set to true if extracted using a extractor in tracer
	IsExtractedContext bool
}

// IsValid indicates whether this context actually represents a valid trace.
func (context SpanContext) IsValid() bool {
	return context.TraceID != "" && context.SpanID != ""
}

/*ForeachBaggageItem implements opentracing.spancontext*/
func (context SpanContext) ForeachBaggageItem(handler func(k, v string) bool) {
	for k, v := range context.Baggage {
		if !handler(k, v) {
			break
		}
	}
}

// WithBaggageItem creates a new context with an extra baggage item.
func (context SpanContext) WithBaggageItem(key, value string) *SpanContext {
	var newBaggage map[string]string
	if context.Baggage == nil {
		newBaggage = map[string]string{key: value}
	} else {
		newBaggage = make(map[string]string, len(context.Baggage)+1)
		for k, v := range context.Baggage {
			newBaggage[k] = v
		}
		newBaggage[key] = value
	}
	return &SpanContext{
		TraceID:  context.TraceID,
		SpanID:   context.SpanID,
		ParentID: context.ParentID,
		Baggage:  newBaggage,
	}
}

/*ToString represents the string*/
func (context SpanContext) ToString() string {
	return fmt.Sprintf("%+v", context)
}
