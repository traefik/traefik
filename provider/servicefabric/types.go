package servicefabric

import "encoding/xml"

// ApplicationsData encapsulates the response
// model for Applications in the Service
// Fabric API
type ApplicationsData struct {
	ContinuationToken *string           `json:"ContinuationToken"`
	Items             []ApplicationItem `json:"Items"`
}

// ApplicationItem encapsulates the nested
// model for items within the ApplicationData
// model
type ApplicationItem struct {
	HealthState string `json:"HealthState"`
	ID          string `json:"Id"`
	Name        string `json:"Name"`
	Parameters  []*struct {
		Key   string `json:"Key"`
		Value string `json:"Value"`
	} `json:"Parameters"`
	Status      string `json:"Status"`
	TypeName    string `json:"TypeName"`
	TypeVersion string `json:"TypeVersion"`
}

// ServicesData encapsulates the response
// model for Services in the Service
// Fabric API
type ServicesData struct {
	ContinuationToken *string       `json:"ContinuationToken"`
	Items             []ServiceItem `json:"Items"`
}

// ServiceItem encapsulates the service information
// returned for each service in the Services data model
type ServiceItem struct {
	HasPersistedState bool   `json:"HasPersistedState"`
	HealthState       string `json:"HealthState"`
	ID                string `json:"Id"`
	IsServiceGroup    bool   `json:"IsServiceGroup"`
	ManifestVersion   string `json:"ManifestVersion"`
	Name              string `json:"Name"`
	ServiceKind       string `json:"ServiceKind"`
	ServiceStatus     string `json:"ServiceStatus"`
	TypeName          string `json:"TypeName"`
}

// PartitionsData encapsulates the response
// model for Parititons in the Service
// Fabric API
type PartitionsData struct {
	ContinuationToken *string `json:"ContinuationToken"`
	Items             []*struct {
		CurrentConfigurationEpoch struct {
			ConfigurationVersion string `json:"ConfigurationVersion"`
			DataLossVersion      string `json:"DataLossVersion"`
		} `json:"CurrentConfigurationEpoch"`
		HealthState          string `json:"HealthState"`
		MinReplicaSetSize    int64  `json:"MinReplicaSetSize"`
		PartitionInformation struct {
			HighKey              string `json:"HighKey"`
			ID                   string `json:"Id"`
			LowKey               string `json:"LowKey"`
			ServicePartitionKind string `json:"ServicePartitionKind"`
		} `json:"PartitionInformation"`
		PartitionStatus      string `json:"PartitionStatus"`
		ServiceKind          string `json:"ServiceKind"`
		TargetReplicaSetSize int64  `json:"TargetReplicaSetSize"`
	} `json:"Items"`
}

// ReplicasData encapsulates the response
// model for Replicas in the Service
// Fabric API
type ReplicasData struct {
	ContinuationToken *string `json:"ContinuationToken"`
	Items             []*struct {
		Address                      string `json:"Address"`
		HealthState                  string `json:"HealthState"`
		LastInBuildDurationInSeconds string `json:"LastInBuildDurationInSeconds"`
		NodeName                     string `json:"NodeName"`
		ReplicaID                    string `json:"ReplicaId"`
		ReplicaRole                  string `json:"ReplicaRole"`
		ReplicaStatus                string `json:"ReplicaStatus"`
		ServiceKind                  string `json:"ServiceKind"`
	} `json:"Items"`
}

// InstancesData encapsulates the response
// model for Instances in the Service
// Fabric API
type InstancesData struct {
	ContinuationToken *string `json:"ContinuationToken"`
	Items             []*struct {
		Address                      string `json:"Address"`
		HealthState                  string `json:"HealthState"`
		LastInBuildDurationInSeconds string `json:"LastInBuildDurationInSeconds"`
		NodeName                     string `json:"NodeName"`
		InstanceID                   string `json:"InstanceId"`
		ReplicaStatus                string `json:"ReplicaStatus"`
		ServiceKind                  string `json:"ServiceKind"`
	} `json:"Items"`
}

// ServiceType encapsulates the response
// model for Service Descriptions in the
// Service Fabric API
type ServiceType struct {
	ServiceTypeDescription struct {
		IsStateful           bool   `json:"IsStateful"`
		ServiceTypeName      string `json:"ServiceTypeName"`
		PlacementConstraints string `json:"PlacementConstraints"`
		HasPersistedState    bool   `json:"HasPersistedState"`
		Kind                 string `json:"Kind"`
		Extensions           []struct {
			Key   string `json:"Key"`
			Value string `json:"Value"`
		} `json:"Extensions"`
		LoadMetrics              []interface{} `json:"LoadMetrics"`
		ServicePlacementPolicies []interface{} `json:"ServicePlacementPolicies"`
	} `json:"ServiceTypeDescription"`
	ServiceManifestVersion string `json:"ServiceManifestVersion"`
	ServiceManifestName    string `json:"ServiceManifestName"`
	IsServiceGroup         bool   `json:"IsServiceGroup"`
}

// ServiceRoutes encapsulate traefik
// frontend configuration nested in the
// Service Fabric service's manifest file
type ServiceRoutes struct {
	Routes []struct {
		Rule string `json:"rule"`
	} `json:"routes"`
}

// ServiceManifest is the XML representation
// of a Service Fabric Service's configuration
// file
type ServiceManifest struct {
	XMLName     xml.Name `xml:"ServiceManifest"`
	Description string   `xml:"Description"`
}
