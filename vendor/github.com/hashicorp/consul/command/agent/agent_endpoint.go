package agent

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/consul/consul/structs"
	"github.com/hashicorp/consul/logger"
	"github.com/hashicorp/consul/types"
	"github.com/hashicorp/logutils"
	"github.com/hashicorp/serf/coordinate"
	"github.com/hashicorp/serf/serf"
)

type AgentSelf struct {
	Config *Config
	Coord  *coordinate.Coordinate
	Member serf.Member
	Stats  map[string]map[string]string
	Meta   map[string]string
}

func (s *HTTPServer) AgentSelf(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	var c *coordinate.Coordinate
	if !s.agent.config.DisableCoordinates {
		var err error
		if c, err = s.agent.GetCoordinate(); err != nil {
			return nil, err
		}
	}

	// Fetch the ACL token, if any, and enforce agent policy.
	var token string
	s.parseToken(req, &token)
	acl, err := s.agent.resolveToken(token)
	if err != nil {
		return nil, err
	}
	if acl != nil && !acl.AgentRead(s.agent.config.NodeName) {
		return nil, permissionDeniedErr
	}

	return AgentSelf{
		Config: s.agent.config,
		Coord:  c,
		Member: s.agent.LocalMember(),
		Stats:  s.agent.Stats(),
		Meta:   s.agent.state.Metadata(),
	}, nil
}

func (s *HTTPServer) AgentReload(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req.Method != "PUT" {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return nil, nil
	}

	// Fetch the ACL token, if any, and enforce agent policy.
	var token string
	s.parseToken(req, &token)
	acl, err := s.agent.resolveToken(token)
	if err != nil {
		return nil, err
	}
	if acl != nil && !acl.AgentWrite(s.agent.config.NodeName) {
		return nil, permissionDeniedErr
	}

	// Trigger the reload
	errCh := make(chan error, 0)
	select {
	case <-s.agent.ShutdownCh():
		return nil, fmt.Errorf("Agent was shutdown before reload could be completed")
	case s.agent.reloadCh <- errCh:
	}

	// Wait for the result of the reload, or for the agent to shutdown
	select {
	case <-s.agent.ShutdownCh():
		return nil, fmt.Errorf("Agent was shutdown before reload could be completed")
	case err := <-errCh:
		return nil, err
	}
}

func (s *HTTPServer) AgentServices(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Fetch the ACL token, if any.
	var token string
	s.parseToken(req, &token)

	services := s.agent.state.Services()
	if err := s.agent.filterServices(token, &services); err != nil {
		return nil, err
	}
	return services, nil
}

func (s *HTTPServer) AgentChecks(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Fetch the ACL token, if any.
	var token string
	s.parseToken(req, &token)

	checks := s.agent.state.Checks()
	if err := s.agent.filterChecks(token, &checks); err != nil {
		return nil, err
	}
	return checks, nil
}

