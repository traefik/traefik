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

/*Logger defines a new logger interface*/
type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
	Debug(format string, v ...interface{})
}

/*NullLogger does nothing*/
type NullLogger struct{}

/*Error prints the error message*/
func (logger NullLogger) Error(format string, v ...interface{}) {}

/*Info prints the info message*/
func (logger NullLogger) Info(format string, v ...interface{}) {}

/*Debug prints the info message*/
func (logger NullLogger) Debug(format string, v ...interface{}) {}
