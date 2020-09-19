package ecs

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var v4MetadataJSON = []byte(`
{
    "Cluster": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:cluster/default",
    "TaskARN": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:task/default/febee046097849aba589d4435207c04a",
    "Family": "query-metadata",
    "Revision": "7",
    "DesiredStatus": "RUNNING",
    "KnownStatus": "RUNNING",
    "Limits": {
        "CPU": 0.25,
        "Memory": 512
    },
    "PullStartedAt": "2020-03-26T22:25:40.420726088Z",
    "PullStoppedAt": "2020-03-26T22:26:22.235177616Z",
    "AvailabilityZone": "us-west-2c",
    "Containers": [
        {
            "DockerId": "febee046097849aba589d4435207c04aquery-metadata",
            "Name": "query-metadata",
            "DockerName": "query-metadata",
            "Image": "mreferre/eksutils",
            "ImageID": "sha256:1b146e73f801617610dcb00441c6423e7c85a7583dd4a65ed1be03cb0e123311",
            "Labels": {
                "com.amazonaws.ecs.cluster": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:cluster/default",
                "com.amazonaws.ecs.container-name": "query-metadata",
                "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:task/default/febee046097849aba589d4435207c04a",
                "com.amazonaws.ecs.task-definition-family": "query-metadata",
                "com.amazonaws.ecs.task-definition-version": "7"
            },
            "DesiredStatus": "RUNNING",
            "KnownStatus": "RUNNING",
            "Limits": {
                "CPU": 2
            },
            "CreatedAt": "2020-03-26T22:26:24.534553758Z",
            "StartedAt": "2020-03-26T22:26:24.534553758Z",
            "Type": "NORMAL",
            "Networks": [
                {
                    "NetworkMode": "awsvpc",
                    "IPv4Addresses": [
                        "10.0.0.108"
                    ],
                    "AttachmentIndex": 0,
                    "IPv4SubnetCIDRBlock": "10.0.0.0/24",
                    "MACAddress": "0a:62:17:7a:36:68",
                    "DomainNameServers": [
                        "10.0.0.2"
                    ],
                    "DomainNameSearchList": [
                        "us-west-2.compute.internal"
                    ],
                    "PrivateDNSName": "ip-10-0-0-108.us-west-2.compute.internal",
                    "SubnetGatewayIpv4Address": ""
                }
            ]
        }
    ]
}`)

var v4Metadata = taskMetadata{
	Cluster:       "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:cluster/default",
	TaskARN:       "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:task/default/febee046097849aba589d4435207c04a",
	Family:        "query-metadata",
	Revision:      "7",
	DesiredStatus: "RUNNING",
	KnownStatus:   "RUNNING",
	Limits: taskMetadataLimits{
		CPU:    0.25,
		Memory: 512,
	},
	PullStartedAt:    "2020-03-26T22:25:40.420726088Z",
	PullStoppedAt:    "2020-03-26T22:26:22.235177616Z",
	AvailabilityZone: "us-west-2c",
	Containers: []taskMetadataContainer{
		{
			DockerID:   "febee046097849aba589d4435207c04aquery-metadata",
			Name:       "query-metadata",
			DockerName: "query-metadata",
			Image:      "mreferre/eksutils",
			ImageID:    "sha256:1b146e73f801617610dcb00441c6423e7c85a7583dd4a65ed1be03cb0e123311",
			Labels: map[string]string{
				"com.amazonaws.ecs.cluster":                 "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:cluster/default",
				"com.amazonaws.ecs.container-name":          "query-metadata",
				"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-west-2:&ExampleAWSAccountNo1;:task/default/febee046097849aba589d4435207c04a",
				"com.amazonaws.ecs.task-definition-family":  "query-metadata",
				"com.amazonaws.ecs.task-definition-version": "7",
			},
			DesiredStatus: "RUNNING",
			KnownStatus:   "RUNNING",
			Limits: taskMetadataContainerLimits{
				CPU: 2,
			},
			CreatedAt: "2020-03-26T22:26:24.534553758Z",
			StartedAt: "2020-03-26T22:26:24.534553758Z",
			Type:      "NORMAL",
			Networks: []taskMetadataNetwork{
				{
					NetworkMode: "awsvpc",
					IPv4Addresses: []string{
						"10.0.0.108",
					},
					AttachmentIndex:     0,
					IPv4SubnetCIDRBlock: "10.0.0.0/24",
					MACAddress:          "0a:62:17:7a:36:68",
					DomainNameServers: []string{
						"10.0.0.2",
					},
					DomainNameSearchList: []string{
						"us-west-2.compute.internal",
					},
					PrivateDNSName:           "ip-10-0-0-108.us-west-2.compute.internal",
					SubnetGatewayIpv4Address: "",
				},
			},
		},
	},
}

