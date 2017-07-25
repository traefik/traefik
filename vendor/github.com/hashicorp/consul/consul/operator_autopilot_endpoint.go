package consul

import (
	"fmt"

	"github.com/hashicorp/consul/consul/structs"
)

// AutopilotGetConfiguration is used to retrieve the current Autopilot configuration.
func (op *Operator) AutopilotGetConfiguration(args *structs.DCSpecificRequest, reply *structs.AutopilotConfig) error {
	if done, err := op.srv.forward("Operator.AutopilotGetConfiguration", args, args, reply); done {
		return err
	}

	// This action requires operator read access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorRead() {
		return permissionDeniedErr
	}

	state := op.srv.fsm.State()
	_, config, err := state.AutopilotConfig()
	if err != nil {
		return err
	}

	*reply = *config

	return nil
}

// AutopilotSetConfiguration is used to set the current Autopilot configuration.
func (op *Operator) AutopilotSetConfiguration(args *structs.AutopilotSetConfigRequest, reply *bool) error {
	if done, err := op.srv.forward("Operator.AutopilotSetConfiguration", args, args, reply); done {
		return err
	}

	// This action requires operator write access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorWrite() {
		return permissionDeniedErr
	}

	// Apply the update
	resp, err := op.srv.raftApply(structs.AutopilotRequestType, args)
	if err != nil {
		op.srv.logger.Printf("[ERR] consul.operator: Apply failed: %v", err)
		return err
	}
	if respErr, ok := resp.(error); ok {
		return respErr
	}

	// Check if the return type is a bool.
	if respBool, ok := resp.(bool); ok {
		*reply = respBool
	}
	return nil
}

// ServerHealth is used to get the current health of the servers.
func (op *Operator) ServerHealth(args *structs.DCSpecificRequest, reply *structs.OperatorHealthReply) error {
	// This must be sent to the leader, so we fix the args since we are
	// re-using a structure where we don't support all the options.
	args.RequireConsistent = true
	args.AllowStale = false
	if done, err := op.srv.forward("Operator.ServerHealth", args, args, reply); done {
		return err
	}

	// This action requires operator read access.
	acl, err := op.srv.resolveToken(args.Token)
	if err != nil {
		return err
	}
	if acl != nil && !acl.OperatorRead() {
		return permissionDeniedErr
	}

	// Exit early if the min Raft version is too low
	minRaftProtocol, err := ServerMinRaftProtocol(op.srv.LANMembers())
	if err != nil {
		return fmt.Errorf("error getting server raft protocol versions: %s", err)
	}
	if minRaftProtocol < 3 {
		return fmt.Errorf("all servers must have raft_protocol set to 3 or higher to use this endpoint")
	}

	*reply = op.srv.getClusterHealth()

	return nil
}
