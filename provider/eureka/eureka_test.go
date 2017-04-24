package eureka

import (
	"testing"

	"github.com/ArthurHlt/go-eureka-client/eureka"
)

func TestEurekaGetPort(t *testing.T) {
	cases := []struct {
		expectedPort string
		instanceInfo eureka.InstanceInfo
	}{
		{
			expectedPort: "80",
			instanceInfo: eureka.InstanceInfo{
				SecurePort: &eureka.Port{
					Port: 443, Enabled: false,
				},
				Port: &eureka.Port{
					Port: 80, Enabled: true,
				},
			},
		},
		{
			expectedPort: "443",
			instanceInfo: eureka.InstanceInfo{
				SecurePort: &eureka.Port{
					Port: 443, Enabled: true,
				},
				Port: &eureka.Port{
					Port: 80, Enabled: false,
				},
			},
		},
	}

	eurekaProvider := &Provider{}
	for _, c := range cases {
		port := eurekaProvider.getPort(c.instanceInfo)
		if port != c.expectedPort {
			t.Fatalf("Should have been %s, got %s", c.expectedPort, port)
		}
	}
}

func TestEurekaGetProtocol(t *testing.T) {
	cases := []struct {
		expectedProtocol string
		instanceInfo     eureka.InstanceInfo
	}{
		{
			expectedProtocol: "http",
			instanceInfo: eureka.InstanceInfo{
				SecurePort: &eureka.Port{
					Port: 443, Enabled: false,
				},
				Port: &eureka.Port{
					Port: 80, Enabled: true,
				},
			},
		},
		{
			expectedProtocol: "https",
			instanceInfo: eureka.InstanceInfo{
				SecurePort: &eureka.Port{
					Port: 443, Enabled: true,
				},
				Port: &eureka.Port{
					Port: 80, Enabled: false,
				},
			},
		},
	}

	eurekaProvider := &Provider{}
	for _, c := range cases {
		protocol := eurekaProvider.getProtocol(c.instanceInfo)
		if protocol != c.expectedProtocol {
			t.Fatalf("Should have been %s, got %s", c.expectedProtocol, protocol)
		}
	}
}

func TestEurekaGetWeight(t *testing.T) {
	cases := []struct {
		expectedWeight string
		instanceInfo   eureka.InstanceInfo
	}{
		{
			expectedWeight: "0",
			instanceInfo: eureka.InstanceInfo{
				Port: &eureka.Port{
					Port: 80, Enabled: true,
				},
				Metadata: &eureka.MetaData{
					Map: map[string]string{},
				},
			},
		},
		{
			expectedWeight: "10",
			instanceInfo: eureka.InstanceInfo{
				Port: &eureka.Port{
					Port: 80, Enabled: true,
				},
				Metadata: &eureka.MetaData{
					Map: map[string]string{
						"traefik.weight": "10",
					},
				},
			},
		},
	}

	eurekaProvider := &Provider{}
	for _, c := range cases {
		weight := eurekaProvider.getWeight(c.instanceInfo)
		if weight != c.expectedWeight {
			t.Fatalf("Should have been %s, got %s", c.expectedWeight, weight)
		}
	}
}

func TestEurekaGetInstanceId(t *testing.T) {
	cases := []struct {
		expectedID   string
		instanceInfo eureka.InstanceInfo
	}{
		{
			expectedID: "MyInstanceId",
			instanceInfo: eureka.InstanceInfo{
				IpAddr: "10.11.12.13",
				SecurePort: &eureka.Port{
					Port: 443, Enabled: false,
				},
				Port: &eureka.Port{
					Port: 80, Enabled: true,
				},
				Metadata: &eureka.MetaData{
					Map: map[string]string{
						"traefik.backend.id": "MyInstanceId",
					},
				},
			},
		},
		{
			expectedID: "10-11-12-13-80",
			instanceInfo: eureka.InstanceInfo{
				IpAddr: "10.11.12.13",
				SecurePort: &eureka.Port{
					Port: 443, Enabled: false,
				},
				Port: &eureka.Port{
					Port: 80, Enabled: true,
				},
				Metadata: &eureka.MetaData{
					Map: map[string]string{},
				},
			},
		},
	}

	eurekaProvider := &Provider{}
	for _, c := range cases {
		id := eurekaProvider.getInstanceID(c.instanceInfo)
		if id != c.expectedID {
			t.Fatalf("Should have been %s, got %s", c.expectedID, id)
		}
	}
}
