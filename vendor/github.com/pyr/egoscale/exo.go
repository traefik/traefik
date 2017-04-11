package main

import (
	"egoscale"
	"fmt"
	"flag"
	"os"
)

var apikey = flag.String("xk", "", "Exoscale API Key")
var apisecret = flag.String("xs", "", "Exoscale API Secret")
var endpoint = flag.String("xe", "https://api.exoscale.ch/compute", "Exoscale API Endpoint")

func main() {

	flag.Parse()
	client := egoscale.NewClient(*endpoint, *apikey, *apisecret)


	vms, err := client.ListVirtualMachines()
	if err != nil {
		fmt.Printf("got error: %s\n", err)
		os.Exit(1)
	}

	for _, vm := range(vms) {

		fmt.Println("vm:", vm.Displayname)
		for _, nic := range(vm.Nic) {
			fmt.Println("ip:", nic.Ipaddress)
		}
		for _, sg := range(vm.SecurityGroups) {
			fmt.Println("securitygroup:", sg.Name)
		}
	}
	os.Exit(0)


}