var v3MetadataJSON = []byte(`
{
  "Cluster": "default",
  "TaskARN": "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
  "Family": "nginx",
  "Revision": "5",
  "DesiredStatus": "RUNNING",
  "KnownStatus": "RUNNING",
  "Containers": [
    {
      "DockerId": "731a0d6a3b4210e2448339bc7015aaa79bfe4fa256384f4102db86ef94cbbc4c",
      "Name": "~internal~ecs~pause",
      "DockerName": "ecs-nginx-5-internalecspause-acc699c0cbf2d6d11700",
      "Image": "amazon/amazon-ecs-pause:0.1.0",
      "ImageID": "",
      "Labels": {
        "com.amazonaws.ecs.cluster": "default",
        "com.amazonaws.ecs.container-name": "~internal~ecs~pause",
        "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
        "com.amazonaws.ecs.task-definition-family": "nginx",
        "com.amazonaws.ecs.task-definition-version": "5"
      },
      "DesiredStatus": "RESOURCES_PROVISIONED",
      "KnownStatus": "RESOURCES_PROVISIONED",
      "Limits": {
        "CPU": 0,
        "Memory": 0
      },
      "CreatedAt": "2018-02-01T20:55:08.366329616Z",
      "StartedAt": "2018-02-01T20:55:09.058354915Z",
      "Type": "CNI_PAUSE",
      "Networks": [
        {
          "NetworkMode": "awsvpc",
          "IPv4Addresses": [
            "10.0.2.106"
          ]
        }
      ]
    },
    {
      "DockerId": "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946",
      "Name": "nginx-curl",
      "DockerName": "ecs-nginx-5-nginx-curl-ccccb9f49db0dfe0d901",
      "Image": "nrdlngr/nginx-curl",
      "ImageID": "sha256:2e00ae64383cfc865ba0a2ba37f61b50a120d2d9378559dcd458dc0de47bc165",
      "Labels": {
        "com.amazonaws.ecs.cluster": "default",
        "com.amazonaws.ecs.container-name": "nginx-curl",
        "com.amazonaws.ecs.task-arn": "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
        "com.amazonaws.ecs.task-definition-family": "nginx",
        "com.amazonaws.ecs.task-definition-version": "5"
      },
      "DesiredStatus": "RUNNING",
      "KnownStatus": "RUNNING",
      "Limits": {
        "CPU": 512,
        "Memory": 512
      },
      "CreatedAt": "2018-02-01T20:55:10.554941919Z",
      "StartedAt": "2018-02-01T20:55:11.064236631Z",
      "Type": "NORMAL",
      "Networks": [
        {
          "NetworkMode": "awsvpc",
          "IPv4Addresses": [
            "10.0.2.106"
          ]
        }
      ]
    }
  ],
  "PullStartedAt": "2018-02-01T20:55:09.372495529Z",
  "PullStoppedAt": "2018-02-01T20:55:10.552018345Z",
  "AvailabilityZone": "us-east-2b"
}`)

