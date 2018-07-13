package roundrobin

import "net/http"

// RequestRewriteListener function to rewrite request
type RequestRewriteListener func(oldReq *http.Request, newReq *http.Request)
