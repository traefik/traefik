package client

const (
	LAUNCH_CONFIG_TYPE = "launchConfig"
)

type LaunchConfig struct {
	Resource

	AccountId string `json:"accountId,omitempty" yaml:"account_id,omitempty"`

	AgentId string `json:"agentId,omitempty" yaml:"agent_id,omitempty"`

	AllocationState string `json:"allocationState,omitempty" yaml:"allocation_state,omitempty"`

	BlkioDeviceOptions map[string]interface{} `json:"blkioDeviceOptions,omitempty" yaml:"blkio_device_options,omitempty"`

	BlkioWeight int64 `json:"blkioWeight,omitempty" yaml:"blkio_weight,omitempty"`

	Build *DockerBuild `json:"build,omitempty" yaml:"build,omitempty"`

	CapAdd []string `json:"capAdd,omitempty" yaml:"cap_add,omitempty"`

	CapDrop []string `json:"capDrop,omitempty" yaml:"cap_drop,omitempty"`

	CgroupParent string `json:"cgroupParent,omitempty" yaml:"cgroup_parent,omitempty"`

	Command []string `json:"command,omitempty" yaml:"command,omitempty"`

	Count int64 `json:"count,omitempty" yaml:"count,omitempty"`

	CpuCount int64 `json:"cpuCount,omitempty" yaml:"cpu_count,omitempty"`

	CpuPercent int64 `json:"cpuPercent,omitempty" yaml:"cpu_percent,omitempty"`

	CpuPeriod int64 `json:"cpuPeriod,omitempty" yaml:"cpu_period,omitempty"`

	CpuQuota int64 `json:"cpuQuota,omitempty" yaml:"cpu_quota,omitempty"`

	CpuRealtimePeriod int64 `json:"cpuRealtimePeriod,omitempty" yaml:"cpu_realtime_period,omitempty"`

	CpuRealtimeRuntime int64 `json:"cpuRealtimeRuntime,omitempty" yaml:"cpu_realtime_runtime,omitempty"`

	CpuSet string `json:"cpuSet,omitempty" yaml:"cpu_set,omitempty"`

	CpuSetMems string `json:"cpuSetMems,omitempty" yaml:"cpu_set_mems,omitempty"`

	CpuShares int64 `json:"cpuShares,omitempty" yaml:"cpu_shares,omitempty"`

	CreateIndex int64 `json:"createIndex,omitempty" yaml:"create_index,omitempty"`

	Created string `json:"created,omitempty" yaml:"created,omitempty"`

	Data map[string]interface{} `json:"data,omitempty" yaml:"data,omitempty"`

	DataVolumeMounts map[string]interface{} `json:"dataVolumeMounts,omitempty" yaml:"data_volume_mounts,omitempty"`

	DataVolumes []string `json:"dataVolumes,omitempty" yaml:"data_volumes,omitempty"`

	DataVolumesFrom []string `json:"dataVolumesFrom,omitempty" yaml:"data_volumes_from,omitempty"`

	DataVolumesFromLaunchConfigs []string `json:"dataVolumesFromLaunchConfigs,omitempty" yaml:"data_volumes_from_launch_configs,omitempty"`

	DeploymentUnitUuid string `json:"deploymentUnitUuid,omitempty" yaml:"deployment_unit_uuid,omitempty"`

	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	Devices []string `json:"devices,omitempty" yaml:"devices,omitempty"`

	DiskQuota int64 `json:"diskQuota,omitempty" yaml:"disk_quota,omitempty"`

	Disks []VirtualMachineDisk `json:"disks,omitempty" yaml:"disks,omitempty"`

	Dns []string `json:"dns,omitempty" yaml:"dns,omitempty"`

	DnsOpt []string `json:"dnsOpt,omitempty" yaml:"dns_opt,omitempty"`

	DnsSearch []string `json:"dnsSearch,omitempty" yaml:"dns_search,omitempty"`

	DomainName string `json:"domainName,omitempty" yaml:"domain_name,omitempty"`

	DrainTimeoutMs int64 `json:"drainTimeoutMs,omitempty" yaml:"drain_timeout_ms,omitempty"`

	EntryPoint []string `json:"entryPoint,omitempty" yaml:"entry_point,omitempty"`

	Environment map[string]interface{} `json:"environment,omitempty" yaml:"environment,omitempty"`

	Expose []string `json:"expose,omitempty" yaml:"expose,omitempty"`

	ExternalId string `json:"externalId,omitempty" yaml:"external_id,omitempty"`

	ExtraHosts []string `json:"extraHosts,omitempty" yaml:"extra_hosts,omitempty"`

	FirstRunning string `json:"firstRunning,omitempty" yaml:"first_running,omitempty"`

	GroupAdd []string `json:"groupAdd,omitempty" yaml:"group_add,omitempty"`

	HealthCheck *InstanceHealthCheck `json:"healthCheck,omitempty" yaml:"health_check,omitempty"`

	HealthCmd []string `json:"healthCmd,omitempty" yaml:"health_cmd,omitempty"`

	HealthInterval int64 `json:"healthInterval,omitempty" yaml:"health_interval,omitempty"`

	HealthRetries int64 `json:"healthRetries,omitempty" yaml:"health_retries,omitempty"`

	HealthState string `json:"healthState,omitempty" yaml:"health_state,omitempty"`

	HealthTimeout int64 `json:"healthTimeout,omitempty" yaml:"health_timeout,omitempty"`

	HostId string `json:"hostId,omitempty" yaml:"host_id,omitempty"`

	Hostname string `json:"hostname,omitempty" yaml:"hostname,omitempty"`

	ImageUuid string `json:"imageUuid,omitempty" yaml:"image_uuid,omitempty"`

	InstanceLinks map[string]interface{} `json:"instanceLinks,omitempty" yaml:"instance_links,omitempty"`

	InstanceTriggeredStop string `json:"instanceTriggeredStop,omitempty" yaml:"instance_triggered_stop,omitempty"`

	IoMaximumBandwidth int64 `json:"ioMaximumBandwidth,omitempty" yaml:"io_maximum_bandwidth,omitempty"`

	IoMaximumIOps int64 `json:"ioMaximumIOps,omitempty" yaml:"io_maximum_iops,omitempty"`

	Ip string `json:"ip,omitempty" yaml:"ip,omitempty"`

	Ip6 string `json:"ip6,omitempty" yaml:"ip6,omitempty"`

	IpcMode string `json:"ipcMode,omitempty" yaml:"ipc_mode,omitempty"`

	Isolation string `json:"isolation,omitempty" yaml:"isolation,omitempty"`

	KernelMemory int64 `json:"kernelMemory,omitempty" yaml:"kernel_memory,omitempty"`

	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`

	Labels map[string]interface{} `json:"labels,omitempty" yaml:"labels,omitempty"`

	LogConfig *LogConfig `json:"logConfig,omitempty" yaml:"log_config,omitempty"`

	LxcConf map[string]interface{} `json:"lxcConf,omitempty" yaml:"lxc_conf,omitempty"`

	Memory int64 `json:"memory,omitempty" yaml:"memory,omitempty"`

	MemoryMb int64 `json:"memoryMb,omitempty" yaml:"memory_mb,omitempty"`

	MemoryReservation int64 `json:"memoryReservation,omitempty" yaml:"memory_reservation,omitempty"`

	MemorySwap int64 `json:"memorySwap,omitempty" yaml:"memory_swap,omitempty"`

	MemorySwappiness int64 `json:"memorySwappiness,omitempty" yaml:"memory_swappiness,omitempty"`

	MilliCpuReservation int64 `json:"milliCpuReservation,omitempty" yaml:"milli_cpu_reservation,omitempty"`

	Mounts []MountEntry `json:"mounts,omitempty" yaml:"mounts,omitempty"`

	NativeContainer bool `json:"nativeContainer,omitempty" yaml:"native_container,omitempty"`

	NetAlias []string `json:"netAlias,omitempty" yaml:"net_alias,omitempty"`

	NetworkContainerId string `json:"networkContainerId,omitempty" yaml:"network_container_id,omitempty"`

	NetworkIds []string `json:"networkIds,omitempty" yaml:"network_ids,omitempty"`

	NetworkLaunchConfig string `json:"networkLaunchConfig,omitempty" yaml:"network_launch_config,omitempty"`

	NetworkMode string `json:"networkMode,omitempty" yaml:"network_mode,omitempty"`

	OomKillDisable bool `json:"oomKillDisable,omitempty" yaml:"oom_kill_disable,omitempty"`

	OomScoreAdj int64 `json:"oomScoreAdj,omitempty" yaml:"oom_score_adj,omitempty"`

	PidMode string `json:"pidMode,omitempty" yaml:"pid_mode,omitempty"`

	PidsLimit int64 `json:"pidsLimit,omitempty" yaml:"pids_limit,omitempty"`

	Ports []string `json:"ports,omitempty" yaml:"ports,omitempty"`

	PrimaryIpAddress string `json:"primaryIpAddress,omitempty" yaml:"primary_ip_address,omitempty"`

	PrimaryNetworkId string `json:"primaryNetworkId,omitempty" yaml:"primary_network_id,omitempty"`

	Privileged bool `json:"privileged,omitempty" yaml:"privileged,omitempty"`

	PublishAllPorts bool `json:"publishAllPorts,omitempty" yaml:"publish_all_ports,omitempty"`

	ReadOnly bool `json:"readOnly,omitempty" yaml:"read_only,omitempty"`

	RegistryCredentialId string `json:"registryCredentialId,omitempty" yaml:"registry_credential_id,omitempty"`

	RemoveTime string `json:"removeTime,omitempty" yaml:"remove_time,omitempty"`

	Removed string `json:"removed,omitempty" yaml:"removed,omitempty"`

	RequestedHostId string `json:"requestedHostId,omitempty" yaml:"requested_host_id,omitempty"`

	RequestedIpAddress string `json:"requestedIpAddress,omitempty" yaml:"requested_ip_address,omitempty"`

	RunInit bool `json:"runInit,omitempty" yaml:"run_init,omitempty"`

	Secrets []SecretReference `json:"secrets,omitempty" yaml:"secrets,omitempty"`

	SecurityOpt []string `json:"securityOpt,omitempty" yaml:"security_opt,omitempty"`

	ServiceId string `json:"serviceId,omitempty" yaml:"service_id,omitempty"`

	ServiceIds []string `json:"serviceIds,omitempty" yaml:"service_ids,omitempty"`

	ShmSize int64 `json:"shmSize,omitempty" yaml:"shm_size,omitempty"`

	StackId string `json:"stackId,omitempty" yaml:"stack_id,omitempty"`

	StartCount int64 `json:"startCount,omitempty" yaml:"start_count,omitempty"`

	StartOnCreate bool `json:"startOnCreate,omitempty" yaml:"start_on_create,omitempty"`

	State string `json:"state,omitempty" yaml:"state,omitempty"`

	StdinOpen bool `json:"stdinOpen,omitempty" yaml:"stdin_open,omitempty"`

	StopSignal string `json:"stopSignal,omitempty" yaml:"stop_signal,omitempty"`

	StopTimeout int64 `json:"stopTimeout,omitempty" yaml:"stop_timeout,omitempty"`

	StorageOpt map[string]interface{} `json:"storageOpt,omitempty" yaml:"storage_opt,omitempty"`

	Sysctls map[string]interface{} `json:"sysctls,omitempty" yaml:"sysctls,omitempty"`

	System bool `json:"system,omitempty" yaml:"system,omitempty"`

	Tmpfs map[string]interface{} `json:"tmpfs,omitempty" yaml:"tmpfs,omitempty"`

	Token string `json:"token,omitempty" yaml:"token,omitempty"`

	Transitioning string `json:"transitioning,omitempty" yaml:"transitioning,omitempty"`

	TransitioningMessage string `json:"transitioningMessage,omitempty" yaml:"transitioning_message,omitempty"`

	TransitioningProgress int64 `json:"transitioningProgress,omitempty" yaml:"transitioning_progress,omitempty"`

	Tty bool `json:"tty,omitempty" yaml:"tty,omitempty"`

	Ulimits []Ulimit `json:"ulimits,omitempty" yaml:"ulimits,omitempty"`

	User string `json:"user,omitempty" yaml:"user,omitempty"`

	UserPorts []string `json:"userPorts,omitempty" yaml:"user_ports,omitempty"`

	Userdata string `json:"userdata,omitempty" yaml:"userdata,omitempty"`

	UsernsMode string `json:"usernsMode,omitempty" yaml:"userns_mode,omitempty"`

	Uts string `json:"uts,omitempty" yaml:"uts,omitempty"`

	Uuid string `json:"uuid,omitempty" yaml:"uuid,omitempty"`

	Vcpu int64 `json:"vcpu,omitempty" yaml:"vcpu,omitempty"`

	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	VolumeDriver string `json:"volumeDriver,omitempty" yaml:"volume_driver,omitempty"`

	WorkingDir string `json:"workingDir,omitempty" yaml:"working_dir,omitempty"`
}

type LaunchConfigCollection struct {
	Collection
	Data   []LaunchConfig `json:"data,omitempty"`
	client *LaunchConfigClient
}

type LaunchConfigClient struct {
	rancherClient *RancherClient
}

type LaunchConfigOperations interface {
	List(opts *ListOpts) (*LaunchConfigCollection, error)
	Create(opts *LaunchConfig) (*LaunchConfig, error)
	Update(existing *LaunchConfig, updates interface{}) (*LaunchConfig, error)
	ById(id string) (*LaunchConfig, error)
	Delete(container *LaunchConfig) error

	ActionAllocate(*LaunchConfig) (*Instance, error)

	ActionConsole(*LaunchConfig, *InstanceConsoleInput) (*InstanceConsole, error)

	ActionCreate(*LaunchConfig) (*Instance, error)

	ActionDeallocate(*LaunchConfig) (*Instance, error)

	ActionError(*LaunchConfig) (*Instance, error)

	ActionExecute(*LaunchConfig, *ContainerExec) (*HostAccess, error)

	ActionMigrate(*LaunchConfig) (*Instance, error)

	ActionProxy(*LaunchConfig, *ContainerProxy) (*HostAccess, error)

	ActionPurge(*LaunchConfig) (*Instance, error)

	ActionRemove(*LaunchConfig) (*Instance, error)

	ActionRestart(*LaunchConfig) (*Instance, error)

	ActionStart(*LaunchConfig) (*Instance, error)

	ActionStop(*LaunchConfig, *InstanceStop) (*Instance, error)

	ActionUpdate(*LaunchConfig) (*Instance, error)

	ActionUpdatehealthy(*LaunchConfig) (*Instance, error)

	ActionUpdatereinitializing(*LaunchConfig) (*Instance, error)

	ActionUpdateunhealthy(*LaunchConfig) (*Instance, error)
}

func newLaunchConfigClient(rancherClient *RancherClient) *LaunchConfigClient {
	return &LaunchConfigClient{
		rancherClient: rancherClient,
	}
}

func (c *LaunchConfigClient) Create(container *LaunchConfig) (*LaunchConfig, error) {
	resp := &LaunchConfig{}
	err := c.rancherClient.doCreate(LAUNCH_CONFIG_TYPE, container, resp)
	return resp, err
}

func (c *LaunchConfigClient) Update(existing *LaunchConfig, updates interface{}) (*LaunchConfig, error) {
	resp := &LaunchConfig{}
	err := c.rancherClient.doUpdate(LAUNCH_CONFIG_TYPE, &existing.Resource, updates, resp)
	return resp, err
}

func (c *LaunchConfigClient) List(opts *ListOpts) (*LaunchConfigCollection, error) {
	resp := &LaunchConfigCollection{}
	err := c.rancherClient.doList(LAUNCH_CONFIG_TYPE, opts, resp)
	resp.client = c
	return resp, err
}

func (cc *LaunchConfigCollection) Next() (*LaunchConfigCollection, error) {
	if cc != nil && cc.Pagination != nil && cc.Pagination.Next != "" {
		resp := &LaunchConfigCollection{}
		err := cc.client.rancherClient.doNext(cc.Pagination.Next, resp)
		resp.client = cc.client
		return resp, err
	}
	return nil, nil
}

func (c *LaunchConfigClient) ById(id string) (*LaunchConfig, error) {
	resp := &LaunchConfig{}
	err := c.rancherClient.doById(LAUNCH_CONFIG_TYPE, id, resp)
	if apiError, ok := err.(*ApiError); ok {
		if apiError.StatusCode == 404 {
			return nil, nil
		}
	}
	return resp, err
}

func (c *LaunchConfigClient) Delete(container *LaunchConfig) error {
	return c.rancherClient.doResourceDelete(LAUNCH_CONFIG_TYPE, &container.Resource)
}

func (c *LaunchConfigClient) ActionAllocate(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "allocate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionConsole(resource *LaunchConfig, input *InstanceConsoleInput) (*InstanceConsole, error) {

	resp := &InstanceConsole{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "console", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionCreate(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "create", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionDeallocate(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "deallocate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionError(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "error", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionExecute(resource *LaunchConfig, input *ContainerExec) (*HostAccess, error) {

	resp := &HostAccess{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "execute", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionMigrate(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "migrate", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionProxy(resource *LaunchConfig, input *ContainerProxy) (*HostAccess, error) {

	resp := &HostAccess{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "proxy", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionPurge(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "purge", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionRemove(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "remove", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionRestart(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "restart", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionStart(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "start", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionStop(resource *LaunchConfig, input *InstanceStop) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "stop", &resource.Resource, input, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionUpdate(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "update", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionUpdatehealthy(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "updatehealthy", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionUpdatereinitializing(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "updatereinitializing", &resource.Resource, nil, resp)

	return resp, err
}

func (c *LaunchConfigClient) ActionUpdateunhealthy(resource *LaunchConfig) (*Instance, error) {

	resp := &Instance{}

	err := c.rancherClient.doAction(LAUNCH_CONFIG_TYPE, "updateunhealthy", &resource.Resource, nil, resp)

	return resp, err
}
