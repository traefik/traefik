package agent

import (
	"bytes"
	"net/http"

	"github.com/hashicorp/consul/consul/structs"
)

// Snapshot handles requests to take and restore snapshots. This uses a special
// mechanism to make the RPC since we potentially stream large amounts of data
// as part of these requests.
func (s *HTTPServer) Snapshot(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	var args structs.SnapshotRequest
	s.parseDC(req, &args.Datacenter)
	s.parseToken(req, &args.Token)
	if _, ok := req.URL.Query()["stale"]; ok {
		args.AllowStale = true
	}

	switch req.Method {
	case "GET":
		args.Op = structs.SnapshotSave

		// Headers need to go out before we stream the body.
		replyFn := func(reply *structs.SnapshotResponse) error {
			setMeta(resp, &reply.QueryMeta)
			return nil
		}

		// Don't bother sending any request body through since it will
		// be ignored.
		var null bytes.Buffer
		if err := s.agent.SnapshotRPC(&args, &null, resp, replyFn); err != nil {
			return nil, err
		}
		return nil, nil

	case "PUT":
		args.Op = structs.SnapshotRestore
		if err := s.agent.SnapshotRPC(&args, req.Body, resp, nil); err != nil {
			return nil, err
		}
		return nil, nil

	default:
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return nil, nil
	}
}
