package ext // import "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"

const (
	// AppTypeWeb specifies the Web span type and can be used as a tag value
	// for a span's SpanType tag.
	AppTypeWeb = "web"

	// AppTypeDB specifies the DB span type and can be used as a tag value
	// for a span's SpanType tag.
	AppTypeDB = "db"

	// AppTypeCache specifies the Cache span type and can be used as a tag value
	// for a span's SpanType tag.
	AppTypeCache = "cache"

	// AppTypeRPC specifies the RPC span type and can be used as a tag value
	// for a span's SpanType tag.
	AppTypeRPC = "rpc"
)