var v3Metadata = taskMetadata{
	Cluster:       "default",
	TaskARN:       "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
	Family:        "nginx",
	Revision:      "5",
	DesiredStatus: "RUNNING",
	KnownStatus:   "RUNNING",
	Containers: []taskMetadataContainer{
		{
			DockerID:   "731a0d6a3b4210e2448339bc7015aaa79bfe4fa256384f4102db86ef94cbbc4c",
			Name:       "~internal~ecs~pause",
			DockerName: "ecs-nginx-5-internalecspause-acc699c0cbf2d6d11700",
			Image:      "amazon/amazon-ecs-pause:0.1.0",
			ImageID:    "",
			Labels: map[string]string{
				"com.amazonaws.ecs.cluster":                 "default",
				"com.amazonaws.ecs.container-name":          "~internal~ecs~pause",
				"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
				"com.amazonaws.ecs.task-definition-family":  "nginx",
				"com.amazonaws.ecs.task-definition-version": "5",
			},
			DesiredStatus: "RESOURCES_PROVISIONED",
			KnownStatus:   "RESOURCES_PROVISIONED",
			Limits: taskMetadataContainerLimits{
				CPU:    0,
				Memory: 0,
			},
			CreatedAt: "2018-02-01T20:55:08.366329616Z",
			StartedAt: "2018-02-01T20:55:09.058354915Z",
			Type:      "CNI_PAUSE",
			Networks: []taskMetadataNetwork{
				{
					NetworkMode: "awsvpc",
					IPv4Addresses: []string{
						"10.0.2.106",
					},
				},
			},
		},
		{
			DockerID:   "43481a6ce4842eec8fe72fc28500c6b52edcc0917f105b83379f88cac1ff3946",
			Name:       "nginx-curl",
			DockerName: "ecs-nginx-5-nginx-curl-ccccb9f49db0dfe0d901",
			Image:      "nrdlngr/nginx-curl",
			ImageID:    "sha256:2e00ae64383cfc865ba0a2ba37f61b50a120d2d9378559dcd458dc0de47bc165",
			Labels: map[string]string{
				"com.amazonaws.ecs.cluster":                 "default",
				"com.amazonaws.ecs.container-name":          "nginx-curl",
				"com.amazonaws.ecs.task-arn":                "arn:aws:ecs:us-east-2:012345678910:task/9781c248-0edd-4cdb-9a93-f63cb662a5d3",
				"com.amazonaws.ecs.task-definition-family":  "nginx",
				"com.amazonaws.ecs.task-definition-version": "5",
			},
			DesiredStatus: "RUNNING",
			KnownStatus:   "RUNNING",
			Limits: taskMetadataContainerLimits{
				CPU:    512,
				Memory: 512,
			},
			CreatedAt: "2018-02-01T20:55:10.554941919Z",
			StartedAt: "2018-02-01T20:55:11.064236631Z",
			Type:      "NORMAL",
			Networks: []taskMetadataNetwork{
				{
					NetworkMode: "awsvpc",
					IPv4Addresses: []string{
						"10.0.2.106",
					},
				},
			},
		},
	},
	PullStartedAt:    "2018-02-01T20:55:09.372495529Z",
	PullStoppedAt:    "2018-02-01T20:55:10.552018345Z",
	AvailabilityZone: "us-east-2b",
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func mockClient(respTime time.Duration, resp *http.Response) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			time.Sleep(respTime)

			if resp == nil {
				return nil
			}

			if resp.Header == nil {
				resp.Header = make(http.Header)
			}

			if strings.HasPrefix(req.URL.Path, "/v3") && strings.HasSuffix(req.URL.Path, TaskMetadataEndpointV3Path) {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(v3MetadataJSON))
			}
			if strings.HasPrefix(req.URL.Path, "/v4") && strings.HasSuffix(req.URL.Path, TaskMetadataEndpointV4Path) {
				resp.Body = ioutil.NopCloser(bytes.NewBuffer(v4MetadataJSON))
			}

			return resp
		}),
	}
}

func Test_fetchECSTaskMetadata(t *testing.T) {
	tests := []struct {
		desc           string
		environ        map[string]string
		resp           *http.Response
		expData        *taskMetadata
		expClusterName string
		expErr         bool
		respTimeout    time.Duration
		respTime       time.Duration
	}{
		{
			desc: "should return the parsed v4 task metadata",
			environ: map[string]string{
				ContainerMetadataEndpointV4EnvKey: "http://169.254.170.2/v4/container-id",
			},
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			expData:        &v4Metadata,
			expClusterName: "default",
		},
		{
			desc: "should return the parsed v3 task metadata",
			environ: map[string]string{
				ContainerMetadataEndpointV3EnvKey: "http://169.254.170.2/v3/container-id",
			},
			resp: &http.Response{
				StatusCode: http.StatusOK,
			},
			expData:        &v3Metadata,
			expClusterName: "default",
		},
		{
			desc:   "no metadata environment key",
			expErr: true,
		},
		{
			desc: "task metadata fetch timeout",
			environ: map[string]string{
				ContainerMetadataEndpointV4EnvKey: "http://169.254.170.2/v4/container-id",
			},
			respTimeout: time.Second * 1,
			respTime:    time.Second * 2,
			expErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			for k, v := range test.environ {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range test.environ {
					os.Unsetenv(k)
				}
			}()

			client := mockClient(test.respTime, test.resp)
			if 0 < test.respTimeout {
				client.Timeout = test.respTimeout
			}

			meta, err := fetchECSTaskMetadata(context.TODO(), client)
			if test.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.EqualValues(t, test.expData, meta)
			assert.Equal(t, test.expClusterName, meta.ClusterName())
		})
	}
}
