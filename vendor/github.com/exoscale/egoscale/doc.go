/*

Package egoscale is a mapping for the Exoscale API (https://community.exoscale.com/api/compute/).

Requests and Responses

To build a request, construct the adequate struct. This library expects a pointer for efficiency reasons only. The response is a struct corresponding to the data at stake. E.g. DeployVirtualMachine gives a VirtualMachine, as a pointer as well to avoid big copies.

Then everything within the struct is not a pointer. Find below some examples of how egoscale may be used. If anything feels odd or unclear, please let us know: https://github.com/exoscale/egoscale/issues

	req := &egoscale.DeployVirtualMachine{
		Size:              10,
		ServiceOfferingID: egoscale.MustParseUUID("..."),
		TemplateID:        egoscale.MustParseUUID("..."),
		ZoneID:            egoscale.MastParseUUID("..."),
	}

	fmt.Println("Deployment started")
	resp, err := cs.Request(req)
	if err != nil {
		panic(err)
	}

	vm := resp.(*egoscale.VirtualMachine)
	fmt.Printf("Virtual Machine ID: %s\n", vm.ID)

This example deploys a virtual machine while controlling the job status as it goes. It enables a finer control over errors, e.g. HTTP timeout, and eventually a way to kill it of (from the client side).

	req := &egoscale.DeployVirtualMachine{
		Size:              10,
		ServiceOfferingID: egoscale.MustParseUUID("..."),
		TemplateID:        egoscale.MustParseUUID("..."),
		ZoneID:            egoscale.MustParseUUID("..."),
	}
	vm := &egoscale.VirtualMachine{}

	fmt.Println("Deployment started")
	cs.AsyncRequest(req, func(jobResult *egoscale.AsyncJobResult, err error) bool {
		if err != nil {
			// any kind of error
			panic(err)
		}

		// Keep waiting
		if jobResult.JobStatus == egoscale.Pending {
			fmt.Println("wait...")
			return true
		}

		// Unmarshal the response into the response struct
		if err := jobResult.Response(vm); err != nil {
			// JSON unmarshaling error
			panic(err)
		}

		// Stop waiting
		return false
	})

	fmt.Printf("Virtual Machine ID: %s\n", vm.ID)

Debugging and traces

As this library is mostly an HTTP client, you can reuse all the existing tools around it.

	cs := egoscale.NewClient("https://api.exoscale.com/compute", "EXO...", "...")
	// sets a logger on stderr
	cs.Logger = log.New(os.Stderr, "prefix", log.LstdFlags)
	// activates the HTTP traces
	cs.TraceOn()

Nota bene: when running the tests or the egoscale library via another tool, e.g. the exo cli, the environment variable EXOSCALE_TRACE=prefix does the above configuration for you. As a developer using egoscale as a library, you'll find it more convenient to plug your favorite io.Writer as it's a Logger.


APIs

All the available APIs on the server and provided by the API Discovery plugin.

	cs := egoscale.NewClient("https://api.exoscale.com/compute", "EXO...", "...")

	resp, err := cs.Request(&egoscale.ListAPIs{})
	if err != nil {
		panic(err)
	}

	for _, api := range resp.(*egoscale.ListAPIsResponse).API {
		fmt.Printf("%s %s\n", api.Name, api.Description)
	}
	// Output:
	// listNetworks Lists all available networks
	// ...

Security Groups

Security Groups provide a way to isolate traffic to VMs. Rules are added via the two Authorization commands.

	resp, err := cs.Request(&egoscale.CreateSecurityGroup{
		Name: "Load balancer",
		Description: "Open HTTP/HTTPS ports from the outside world",
	})
	securityGroup := resp.(*egoscale.SecurityGroup)

	resp, err = cs.Request(&egoscale.AuthorizeSecurityGroupIngress{
		Description:     "SSH traffic",
		SecurityGroupID: securityGroup.ID,
		CidrList:        []CIDR{
			*egoscale.MustParseCIDR("0.0.0.0/0"),
			*egoscale.MustParseCIDR("::/0"),
		},
		Protocol:        "tcp",
		StartPort:       22,
		EndPort:         22,
	})
	// The modified SecurityGroup is returned
	securityGroup := resp.(*egoscale.SecurityGroup)

	// ...
	err = client.BooleanRequest(&egoscale.DeleteSecurityGroup{
		ID: securityGroup.ID,
	})
	// ...

Security Group also implement the generic List, Get and Delete interfaces (Listable and Deletable).

	// List all Security Groups
	sgs, _ := cs.List(&egoscale.SecurityGroup{})
	for _, s := range sgs {
		sg := s.(egoscale.SecurityGroup)
		// ...
	}

	// Get a Security Group
	sgQuery := &egoscale.SecurityGroup{Name: "Load balancer"}
	resp, err := cs.Get(sgQuery); err != nil {
		...
	}
	sg := resp.(*egoscale.SecurityGroup)

	if err := cs.Delete(sg); err != nil {
		...
	}
	// The SecurityGroup has been deleted

See: https://community.exoscale.com/documentation/compute/security-groups/

Zones

A Zone corresponds to a Data Center. You may list them. Zone implements the Listable interface, which let you perform a list in two different ways. The first exposes the underlying request while the second one hide them and you only manipulate the structs of your interest.

	// Using ListZones request
	req := &egoscale.ListZones{}
	resp, err := client.Request(req)
	if err != nil {
		panic(err)
	}

	for _, zone := range resp.(*egoscale.ListZonesResponse) {
		...
	}

	// Using client.List
	zone := &egoscale.Zone{}
	zones, err := client.List(zone)
	if err != nil {
		panic(err)
	}

	for _, z := range zones {
		zone := z.(egoscale.Zone)
		...
	}

Elastic IPs

An Elastic IP is a way to attach an IP address to many Virtual Machines. The API side of the story configures the external environment, like the routing. Some work is required within the machine to properly configure the interfaces.

See: https://community.exoscale.com/documentation/compute/eip/

*/
package egoscale
