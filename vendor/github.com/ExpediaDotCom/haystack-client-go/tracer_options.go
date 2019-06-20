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
	"github.com/opentracing/opentracing-go"
)

// TracerOption is a function that sets some option on the tracer
type TracerOption func(tracer *Tracer)

/*TracerOptions a list of tracer options*/
type TracerOptions struct{}

/*TracerOptionsFactory factory to create multiple tracer options*/
var TracerOptionsFactory TracerOptions

/*Propagator registers a new Propagator*/
func (t TracerOptions) Propagator(format interface{}, propagator Propagator) TracerOption {
	return func(tracer *Tracer) {
		tracer.propagators[format] = propagator
	}
}

/*Tag adds a common tag in every span*/
func (t TracerOptions) Tag(key string, value interface{}) TracerOption {
	return func(tracer *Tracer) {
		tracer.commonTags = append(tracer.commonTags, opentracing.Tag{Key: key, Value: value})
	}
}

/*Logger set the logger type*/
func (t TracerOptions) Logger(logger Logger) TracerOption {
	return func(tracer *Tracer) {
		tracer.logger = logger
	}
}

/*UseDualSpanMode sets the tracer in dual span mode*/
func (t TracerOptions) UseDualSpanMode() TracerOption {
	return func(tracer *Tracer) {
		tracer.useDualSpanMode = true
	}
}
