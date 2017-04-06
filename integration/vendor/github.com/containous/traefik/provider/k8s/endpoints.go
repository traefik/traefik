package k8s

// Endpoints is a collection of endpoints that implement the actual service.  Example:
//   Name: "mysvc",
//   Subsets: [
//     {
//       Addresses: [{"ip": "10.10.1.1"}, {"ip": "10.10.2.2"}],
//       Ports: [{"name": "a", "port": 8675}, {"name": "b", "port": 309}]
//     },
//     {
//       Addresses: [{"ip": "10.10.3.3"}],
//       Ports: [{"name": "a", "port": 93}, {"name": "b", "port": 76}]
//     },
//  ]
type Endpoints struct {
	TypeMeta   `json:",inline"`
	ObjectMeta `json:"metadata,omitempty"`

	// The set of all endpoints is the union of all subsets.
	Subsets []EndpointSubset
}

// EndpointSubset is a group of addresses with a common set of ports.  The
// expanded set of endpoints is the Cartesian product of Addresses x Ports.
// For example, given:
//   {
//     Addresses: [{"ip": "10.10.1.1"}, {"ip": "10.10.2.2"}],
//     Ports:     [{"name": "a", "port": 8675}, {"name": "b", "port": 309}]
//   }
// The resulting set of endpoints can be viewed as:
//     a: [ 10.10.1.1:8675, 10.10.2.2:8675 ],
//     b: [ 10.10.1.1:309, 10.10.2.2:309 ]
type EndpointSubset struct {
	Addresses         []EndpointAddress
	NotReadyAddresses []EndpointAddress
	Ports             []EndpointPort
}

// EndpointAddress is a tuple that describes single IP address.
type EndpointAddress struct {
	// The IP of this endpoint.
	// IPv6 is also accepted but not fully supported on all platforms. Also, certain
	// kubernetes components, like kube-proxy, are not IPv6 ready.
	// TODO: This should allow hostname or IP, see #4447.
	IP string
	// Optional: Hostname of this endpoint
	// Meant to be used by DNS servers etc.
	Hostname string `json:"hostname,omitempty"`
	// Optional: The kubernetes object related to the entry point.
	TargetRef *ObjectReference
}

// EndpointPort is a tuple that describes a single port.
type EndpointPort struct {
	// The name of this port (corresponds to ServicePort.Name).  Optional
	// if only one port is defined.  Must be a DNS_LABEL.
	Name string

	// The port number.
	Port int32

	// The IP protocol for this port.
	Protocol Protocol
}

// ObjectReference contains enough information to let you inspect or modify the referred object.
type ObjectReference struct {
	Kind            string `json:"kind,omitempty"`
	Namespace       string `json:"namespace,omitempty"`
	Name            string `json:"name,omitempty"`
	UID             UID    `json:"uid,omitempty"`
	APIVersion      string `json:"apiVersion,omitempty"`
	ResourceVersion string `json:"resourceVersion,omitempty"`

	// Optional. If referring to a piece of an object instead of an entire object, this string
	// should contain information to identify the sub-object. For example, if the object
	// reference is to a container within a pod, this would take on a value like:
	// "spec.containers{name}" (where "name" refers to the name of the container that triggered
	// the event) or if no container name is specified "spec.containers[2]" (container with
	// index 2 in this pod). This syntax is chosen only to have some well-defined way of
	// referencing a part of an object.
	// TODO: this design is not final and this field is subject to change in the future.
	FieldPath string `json:"fieldPath,omitempty"`
}
