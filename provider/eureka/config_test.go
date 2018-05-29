package eureka

import (
	"strconv"
	"testing"

	"github.com/ArthurHlt/go-eureka-client/eureka"
	"github.com/containous/traefik/provider/label"
	"github.com/stretchr/testify/assert"
)

func TestGetPort(t *testing.T) {
	testCases := []struct {
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

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			port := getPort(test.instanceInfo)
			assert.Equal(t, test.expectedPort, port)
		})
	}
}

func TestGetProtocol(t *testing.T) {
	testCases := []struct {
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

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			protocol := getProtocol(test.instanceInfo)
			assert.Equal(t, test.expectedProtocol, protocol)
		})
	}
}

func TestGetWeight(t *testing.T) {
	testCases := []struct {
		expectedWeight int
		instanceInfo   eureka.InstanceInfo
	}{
		{
			expectedWeight: label.DefaultWeight,
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
			expectedWeight: 10,
			instanceInfo: eureka.InstanceInfo{
				Port: &eureka.Port{
					Port: 80, Enabled: true,
				},
				Metadata: &eureka.MetaData{
					Map: map[string]string{
						label.TraefikWeight: "10",
					},
				},
			},
		},
	}

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			weight := getWeight(test.instanceInfo)
			assert.Equal(t, test.expectedWeight, weight)
		})
	}
}

func TestGetInstanceId(t *testing.T) {
	testCases := []struct {
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
						label.TraefikBackendID: "MyInstanceId",
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

	for i, test := range testCases {
		test := test
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			id := getInstanceID(test.instanceInfo)
			assert.Equal(t, test.expectedID, id)
		})
	}
}
