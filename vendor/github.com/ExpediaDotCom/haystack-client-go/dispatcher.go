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
	"os"
	"os/signal"
	"time"

	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
)

/*Dispatcher dispatches the span object*/
type Dispatcher interface {
	Name() string
	Dispatch(span *_Span)
	DispatchProtoSpan(span *Span)
	Close()
	SetLogger(logger Logger)
}

/*InMemoryDispatcher implements the Dispatcher interface*/
type InMemoryDispatcher struct {
	spans  []*_Span
	logger Logger
}

/*NewInMemoryDispatcher creates a new in memory dispatcher*/
func NewInMemoryDispatcher() Dispatcher {
	return &InMemoryDispatcher{}
}

/*Name gives the Dispatcher name*/
func (d *InMemoryDispatcher) Name() string {
	return "InMemoryDispatcher"
}

/*SetLogger sets the logger to use*/
func (d *InMemoryDispatcher) SetLogger(logger Logger) {
	d.logger = logger
}

/*Dispatch dispatches the span object*/
func (d *InMemoryDispatcher) Dispatch(span *_Span) {
	d.spans = append(d.spans, span)
}

/*DispatchProtoSpan dispatches proto span object*/
func (d *InMemoryDispatcher) DispatchProtoSpan(span *Span) {
	/* not implemented */
}

/*Close down the inMemory dispatcher*/
func (d *InMemoryDispatcher) Close() {
	d.spans = nil
}

/*FileDispatcher file dispatcher*/
type FileDispatcher struct {
	fileHandle *os.File
	logger     Logger
}

/*NewFileDispatcher creates a new file dispatcher*/
func NewFileDispatcher(filename string) Dispatcher {
	fd, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	return &FileDispatcher{
		fileHandle: fd,
	}
}

/*Name gives the Dispatcher name*/
func (d *FileDispatcher) Name() string {
	return "FileDispatcher"
}

/*SetLogger sets the logger to use*/
func (d *FileDispatcher) SetLogger(logger Logger) {
	d.logger = logger
}

/*Dispatch dispatches the span object*/
func (d *FileDispatcher) Dispatch(span *_Span) {
	_, err := d.fileHandle.WriteString(span.String() + "\n")
	if err != nil {
		panic(err)
	}
}

/*DispatchProtoSpan dispatches proto span object*/
func (d *FileDispatcher) DispatchProtoSpan(span *Span) {
	/* not implemented */
}

/*Close down the file dispatcher*/
func (d *FileDispatcher) Close() {
	err := d.fileHandle.Close()
	if err != nil {
		panic(err)
	}
}

/*RemoteDispatcher dispatcher, client can be grpc or http*/
type RemoteDispatcher struct {
	client      RemoteClient
	timeout     time.Duration
	logger      Logger
	spanChannel chan *Span
}

/*NewHTTPDispatcher creates a new haystack-agent dispatcher*/
func NewHTTPDispatcher(url string, timeout time.Duration, headers map[string]string, maxQueueLength int) Dispatcher {
	dispatcher := &RemoteDispatcher{
		client:      NewHTTPClient(url, headers, timeout),
		timeout:     timeout,
		spanChannel: make(chan *Span, maxQueueLength),
	}

	go startListener(dispatcher)
	return dispatcher
}

/*NewDefaultHTTPDispatcher creates a new http dispatcher*/
func NewDefaultHTTPDispatcher() Dispatcher {
	return NewHTTPDispatcher("http://haystack-collector/span", 3*time.Second, make(map[string](string)), 1000)
}

/*NewAgentDispatcher creates a new haystack-agent dispatcher*/
func NewAgentDispatcher(host string, port int, timeout time.Duration, maxQueueLength int) Dispatcher {
	dispatcher := &RemoteDispatcher{
		client:      NewGrpcClient(host, port, timeout),
		timeout:     timeout,
		spanChannel: make(chan *Span, maxQueueLength),
	}

	go startListener(dispatcher)
	return dispatcher
}

/*NewDefaultAgentDispatcher creates a new haystack-agent dispatcher*/
func NewDefaultAgentDispatcher() Dispatcher {
	return NewAgentDispatcher("haystack-agent", 35000, 3*time.Second, 1000)
}

