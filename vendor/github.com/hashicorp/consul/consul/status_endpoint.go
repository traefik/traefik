package consul

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/consul/consul/structs"
)

// Status endpoint is used to check on server status
type Status struct {
	server *Server
}

// Ping is used to just check for connectivity
func (s *Status) Ping(args struct{}, reply *struct{}) error {
	return nil
}

// Leader is used to get the address of the leader
func (s *Status) Leader(args struct{}, reply *string) error {
	leader := string(s.server.raft.Leader())
	if leader != "" {
		*reply = leader
	} else {
		*reply = ""
	}
	return nil
}

// Peers is used to get all the Raft peers
func (s *Status) Peers(args struct{}, reply *[]string) error {
	future := s.server.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return err
	}

	for _, server := range future.Configuration().Servers {
		*reply = append(*reply, string(server.Address))
	}
	return nil
}

// Used by Autopilot to query the raft stats of the local server.
func (s *Status) RaftStats(args struct{}, reply *structs.ServerStats) error {
	stats := s.server.raft.Stats()

	var err error
	reply.LastContact = stats["last_contact"]
	reply.LastIndex, err = strconv.ParseUint(stats["last_log_index"], 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing server's last_log_index value: %s", err)
	}
	reply.LastTerm, err = strconv.ParseUint(stats["last_log_term"], 10, 64)
	if err != nil {
		return fmt.Errorf("error parsing server's last_log_term value: %s", err)
	}

	return nil
}
