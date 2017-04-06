package main

import (
	"egoscale"
	"fmt"
	"os"
	"time"
)

func main() {

	endpoint := os.Getenv("EXOSCALE_ENDPOINT")
	apiKey := os.Getenv("EXOSCALE_API_KEY")
	apiSecret := os.Getenv("EXOSCALE_API_SECRET")
	client := egoscale.NewClient(endpoint, apiKey, apiSecret)

	topo, err := client.GetTopology()
	if err != nil {
		fmt.Printf("got error: %+v\n", err)
		return
	}

	rules := []egoscale.SecurityGroupRule{
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "TCP",
			Port:            22,
		},
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "TCP",
			Port:            2376,
		},
		{
			SecurityGroupId: "",
			Cidr:            "0.0.0.0/0",
			Protocol:        "ICMP",
			IcmpType:        8,
			IcmpCode:        0,
		},
	}

	sgid, present := topo.SecurityGroups["egoscale"]
	if !present {
		resp, err := client.CreateSecurityGroupWithRules("egoscale", rules, make([]egoscale.SecurityGroupRule, 0, 0))
		if err != nil {
			fmt.Printf("got error: %+v\n", err)
			return
		}
		sgid = resp.Id
	}

	agid, present := topo.AffinityGroups["egoscale"]
	if !present {
		//Affinity Group Create is an async call
		jobid, err := client.CreateAffinityGroup("egoscale")

		var resp *egoscale.QueryAsyncJobResultResponse
		for i := 0; i <= 10; i++ {
			resp, err = client.PollAsyncJob(jobid)
			if err != nil {
				fmt.Printf("got error: %+v\n", err)
				return
			}

			if resp.Jobstatus == 1 {
				break
			}
			time.Sleep(5 * time.Second)
		}
	}
	fmt.Printf("Affinity Group ID :%v\n", agid)

	profile := egoscale.MachineProfile{
		Template:        topo.Images["ubuntu-14.04"][10],
		ServiceOffering: topo.Profiles["large"],
		SecurityGroups:  []string{sgid},
		Keypair:         topo.Keypairs[0],
		AffinityGroups:  []string{"egoscale"},
		Userdata:        "#cloud-config\nmanage_etc_hosts: true\nfqdn: deployed-by-egoscale\n",
		Zone:            topo.Zones["ch-gva-2"],
		Name:            "deployed-by-egoscale",
	}

	jobid, err := client.CreateVirtualMachine(profile)

	if err != nil {
		fmt.Printf("got error: %+v\n", err)
		return
	}

	var resp *egoscale.QueryAsyncJobResultResponse

	for i := 0; i <= 10; i++ {
		resp, err = client.PollAsyncJob(jobid)
		if err != nil {
			fmt.Printf("got error: %+v\n", err)
			return
		}

		if resp.Jobstatus == 1 {
			break
		}
		time.Sleep(5 * time.Second)
	}

	vm, err := client.AsyncToVirtualMachine(*resp)

	if err != nil {
		fmt.Printf("got error: %+v\n", err)
	}

	fmt.Printf("new machine up and running at: %s\n", vm.Nic[0].Ipaddress)

}