func startListener(dispatcher *RemoteDispatcher) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)

	for {
		select {
		case sp := <-dispatcher.spanChannel:
			dispatcher.client.Send(sp)
		case <-signals:
			break
		}
	}
}

/*Name gives the Dispatcher name*/
func (d *RemoteDispatcher) Name() string {
	return "RemoteDispatcher"
}

/*SetLogger sets the logger to use*/
func (d *RemoteDispatcher) SetLogger(logger Logger) {
	d.logger = logger
	d.client.SetLogger(logger)
}

/*Dispatch dispatches the span object*/
func (d *RemoteDispatcher) Dispatch(span *_Span) {
	s := &Span{
		TraceId:       span.context.TraceID,
		SpanId:        span.context.SpanID,
		ParentSpanId:  span.context.ParentID,
		ServiceName:   span.ServiceName(),
		OperationName: span.OperationName(),
		StartTime:     span.startTime.UnixNano() / int64(time.Microsecond),
		Duration:      span.duration.Nanoseconds() / int64(time.Microsecond),
		Tags:          d.tags(span),
		Logs:          d.logs(span),
	}
	d.spanChannel <- s
}

/*DispatchProtoSpan dispatches the proto span object*/
func (d *RemoteDispatcher) DispatchProtoSpan(s *Span) {
	d.spanChannel <- s
}

func (d *RemoteDispatcher) logs(span *_Span) []*Log {
	var spanLogs []*Log
	for _, lg := range span.logs {
		spanLogs = append(spanLogs, &Log{
			Timestamp: lg.Timestamp.UnixNano() / int64(time.Microsecond),
			Fields:    d.logFieldsToTags(lg.Fields),
		})
	}
	return spanLogs
}

func (d *RemoteDispatcher) logFieldsToTags(fields []log.Field) []*Tag {
	var spanTags []*Tag
	for _, field := range fields {
		spanTags = append(spanTags, ConvertToProtoTag(field.Key(), field.Value()))
	}
	return spanTags
}

func (d *RemoteDispatcher) tags(span *_Span) []*Tag {
	var spanTags []*Tag
	for _, tag := range span.tags {
		spanTags = append(spanTags, ConvertToProtoTag(tag.Key, tag.Value))
	}
	return spanTags
}

/*Close down the file dispatcher*/
func (d *RemoteDispatcher) Close() {
	err := d.client.Close()
	if err != nil {
		d.logger.Error("Fail to close the haystack-agent dispatcher %v", err)
	}
}

/*ConvertToProtoTag converts to proto tag*/
func ConvertToProtoTag(key string, value interface{}) *Tag {
	switch v := value.(type) {
	case string:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VStr{
				VStr: value.(string),
			},
			Type: Tag_STRING,
		}
	case int:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VLong{
				VLong: int64(value.(int)),
			},
			Type: Tag_LONG,
		}
	case int32:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VLong{
				VLong: int64(value.(int32)),
			},
			Type: Tag_LONG,
		}
	case int16:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VLong{
				VLong: int64(value.(int16)),
			},
			Type: Tag_LONG,
		}
	case int64:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VLong{
				VLong: value.(int64),
			},
			Type: Tag_LONG,
		}
	case uint16:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VLong{
				VLong: int64(value.(uint16)),
			},
			Type: Tag_LONG,
		}
	case uint32:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VLong{
				VLong: int64(value.(uint32)),
			},
			Type: Tag_LONG,
		}
	case uint64:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VLong{
				VLong: int64(value.(uint64)),
			},
			Type: Tag_LONG,
		}
	case float32:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VDouble{
				VDouble: float64(value.(float32)),
			},
			Type: Tag_DOUBLE,
		}
	case float64:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VDouble{
				VDouble: value.(float64),
			},
			Type: Tag_DOUBLE,
		}
	case bool:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VBool{
				VBool: value.(bool),
			},
			Type: Tag_BOOL,
		}
	case []byte:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VBytes{
				VBytes: value.([]byte),
			},
			Type: Tag_BINARY,
		}
	case ext.SpanKindEnum:
		return &Tag{
			Key: key,
			Myvalue: &Tag_VStr{
				VStr: string(value.(ext.SpanKindEnum)),
			},
			Type: Tag_STRING,
		}
	default:
		panic(fmt.Errorf("unknown format %v", v))
	}
}
