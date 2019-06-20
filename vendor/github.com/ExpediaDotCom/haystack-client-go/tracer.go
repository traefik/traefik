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
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

/*Tracer implements the opentracing.tracer*/
type Tracer struct {
	serviceName     string
	logger          Logger
	dispatcher      Dispatcher
	commonTags      []opentracing.Tag
	timeNow         func() time.Time
	idGenerator     func() string
	propagators     map[interface{}]Propagator
	useDualSpanMode bool
}

/*NewTracer creates a new tracer*/
func NewTracer(
	serviceName string,
	dispatcher Dispatcher,
	options ...TracerOption,
) (opentracing.Tracer, io.Closer) {
	tracer := &Tracer{
		serviceName:     serviceName,
		dispatcher:      dispatcher,
		useDualSpanMode: false,
	}
	tracer.propagators = make(map[interface{}]Propagator)
	tracer.propagators[opentracing.TextMap] = NewDefaultTextMapPropagator()
	tracer.propagators[opentracing.HTTPHeaders] = NewTextMapPropagator(PropagatorOpts{}, URLCodex{})
	for _, option := range options {
		option(tracer)
	}

	if tracer.timeNow == nil {
		tracer.timeNow = time.Now
	}

	if tracer.logger == nil {
		tracer.logger = NullLogger{}
	}

	if tracer.idGenerator == nil {
		tracer.idGenerator = func() string {
			_uuid, err := uuid.NewUUID()
			if err != nil {
				panic(err)
			}
			return _uuid.String()
		}
	}

	dispatcher.SetLogger(tracer.logger)
	return tracer, tracer
}

/*StartSpan starts a new span*/
func (tracer *Tracer) StartSpan(
	operationName string,
	options ...opentracing.StartSpanOption,
) opentracing.Span {
	sso := opentracing.StartSpanOptions{}

	for _, o := range options {
		o.Apply(&sso)
	}

	if sso.StartTime.IsZero() {
		sso.StartTime = tracer.timeNow()
	}

	var followsFromIsParent = false
	var parent *SpanContext

	for _, ref := range sso.References {
		if ref.Type == opentracing.ChildOfRef {
			if parent == nil || followsFromIsParent {
				parent = ref.ReferencedContext.(*SpanContext)
			}
		} else if ref.Type == opentracing.FollowsFromRef {
			if parent == nil {
				parent = ref.ReferencedContext.(*SpanContext)
				followsFromIsParent = true
			}
		}
	}

	spanContext := tracer.createSpanContext(parent, tracer.isServerSpan(sso.Tags))

	span := &_Span{
		tracer:        tracer,
		context:       spanContext,
		operationName: operationName,
		startTime:     sso.StartTime,
		duration:      0,
	}

	for _, tag := range tracer.Tags() {
		span.SetTag(tag.Key, tag.Value)
	}
	for k, v := range sso.Tags {
		span.SetTag(k, v)
	}

	return span
}

func (tracer *Tracer) isServerSpan(spanTags map[string]interface{}) bool {
	if spanKind, ok := spanTags[string(ext.SpanKind)]; ok && spanKind == "server" {
		return true
	}
	return false
}

func (tracer *Tracer) createSpanContext(parent *SpanContext, isServerSpan bool) *SpanContext {
	if parent == nil || !parent.IsValid() {
		return &SpanContext{
			TraceID: tracer.idGenerator(),
			SpanID:  tracer.idGenerator(),
		}
	}

	// This is a check to see if the tracer is configured to support single
	// single span type (Zipkin style shared span id) or
	// dual span type (client and server having their own span ids ).
	// a. If tracer is not of dualSpanType and if it is a server span then we
	// just return the parent context with the same shared span ids
	// b. If tracer is not of dualSpanType and if the parent context is an extracted one from the wire
	// then we assume this is the first span in the server and so just return the parent context
	// with the same shared span ids
	if !tracer.useDualSpanMode && (isServerSpan || parent.IsExtractedContext) {
		return &SpanContext{
			TraceID:            parent.TraceID,
			SpanID:             parent.SpanID,
			ParentID:           parent.ParentID,
			Baggage:            parent.Baggage,
			IsExtractedContext: false,
		}
	}
	return &SpanContext{
		TraceID:            parent.TraceID,
		SpanID:             tracer.idGenerator(),
		ParentID:           parent.SpanID,
		Baggage:            parent.Baggage,
		IsExtractedContext: false,
	}
}

/*Inject implements Inject() method of opentracing.Tracer*/
func (tracer *Tracer) Inject(ctx opentracing.SpanContext, format interface{}, carrier interface{}) error {
	c, ok := ctx.(*SpanContext)
	if !ok {
		return opentracing.ErrInvalidSpanContext
	}
	if injector, ok := tracer.propagators[format]; ok {
		return injector.Inject(c, carrier)
	}
	return opentracing.ErrUnsupportedFormat
}

/*Extract implements Extract() method of opentracing.Tracer*/
func (tracer *Tracer) Extract(
	format interface{},
	carrier interface{},
) (opentracing.SpanContext, error) {
	if extractor, ok := tracer.propagators[format]; ok {
		return extractor.Extract(carrier)
	}
	return nil, opentracing.ErrUnsupportedFormat
}

/*Tags return all common tags */
func (tracer *Tracer) Tags() []opentracing.Tag {
	return tracer.commonTags
}

/*DispatchSpan dispatches the span to a dispatcher*/
func (tracer *Tracer) DispatchSpan(span *_Span) {
	if tracer.dispatcher != nil {
		tracer.dispatcher.Dispatch(span)
	}
}

/*Close closes the tracer*/
func (tracer *Tracer) Close() error {
	if tracer.dispatcher != nil {
		tracer.dispatcher.Close()
	}
	return nil
}