func (s *HTTPServer) AgentMembers(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Fetch the ACL token, if any.
	var token string
	s.parseToken(req, &token)

	// Check if the WAN is being queried
	wan := false
	if other := req.URL.Query().Get("wan"); other != "" {
		wan = true
	}

	var members []serf.Member
	if wan {
		members = s.agent.WANMembers()
	} else {
		members = s.agent.LANMembers()
	}
	if err := s.agent.filterMembers(token, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (s *HTTPServer) AgentJoin(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Fetch the ACL token, if any, and enforce agent policy.
	var token string
	s.parseToken(req, &token)
	acl, err := s.agent.resolveToken(token)
	if err != nil {
		return nil, err
	}
	if acl != nil && !acl.AgentWrite(s.agent.config.NodeName) {
		return nil, permissionDeniedErr
	}

	// Check if the WAN is being queried
	wan := false
	if other := req.URL.Query().Get("wan"); other != "" {
		wan = true
	}

	// Get the address
	addr := strings.TrimPrefix(req.URL.Path, "/v1/agent/join/")
	if wan {
		_, err := s.agent.JoinWAN([]string{addr})
		return nil, err
	} else {
		_, err := s.agent.JoinLAN([]string{addr})
		return nil, err
	}
}

func (s *HTTPServer) AgentLeave(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req.Method != "PUT" {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return nil, nil
	}

	// Fetch the ACL token, if any, and enforce agent policy.
	var token string
	s.parseToken(req, &token)
	acl, err := s.agent.resolveToken(token)
	if err != nil {
		return nil, err
	}
	if acl != nil && !acl.AgentWrite(s.agent.config.NodeName) {
		return nil, permissionDeniedErr
	}

	if err := s.agent.Leave(); err != nil {
		return nil, err
	}
	return nil, s.agent.Shutdown()
}

func (s *HTTPServer) AgentForceLeave(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Fetch the ACL token, if any, and enforce agent policy.
	var token string
	s.parseToken(req, &token)
	acl, err := s.agent.resolveToken(token)
	if err != nil {
		return nil, err
	}
	if acl != nil && !acl.AgentWrite(s.agent.config.NodeName) {
		return nil, permissionDeniedErr
	}

	addr := strings.TrimPrefix(req.URL.Path, "/v1/agent/force-leave/")
	return nil, s.agent.ForceLeave(addr)
}

// syncChanges is a helper function which wraps a blocking call to sync
// services and checks to the server. If the operation fails, we only
// only warn because the write did succeed and anti-entropy will sync later.
func (s *HTTPServer) syncChanges() {
	if err := s.agent.state.syncChanges(); err != nil {
		s.logger.Printf("[ERR] agent: failed to sync changes: %v", err)
	}
}

const invalidCheckMessage = "Must provide TTL or Script/DockerContainerID/HTTP/TCP and Interval"

func (s *HTTPServer) AgentRegisterCheck(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	var args CheckDefinition
	// Fixup the type decode of TTL or Interval.
	decodeCB := func(raw interface{}) error {
		return FixupCheckType(raw)
	}
	if err := decodeBody(req, &args, decodeCB); err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf("Request decode failed: %v", err)))
		return nil, nil
	}

	// Verify the check has a name.
	if args.Name == "" {
		resp.WriteHeader(400)
		resp.Write([]byte("Missing check name"))
		return nil, nil
	}

	if args.Status != "" && !structs.ValidStatus(args.Status) {
		resp.WriteHeader(400)
		resp.Write([]byte("Bad check status"))
		return nil, nil
	}

	// Construct the health check.
	health := args.HealthCheck(s.agent.config.NodeName)

	// Verify the check type.
	chkType := &args.CheckType
	if !chkType.Valid() {
		resp.WriteHeader(400)
		resp.Write([]byte(invalidCheckMessage))
		return nil, nil
	}

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetCheckRegister(token, health); err != nil {
		return nil, err
	}

	// Add the check.
	if err := s.agent.AddCheck(health, chkType, true, token); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentDeregisterCheck(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	checkID := types.CheckID(strings.TrimPrefix(req.URL.Path, "/v1/agent/check/deregister/"))

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetCheckUpdate(token, checkID); err != nil {
		return nil, err
	}

	if err := s.agent.RemoveCheck(checkID, true); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentCheckPass(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	checkID := types.CheckID(strings.TrimPrefix(req.URL.Path, "/v1/agent/check/pass/"))
	note := req.URL.Query().Get("note")

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetCheckUpdate(token, checkID); err != nil {
		return nil, err
	}

	if err := s.agent.updateTTLCheck(checkID, structs.HealthPassing, note); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentCheckWarn(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	checkID := types.CheckID(strings.TrimPrefix(req.URL.Path, "/v1/agent/check/warn/"))
	note := req.URL.Query().Get("note")

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetCheckUpdate(token, checkID); err != nil {
		return nil, err
	}

	if err := s.agent.updateTTLCheck(checkID, structs.HealthWarning, note); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentCheckFail(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	checkID := types.CheckID(strings.TrimPrefix(req.URL.Path, "/v1/agent/check/fail/"))
	note := req.URL.Query().Get("note")

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetCheckUpdate(token, checkID); err != nil {
		return nil, err
	}

	if err := s.agent.updateTTLCheck(checkID, structs.HealthCritical, note); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

// checkUpdate is the payload for a PUT to AgentCheckUpdate.
type checkUpdate struct {
	// Status us one of the structs.Health* states, "passing", "warning", or
	// "critical".
	Status string

	// Output is the information to post to the UI for operators as the
	// output of the process that decided to hit the TTL check. This is
	// different from the note field that's associated with the check
	// itself.
	Output string
}

// AgentCheckUpdate is a PUT-based alternative to the GET-based Pass/Warn/Fail
// APIs.
func (s *HTTPServer) AgentCheckUpdate(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	if req.Method != "PUT" {
		resp.WriteHeader(405)
		return nil, nil
	}

	var update checkUpdate
	if err := decodeBody(req, &update, nil); err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf("Request decode failed: %v", err)))
		return nil, nil
	}

	switch update.Status {
	case structs.HealthPassing:
	case structs.HealthWarning:
	case structs.HealthCritical:
	default:
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf("Invalid check status: '%s'", update.Status)))
		return nil, nil
	}

	total := len(update.Output)
	if total > CheckBufSize {
		update.Output = fmt.Sprintf("%s ... (captured %d of %d bytes)",
			update.Output[:CheckBufSize], CheckBufSize, total)
	}

	checkID := types.CheckID(strings.TrimPrefix(req.URL.Path, "/v1/agent/check/update/"))

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetCheckUpdate(token, checkID); err != nil {
		return nil, err
	}

	if err := s.agent.updateTTLCheck(checkID, update.Status, update.Output); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentRegisterService(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	var args ServiceDefinition
	// Fixup the type decode of TTL or Interval if a check if provided.
	decodeCB := func(raw interface{}) error {
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			return nil
		}

		for k, v := range rawMap {
			switch strings.ToLower(k) {
			case "check":
				if err := FixupCheckType(v); err != nil {
					return err
				}
			case "checks":
				chkTypes, ok := v.([]interface{})
				if !ok {
					continue
				}
				for _, chkType := range chkTypes {
					if err := FixupCheckType(chkType); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	if err := decodeBody(req, &args, decodeCB); err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf("Request decode failed: %v", err)))
		return nil, nil
	}

	// Verify the service has a name.
	if args.Name == "" {
		resp.WriteHeader(400)
		resp.Write([]byte("Missing service name"))
		return nil, nil
	}

	// Get the node service.
	ns := args.NodeService()

	// Verify the check type.
	chkTypes := args.CheckTypes()
	for _, check := range chkTypes {
		if check.Status != "" && !structs.ValidStatus(check.Status) {
			resp.WriteHeader(400)
			resp.Write([]byte("Status for checks must 'passing', 'warning', 'critical'"))
			return nil, nil
		}
		if !check.Valid() {
			resp.WriteHeader(400)
			resp.Write([]byte(invalidCheckMessage))
			return nil, nil
		}
	}

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetServiceRegister(token, ns); err != nil {
		return nil, err
	}

	// Add the service.
	if err := s.agent.AddService(ns, chkTypes, true, token); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentDeregisterService(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	serviceID := strings.TrimPrefix(req.URL.Path, "/v1/agent/service/deregister/")

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetServiceUpdate(token, serviceID); err != nil {
		return nil, err
	}

	if err := s.agent.RemoveService(serviceID, true); err != nil {
		return nil, err
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentServiceMaintenance(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Only PUT supported
	if req.Method != "PUT" {
		resp.WriteHeader(405)
		return nil, nil
	}

	// Ensure we have a service ID
	serviceID := strings.TrimPrefix(req.URL.Path, "/v1/agent/service/maintenance/")
	if serviceID == "" {
		resp.WriteHeader(400)
		resp.Write([]byte("Missing service ID"))
		return nil, nil
	}

	// Ensure we have some action
	params := req.URL.Query()
	if _, ok := params["enable"]; !ok {
		resp.WriteHeader(400)
		resp.Write([]byte("Missing value for enable"))
		return nil, nil
	}

	raw := params.Get("enable")
	enable, err := strconv.ParseBool(raw)
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf("Invalid value for enable: %q", raw)))
		return nil, nil
	}

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	if err := s.agent.vetServiceUpdate(token, serviceID); err != nil {
		return nil, err
	}

	if enable {
		reason := params.Get("reason")
		if err = s.agent.EnableServiceMaintenance(serviceID, reason, token); err != nil {
			resp.WriteHeader(404)
			resp.Write([]byte(err.Error()))
			return nil, nil
		}
	} else {
		if err = s.agent.DisableServiceMaintenance(serviceID); err != nil {
			resp.WriteHeader(404)
			resp.Write([]byte(err.Error()))
			return nil, nil
		}
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentNodeMaintenance(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Only PUT supported
	if req.Method != "PUT" {
		resp.WriteHeader(405)
		return nil, nil
	}

	// Ensure we have some action
	params := req.URL.Query()
	if _, ok := params["enable"]; !ok {
		resp.WriteHeader(400)
		resp.Write([]byte("Missing value for enable"))
		return nil, nil
	}

	raw := params.Get("enable")
	enable, err := strconv.ParseBool(raw)
	if err != nil {
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf("Invalid value for enable: %q", raw)))
		return nil, nil
	}

	// Get the provided token, if any, and vet against any ACL policies.
	var token string
	s.parseToken(req, &token)
	acl, err := s.agent.resolveToken(token)
	if err != nil {
		return nil, err
	}
	if acl != nil && !acl.NodeWrite(s.agent.config.NodeName) {
		return nil, permissionDeniedErr
	}

	if enable {
		s.agent.EnableNodeMaintenance(params.Get("reason"), token)
	} else {
		s.agent.DisableNodeMaintenance()
	}
	s.syncChanges()
	return nil, nil
}

