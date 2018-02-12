/*

Package egoscale is a mapping for with the CloudStack API (http://cloudstack.apache.org/api.html) from Go. It has been designed against the Exoscale (https://www.exoscale.ch/) infrastructure but should fit other CloudStack services.

Requests and Responses

The paradigm used in this library is that CloudStack defines two types of requests synchronous (client.Request) and asynchronous (client.AsyncRequest). And when the expected responses is a success message, you may use the boolean requests variants (client.BooleanRequest, client.BooleanAsyncRequest). To build a request, construct the adequate struct. This library expects a pointer for efficiency reasons only. The response is a struct corresponding to the request itself. E.g. DeployVirtualMachine gives DeployVirtualMachineResponse, as a pointer as well to avoid big copies.

Then everything within the struct is not a pointer.

Affinity and Anti-Affinity groups

Affinity and Anti-Affinity groups provide a way to influence where VMs should run. See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#affinity-groups

APIs

All the available APIs on the server and provided by the API Discovery plugin

	cs := egoscale.NewClient("https://api.exoscale.ch/compute", "EXO...", "...")

	resp, err := cs.Request(&egoscale.ListAPIs{})
	if err != nil {
		panic(err)
	}

	for _, api := range resp.(*egoscale.ListAPIsResponse).API {
		fmt.Println("%s %s", api.Name, api.Description)
	}
	// Output:
	// listNetworks Lists all available networks
	// ...


Elastic IPs

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html#about-elastic-ips

Networks

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/4.8/networking_and_traffic.html

NICs

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/networking_and_traffic.html#configuring-multiple-ip-addresses-on-a-single-nic


Security Groups

Security Groups provide a way to isolate traffic to VMs.

	resp, err := cs.Request(&egoscale.CreateSecurityGroup{
		Name: "Load balancer",
		Description: "Opens HTTP/HTTPS ports from the outside world",
	})
	securityGroup := resp.(*egoscale.CreateSecurityGroupResponse).SecurityGroup
	// ...
	err = client.BooleanRequest(&egoscale.DeleteSecurityGroup{
		ID: securityGroup.ID,
	})
	// ...

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/networking_and_traffic.html#security-groups

Service Offerings

A service offering correspond to some hardware features (CPU, RAM).

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/service_offerings.html

SSH Key Pairs

In addition to username and password (disabled on Exoscale), SSH keys are used to log into the infrastructure.

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html#creating-the-ssh-keypair

Virtual Machines

... todo ...

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/stable/virtual_machines.html

Templates

... todo ...

See: http://docs.cloudstack.apache.org/projects/cloudstack-administration/en/latest/templates.html

Zones

A Zone corresponds to a Data Center.

*/
package egoscale
