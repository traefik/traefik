package command

import (
	"fmt"
	"net"
	"regexp"
	"sort"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/command/base"
	"github.com/hashicorp/serf/serf"
	"github.com/ryanuber/columnize"
)

// MembersCommand is a Command implementation that queries a running
// Consul agent what members are part of the cluster currently.
type MembersCommand struct {
	base.Command
}

func (c *MembersCommand) Help() string {
	helpText := `
Usage: consul members [options]

  Outputs the members of a running Consul agent.

` + c.Command.Help()

	return strings.TrimSpace(helpText)
}

func (c *MembersCommand) Run(args []string) int {
	var detailed bool
	var wan bool
	var statusFilter string

	f := c.Command.NewFlagSet(c)
	f.BoolVar(&detailed, "detailed", false,
		"Provides detailed information about nodes.")
	f.BoolVar(&wan, "wan", false,
		"If the agent is in server mode, this can be used to return the other "+
			"peers in the WAN pool.")
	f.StringVar(&statusFilter, "status", ".*",
		"If provided, output is filtered to only nodes matching the regular "+
			"expression for status.")

	if err := c.Command.Parse(args); err != nil {
		return 1
	}

	// Compile the regexp
	statusRe, err := regexp.Compile(statusFilter)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to compile status regexp: %v", err))
		return 1
	}

	client, err := c.Command.HTTPClient()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error connecting to Consul agent: %s", err))
		return 1
	}

	members, err := client.Agent().Members(wan)
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error retrieving members: %s", err))
		return 1
	}

	// Filter the results
	n := len(members)
	for i := 0; i < n; i++ {
		member := members[i]
		statusString := serf.MemberStatus(member.Status).String()
		if !statusRe.MatchString(statusString) {
			members[i], members[n-1] = members[n-1], members[i]
			i--
			n--
			continue
		}
	}
	members = members[:n]

	// No matching members
	if len(members) == 0 {
		return 2
	}

	sort.Sort(ByMemberName(members))

	// Generate the output
	var result []string
	if detailed {
		result = c.detailedOutput(members)
	} else {
		result = c.standardOutput(members)
	}

	// Generate the columnized version
	output := columnize.SimpleFormat(result)
	c.Ui.Output(output)

	return 0
}

// so we can sort members by name
type ByMemberName []*consulapi.AgentMember

func (m ByMemberName) Len() int           { return len(m) }
func (m ByMemberName) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }
func (m ByMemberName) Less(i, j int) bool { return m[i].Name < m[j].Name }

// standardOutput is used to dump the most useful information about nodes
// in a more human-friendly format
func (c *MembersCommand) standardOutput(members []*consulapi.AgentMember) []string {
	result := make([]string, 0, len(members))
	header := "Node|Address|Status|Type|Build|Protocol|DC"
	result = append(result, header)
	for _, member := range members {
		addr := net.TCPAddr{IP: net.ParseIP(member.Addr), Port: int(member.Port)}
		protocol := member.Tags["vsn"]
		build := member.Tags["build"]
		if build == "" {
			build = "< 0.3"
		} else if idx := strings.Index(build, ":"); idx != -1 {
			build = build[:idx]
		}
		dc := member.Tags["dc"]

		statusString := serf.MemberStatus(member.Status).String()
		switch member.Tags["role"] {
		case "node":
			line := fmt.Sprintf("%s|%s|%s|client|%s|%s|%s",
				member.Name, addr.String(), statusString, build, protocol, dc)
			result = append(result, line)
		case "consul":
			line := fmt.Sprintf("%s|%s|%s|server|%s|%s|%s",
				member.Name, addr.String(), statusString, build, protocol, dc)
			result = append(result, line)
		default:
			line := fmt.Sprintf("%s|%s|%s|unknown|||",
				member.Name, addr.String(), statusString)
			result = append(result, line)
		}
	}
	return result
}

// detailedOutput is used to dump all known information about nodes in
// their raw format
func (c *MembersCommand) detailedOutput(members []*consulapi.AgentMember) []string {
	result := make([]string, 0, len(members))
	header := "Node|Address|Status|Tags"
	result = append(result, header)
	for _, member := range members {
		// Get the tags sorted by key
		tagKeys := make([]string, 0, len(member.Tags))
		for key := range member.Tags {
			tagKeys = append(tagKeys, key)
		}
		sort.Strings(tagKeys)

		// Format the tags as tag1=v1,tag2=v2,...
		var tagPairs []string
		for _, key := range tagKeys {
			tagPairs = append(tagPairs, fmt.Sprintf("%s=%s", key, member.Tags[key]))
		}

		tags := strings.Join(tagPairs, ",")

		addr := net.TCPAddr{IP: net.ParseIP(member.Addr), Port: int(member.Port)}
		line := fmt.Sprintf("%s|%s|%s|%s",
			member.Name, addr.String(), serf.MemberStatus(member.Status).String(), tags)
		result = append(result, line)
	}
	return result
}

func (c *MembersCommand) Synopsis() string {
	return "Lists the members of a Consul cluster"
}
