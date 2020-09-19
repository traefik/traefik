package ecs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/traefik/traefik/v2/pkg/log"
)

// please see https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-metadata-endpoint.html

const (
	// ContainerMetadataEndpointV3EnvKey is a environment key of ECS Container metadata V3 endpoint URL.
	ContainerMetadataEndpointV3EnvKey = "ECS_CONTAINER_METADATA_URI"
	// ContainerMetadataEndpointV4EnvKey is a environment key of ECS Container metadata V4 endpoint URL.
	ContainerMetadataEndpointV4EnvKey = "ECS_CONTAINER_METADATA_URI_V4"
	// TaskMetadataEndpointV3Path is path of ECS Task metadata V3 URL.
	TaskMetadataEndpointV3Path = "/task"
	// TaskMetadataEndpointV4Path is path of ECS Task metadata V4 URL.
	TaskMetadataEndpointV4Path = "/task"
)

type taskMetadata struct {
	Cluster          string                  `json:"Cluster"`
	TaskARN          string                  `json:"TaskARN"`
	Family           string                  `json:"Family"`
	Revision         string                  `json:"Revision"`
	DesiredStatus    string                  `json:"DesiredStatus"`
	KnownStatus      string                  `json:"KnownStatus"`
	Limits           taskMetadataLimits      `json:"Limits"`
	PullStartedAt    string                  `json:"PullStartedAt"`
	PullStoppedAt    string                  `json:"PullStoppedAt"`
	AvailabilityZone string                  `json:"AvailabilityZone"`
	Containers       []taskMetadataContainer `json:"Containers"`
}

func (t *taskMetadata) ClusterName() string {
	idx := strings.Index(t.Cluster, ":cluster/")
	if 0 < idx {
		return t.Cluster[idx+9:]
	}
	return t.Cluster
}

type taskMetadataContainer struct {
	DockerID      string                      `json:"DockerId"`
	Name          string                      `json:"Name"`
	DockerName    string                      `json:"DockerName"`
	Image         string                      `json:"Image"`
	ImageID       string                      `json:"ImageID"`
	Ports         []taskMetadataContainerPort `json:"Ports"`
	Labels        map[string]string           `json:"Labels"`
	DesiredStatus string                      `json:"DesiredStatus"`
	KnownStatus   string                      `json:"KnownStatus"`
	Limits        taskMetadataContainerLimits `json:"Limits"`
	CreatedAt     string                      `json:"CreatedAt"`
	StartedAt     string                      `json:"StartedAt"`
	Type          string                      `json:"Type"`
	Networks      []taskMetadataNetwork       `json:"Networks"`
}

type taskMetadataContainerPort struct {
	ContainerPort int64  `json:"ContainerPort"`
	Protocol      string `json:"Protocol"`
	HostPort      int64  `json:"HostPort"`
}

type taskMetadataContainerLimits struct {
	CPU    int64 `json:"CPU"`
	Memory int64 `json:"Memory"`
}

type taskMetadataNetwork struct {
	NetworkMode              string   `json:"NetworkMode"`
	IPv4Addresses            []string `json:"IPv4Addresses"`
	AttachmentIndex          int64    `json:"AttachmentIndex"`
	IPv4SubnetCIDRBlock      string   `json:"IPv4SubnetCIDRBlock"`
	MACAddress               string   `json:"MACAddress"`
	DomainNameServers        []string `json:"DomainNameServers"`
	DomainNameSearchList     []string `json:"DomainNameSearchList"`
	PrivateDNSName           string   `json:"PrivateDNSName"`
	SubnetGatewayIpv4Address string   `json:"SubnetGatewayIpv4Address"`
}

type taskMetadataLimits struct {
	CPU    float64 `json:"CPU"`
	Memory int64   `json:"Memory"`
}

func fetchECSTaskMetadata(ctx context.Context, client *http.Client) (meta *taskMetadata, err error) {
	var endpoint string
	if v4endpoint := os.Getenv(ContainerMetadataEndpointV4EnvKey); v4endpoint != "" {
		endpoint = v4endpoint + TaskMetadataEndpointV4Path
	} else if v3endpoint := os.Getenv(ContainerMetadataEndpointV3EnvKey); v3endpoint != "" {
		endpoint = v3endpoint + TaskMetadataEndpointV3Path
	} else {
		return nil, fmt.Errorf("can't get ecs metadata endpoint from environment")
	}
	log.FromContext(ctx).Infof("endpoint: %s", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-ok response code: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&meta); err != nil {
		return nil, err
	}
	return meta, nil
}