func (s *HTTPServer) AgentMonitor(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	// Only GET supported.
	if req.Method != "GET" {
		resp.WriteHeader(405)
		return nil, nil
	}

	// Fetch the ACL token, if any, and enforce agent policy.
	var token string
	s.parseToken(req, &token)
	acl, err := s.agent.resolveToken(token)
	if err != nil {
		return nil, err
	}
	if acl != nil && !acl.AgentRead(s.agent.config.NodeName) {
		return nil, permissionDeniedErr
	}

	// Get the provided loglevel.
	logLevel := req.URL.Query().Get("loglevel")
	if logLevel == "" {
		logLevel = "INFO"
	}

	// Upper case the level since that's required by the filter.
	logLevel = strings.ToUpper(logLevel)

	// Create a level filter and flusher.
	filter := logger.LevelFilter()
	filter.MinLevel = logutils.LogLevel(logLevel)
	if !logger.ValidateLevelFilter(filter.MinLevel, filter) {
		resp.WriteHeader(400)
		resp.Write([]byte(fmt.Sprintf("Unknown log level: %s", filter.MinLevel)))
		return nil, nil
	}
	flusher, ok := resp.(http.Flusher)
	if !ok {
		return nil, fmt.Errorf("Streaming not supported")
	}

	// Set up a log handler.
	handler := &httpLogHandler{
		filter: filter,
		logCh:  make(chan string, 512),
		logger: s.logger,
	}
	s.agent.logWriter.RegisterHandler(handler)
	defer s.agent.logWriter.DeregisterHandler(handler)
	notify := resp.(http.CloseNotifier).CloseNotify()

	// Stream logs until the connection is closed.
	for {
		select {
		case <-notify:
			s.agent.logWriter.DeregisterHandler(handler)
			if handler.droppedCount > 0 {
				s.agent.logger.Printf("[WARN] agent: Dropped %d logs during monitor request", handler.droppedCount)
			}
			return nil, nil
		case log := <-handler.logCh:
			resp.Write([]byte(log + "\n"))
			flusher.Flush()
		}
	}
}

type httpLogHandler struct {
	filter       *logutils.LevelFilter
	logCh        chan string
	logger       *log.Logger
	droppedCount int
}

func (h *httpLogHandler) HandleLog(log string) {
	// Check the log level
	if !h.filter.Check([]byte(log)) {
		return
	}

	// Do a non-blocking send
	select {
	case h.logCh <- log:
	default:
		// Just increment a counter for dropped logs to this handler; we can't log now
		// because the lock is already held by the LogWriter invoking this
		h.droppedCount += 1
	}
}
