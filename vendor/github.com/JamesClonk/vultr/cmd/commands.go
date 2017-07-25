package cmd

import (
	"fmt"
	"runtime"

	vultr "github.com/JamesClonk/vultr/lib"
	"github.com/jawher/mow.cli"
)

// RegisterCommands registers all CLI commands
func (c *CLI) RegisterCommands() {
	// dns
	c.Command("dns", "modify DNS", func(cmd *cli.Cmd) {
		cmd.Command("domain", "show and change DNS domains", func(cmd *cli.Cmd) {
			cmd.Command("create", "create a DNS domain", dnsDomainCreate)
			cmd.Command("delete", "delete a DNS domain", dnsDomainDelete)
			cmd.Command("list", "list all DNS domains", dnsDomainList)
		})
		cmd.Command("record", "show and change DNS records", func(cmd *cli.Cmd) {
			cmd.Command("create", "create a DNS record", dnsRecordCreate)
			cmd.Command("update", "update a DNS record", dnsRecordUpdate)
			cmd.Command("delete", "delete a DNS record", dnsRecordDelete)
			cmd.Command("list", "list all DNS records", dnsRecordList)
		})
	})

	// info
	c.Command("info", "display account information", accountInfo)

	// iso
	c.Command("iso", "list all ISOs currently available on account", isoList)

	// os
	c.Command("os", "list all available operating systems", osList)

	// applications
	c.Command("apps", "list all available applications", appList)

	// plans
	c.Command("plans", "list all active plans", planList)

	// regions
	c.Command("regions", "list all active regions", regionList)

	// sshkeys
	c.Command("sshkey", "modify SSH public keys", func(cmd *cli.Cmd) {
		cmd.Command("create", "upload and add new SSH public key", sshKeysCreate)
		cmd.Command("update", "update an existing SSH public key", sshKeysUpdate)
		cmd.Command("delete", "remove an existing SSH public key", sshKeysDelete)
		cmd.Command("list", "list all existing SSH public keys", sshKeysList)
	})
	c.Command("sshkeys", "list all existing SSH public keys", sshKeysList)

	// ssh
	c.Command("ssh", "ssh into a virtual machine", sshServer)

	// servers
	c.Command("server", "modify virtual machines", func(cmd *cli.Cmd) {
		cmd.Command("create", "create a new virtual machine", serversCreate)
		cmd.Command("rename", "rename a virtual machine", serversRename)
		cmd.Command("start", "start a virtual machine (restart if already running)", serversStart)
		cmd.Command("halt", "halt a virtual machine (hard power off)", serversHalt)
		cmd.Command("reboot", "reboot a virtual machine (hard reboot)", serversReboot)
		cmd.Command("reinstall", "reinstall OS on a virtual machine (all data will be lost)", serversReinstall)
		cmd.Command("os", "show and change OS on a virtual machine", func(cmd *cli.Cmd) {
			cmd.Command("change", "change operating system of virtual machine (all data will be lost)", serversChangeOS)
			cmd.Command("list", "show a list of operating systems to which can be changed to", serversListOS)
		})
		cmd.Command("app", "show and change application on a virtual machine", func(cmd *cli.Cmd) {
			cmd.Command("change", "change application of virtual machine (all data will be lost)", serversChangeApplication)
			cmd.Command("list", "show a list of available applications to which can be changed to", serversListApplications)
		})
		cmd.Command("iso", "attach/detach ISO of a virtual machine", func(cmd *cli.Cmd) {
			cmd.Command("attach", "attach ISO to a virtual machine (server will hard reboot)", serversAttachISO)
			cmd.Command("detach", "detach ISO from a virtual machine (server will hard reboot)", serversDetachISO)
			cmd.Command("status", "show status of ISO attached to a virtual machine", serversStatusISO)
		})
		cmd.Command("delete", "delete a virtual machine", serversDelete)
		cmd.Command("bandwidth", "list bandwidth used by a virtual machine", serversBandwidth)
		cmd.Command("list", "list all active or pending virtual machines on current account", serversList)
		cmd.Command("show", "show detailed information of a virtual machine", serversShow)
		// ip information
		cmd.Command("list-ipv4", "list IPv4 information of a virtual machine", ipv4List)
		cmd.Command("list-ipv6", "list IPv6 information of a virtual machine", ipv6List)
		// reverse dns
		cmd.Command("reverse-dns", "modify reverse DNS entries", func(cmd *cli.Cmd) {
			cmd.Command("default-ipv4", "reset IPv4 reverse DNS entry back to original setting", reverseIpv4Default)
			cmd.Command("set-ipv4", "set IPv4 reverse DNS entry", reverseIpv4Set)
			cmd.Command("set-ipv6", "set IPv6 reverse DNS entry", reverseIpv6Set)
			cmd.Command("delete-ipv6", "delete IPv6 reverse DNS entry", reverseIpv6Delete)
			cmd.Command("list-ipv6", "list IPv6 reverse DNS entries of a virtual machine", reverseIpv6List)
		})
	})
	c.Command("servers", "list all active or pending virtual machines on current account", serversList)

	// block storage
	c.Command("storage", "modify block storage", func(cmd *cli.Cmd) {
		cmd.Command("create", "create new block storage", blockStorageCreate)
		cmd.Command("resize", "resize existing block storage", blockStorageResize)
		cmd.Command("label", "rename existing block storage", blockStorageLabel)
		cmd.Command("attach", "attach block storage to virtual machine", blockStorageAttach)
		cmd.Command("detach", "detach block storage from virtual machine", blockStorageDetach)
		cmd.Command("delete", "remove block storage", blockStorageDelete)
		cmd.Command("list", "list all block storage", blockStorageList)
	})
	c.Command("storages", "list all block storage", blockStorageList)

	// snapshots
	c.Command("snapshot", "modify snapshots", func(cmd *cli.Cmd) {
		cmd.Command("create", "create a snapshot from an existing virtual machine", snapshotsCreate)
		cmd.Command("delete", "delete a snapshot", snapshotsDelete)
		cmd.Command("list", "list all snapshots on current account", snapshotsList)
	})
	c.Command("snapshots", "list all snapshots on current account", snapshotsList)

	// startup scripts
	c.Command("script", "modify startup scripts", func(cmd *cli.Cmd) {
		cmd.Command("create", "create a new startup script", scriptsCreate)
		cmd.Command("update", "update an existing startup script", scriptsUpdate)
		cmd.Command("delete", "remove an existing startup script", scriptsDelete)
		cmd.Command("list", "list all startup scripts on current account", scriptsList)
		cmd.Command("show", "show complete startup script", scriptsShow)
	})
	c.Command("scripts", "list all startup scripts on current account", scriptsList)

	// reserved ips
	c.Command("reservedip", "modify reserved IPs", func(cmd *cli.Cmd) {
		cmd.Command("attach", "attach reserved IP to an existing virtual machine", reservedIPAttach)
		cmd.Command("convert", "convert existing IP on a virtual machine to a reserved IP", reservedIPConvert)
		cmd.Command("create", "create new reserved IP", reservedIPCreate)
		cmd.Command("delete", "delete reserved IP from your account", reservedIPDestroy)
		cmd.Command("detach", "detach reserved IP from an existing virtual machine", reservedIPDetach)
		cmd.Command("list", "list all active reserved IPs on current account", reservedIPList)
	})
	c.Command("reservedips", "list all active reserved IPs on current account", reservedIPList)

	// version
	c.Command("version", "vultr CLI version", func(cmd *cli.Cmd) {
		cmd.Action = func() {
			lengths := []int{24, 48}
			tabsPrint(columns{"Client version:", vultr.Version}, lengths)
			tabsPrint(columns{"Vultr API endpoint:", vultr.DefaultEndpoint}, lengths)
			tabsPrint(columns{"Vultr API version:", vultr.APIVersion}, lengths)
			tabsPrint(columns{"OS/Arch (client):", fmt.Sprintf("%v/%v", runtime.GOOS, runtime.GOARCH)}, lengths)
			tabsPrint(columns{"Go version:", runtime.Version()}, lengths)
			tabsFlush()
		}
	})
}
