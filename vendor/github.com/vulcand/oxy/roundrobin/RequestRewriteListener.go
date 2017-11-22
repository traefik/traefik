package roundrobin

import "net/http"

type RequestRewriteListener func(oldReq *http.Request, newReq *http.Request)
