package instana

import (
	ot "github.com/opentracing/opentracing-go"
)

type jsonSpan struct {
	TraceID   int64     `json:"t"`
	ParentID  *int64    `json:"p,omitempty"`
	SpanID    int64     `json:"s"`
	Timestamp uint64    `json:"ts"`
	Duration  uint64    `json:"d"`
	Name      string    `json:"n"`
	From      *fromS    `json:"f"`
	Error     bool      `json:"error"`
	Ec        int       `json:"ec,omitempty"`
	Lang      string    `json:"ta,omitempty"`
	Data      *jsonData `json:"data"`
}

type jsonData struct {
	Service string       `json:"service,omitempty"`
	SDK     *jsonSDKData `json:"sdk"`
}

type jsonCustomData struct {
	Tags    ot.Tags                           `json:"tags,omitempty"`
	Logs    map[uint64]map[string]interface{} `json:"logs,omitempty"`
	Baggage map[string]string                 `json:"baggage,omitempty"`
}

type jsonSDKData struct {
	Name      string          `json:"name"`
	Type      string          `json:"type,omitempty"`
	Arguments string          `json:"arguments,omitempty"`
	Return    string          `json:"return,omitempty"`
	Custom    *jsonCustomData `json:"custom,omitempty"`
}
