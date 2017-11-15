package client

type RancherClient struct {
	RancherBaseClient

	Account                                  AccountOperations
	ActiveSetting                            ActiveSettingOperations
	AddOutputsInput                          AddOutputsInputOperations
	AddRemoveServiceLinkInput                AddRemoveServiceLinkInputOperations
	Agent                                    AgentOperations
	Amazonec2Config                          Amazonec2ConfigOperations
	ApiKey                                   ApiKeyOperations
	AuditLog                                 AuditLogOperations
	AzureConfig                              AzureConfigOperations
	Azureadconfig                            AzureadconfigOperations
	Backup                                   BackupOperations
	BackupTarget                             BackupTargetOperations
	BaseMachineConfig                        BaseMachineConfigOperations
	Binding                                  BindingOperations
	BlkioDeviceOption                        BlkioDeviceOptionOperations
	CatalogTemplate                          CatalogTemplateOperations
	Certificate                              CertificateOperations
	ChangeSecretInput                        ChangeSecretInputOperations
	ClusterMembership                        ClusterMembershipOperations
	ComposeConfig                            ComposeConfigOperations
	ComposeConfigInput                       ComposeConfigInputOperations
	ComposeProject                           ComposeProjectOperations
	ComposeService                           ComposeServiceOperations
	ConfigItem                               ConfigItemOperations
	ConfigItemStatus                         ConfigItemStatusOperations
	Container                                ContainerOperations
	ContainerEvent                           ContainerEventOperations
	ContainerExec                            ContainerExecOperations
	ContainerLogs                            ContainerLogsOperations
	ContainerProxy                           ContainerProxyOperations
	Credential                               CredentialOperations
	Databasechangelog                        DatabasechangelogOperations
	Databasechangeloglock                    DatabasechangeloglockOperations
	DefaultNetwork                           DefaultNetworkOperations
	DigitaloceanConfig                       DigitaloceanConfigOperations
	DnsService                               DnsServiceOperations
	DockerBuild                              DockerBuildOperations
	ExtensionImplementation                  ExtensionImplementationOperations
	ExtensionPoint                           ExtensionPointOperations
	ExternalDnsEvent                         ExternalDnsEventOperations
	ExternalEvent                            ExternalEventOperations
	ExternalHandler                          ExternalHandlerOperations
	ExternalHandlerExternalHandlerProcessMap ExternalHandlerExternalHandlerProcessMapOperations
	ExternalHandlerProcess                   ExternalHandlerProcessOperations
	ExternalHandlerProcessConfig             ExternalHandlerProcessConfigOperations
	ExternalHostEvent                        ExternalHostEventOperations
	ExternalService                          ExternalServiceOperations
	ExternalServiceEvent                     ExternalServiceEventOperations
	ExternalStoragePoolEvent                 ExternalStoragePoolEventOperations
	ExternalVolumeEvent                      ExternalVolumeEventOperations
	FieldDocumentation                       FieldDocumentationOperations
	GenericObject                            GenericObjectOperations
	HaConfig                                 HaConfigOperations
	HaConfigInput                            HaConfigInputOperations
	HealthcheckInstanceHostMap               HealthcheckInstanceHostMapOperations
	Host                                     HostOperations
	HostAccess                               HostAccessOperations
	HostApiProxyToken                        HostApiProxyTokenOperations
	HostTemplate                             HostTemplateOperations
	Identity                                 IdentityOperations
	Image                                    ImageOperations
	InServiceUpgradeStrategy                 InServiceUpgradeStrategyOperations
	Instance                                 InstanceOperations
	InstanceConsole                          InstanceConsoleOperations
	InstanceConsoleInput                     InstanceConsoleInputOperations
	InstanceHealthCheck                      InstanceHealthCheckOperations
	InstanceLink                             InstanceLinkOperations
	InstanceStop                             InstanceStopOperations
	IpAddress                                IpAddressOperations
	KubernetesService                        KubernetesServiceOperations
	KubernetesStack                          KubernetesStackOperations
	KubernetesStackUpgrade                   KubernetesStackUpgradeOperations
	Label                                    LabelOperations
	LaunchConfig                             LaunchConfigOperations
	LbConfig                                 LbConfigOperations
	LbTargetConfig                           LbTargetConfigOperations
	LoadBalancerCookieStickinessPolicy       LoadBalancerCookieStickinessPolicyOperations
	LoadBalancerService                      LoadBalancerServiceOperations
	LocalAuthConfig                          LocalAuthConfigOperations
	LogConfig                                LogConfigOperations
	Machine                                  MachineOperations
	MachineDriver                            MachineDriverOperations
	Mount                                    MountOperations
	MountEntry                               MountEntryOperations
	Network                                  NetworkOperations
	NetworkDriver                            NetworkDriverOperations
	NetworkDriverService                     NetworkDriverServiceOperations
	NetworkPolicyRule                        NetworkPolicyRuleOperations
	NetworkPolicyRuleBetween                 NetworkPolicyRuleBetweenOperations
	NetworkPolicyRuleMember                  NetworkPolicyRuleMemberOperations
	NetworkPolicyRuleWithin                  NetworkPolicyRuleWithinOperations
	NfsConfig                                NfsConfigOperations
	Openldapconfig                           OpenldapconfigOperations
	PacketConfig                             PacketConfigOperations
	Password                                 PasswordOperations
	PhysicalHost                             PhysicalHostOperations
	Port                                     PortOperations
	PortRule                                 PortRuleOperations
	ProcessDefinition                        ProcessDefinitionOperations
	ProcessExecution                         ProcessExecutionOperations
	ProcessInstance                          ProcessInstanceOperations
	ProcessPool                              ProcessPoolOperations
	ProcessSummary                           ProcessSummaryOperations
	Project                                  ProjectOperations
	ProjectMember                            ProjectMemberOperations
	ProjectTemplate                          ProjectTemplateOperations
	PublicEndpoint                           PublicEndpointOperations
	Publish                                  PublishOperations
	PullTask                                 PullTaskOperations
	RecreateOnQuorumStrategyConfig           RecreateOnQuorumStrategyConfigOperations
	Register                                 RegisterOperations
	RegistrationToken                        RegistrationTokenOperations
	Registry                                 RegistryOperations
	RegistryCredential                       RegistryCredentialOperations
	ResourceDefinition                       ResourceDefinitionOperations
	RestartPolicy                            RestartPolicyOperations
	RestoreFromBackupInput                   RestoreFromBackupInputOperations
	RevertToSnapshotInput                    RevertToSnapshotInputOperations
	RollingRestartStrategy                   RollingRestartStrategyOperations
	ScalePolicy                              ScalePolicyOperations
	ScheduledUpgrade                         ScheduledUpgradeOperations
	SecondaryLaunchConfig                    SecondaryLaunchConfigOperations
	Secret                                   SecretOperations
	SecretReference                          SecretReferenceOperations
	Service                                  ServiceOperations
	ServiceBinding                           ServiceBindingOperations
	ServiceConsumeMap                        ServiceConsumeMapOperations
	ServiceEvent                             ServiceEventOperations
	ServiceExposeMap                         ServiceExposeMapOperations
	ServiceLink                              ServiceLinkOperations
	ServiceLog                               ServiceLogOperations
	ServiceProxy                             ServiceProxyOperations
	ServiceRestart                           ServiceRestartOperations
	ServiceUpgrade                           ServiceUpgradeOperations
	ServiceUpgradeStrategy                   ServiceUpgradeStrategyOperations
	ServicesPortRange                        ServicesPortRangeOperations
	SetProjectMembersInput                   SetProjectMembersInputOperations
	SetServiceLinksInput                     SetServiceLinksInputOperations
	Setting                                  SettingOperations
	Snapshot                                 SnapshotOperations
	SnapshotBackupInput                      SnapshotBackupInputOperations
	Stack                                    StackOperations
	StackUpgrade                             StackUpgradeOperations
	StateTransition                          StateTransitionOperations
	StatsAccess                              StatsAccessOperations
	StorageDriver                            StorageDriverOperations
	StorageDriverService                     StorageDriverServiceOperations
	StoragePool                              StoragePoolOperations
	Subnet                                   SubnetOperations
	TargetPortRule                           TargetPortRuleOperations
	Task                                     TaskOperations
	TaskInstance                             TaskInstanceOperations
	ToServiceUpgradeStrategy                 ToServiceUpgradeStrategyOperations
	TypeDocumentation                        TypeDocumentationOperations
	Ulimit                                   UlimitOperations
	UserPreference                           UserPreferenceOperations
	VirtualMachine                           VirtualMachineOperations
	VirtualMachineDisk                       VirtualMachineDiskOperations
	Volume                                   VolumeOperations
	VolumeActivateInput                      VolumeActivateInputOperations
	VolumeSnapshotInput                      VolumeSnapshotInputOperations
	VolumeTemplate                           VolumeTemplateOperations
}

func constructClient(rancherBaseClient *RancherBaseClientImpl) *RancherClient {
	client := &RancherClient{
		RancherBaseClient: rancherBaseClient,
	}

	client.Account = newAccountClient(client)
	client.ActiveSetting = newActiveSettingClient(client)
	client.AddOutputsInput = newAddOutputsInputClient(client)
	client.AddRemoveServiceLinkInput = newAddRemoveServiceLinkInputClient(client)
	client.Agent = newAgentClient(client)
	client.Amazonec2Config = newAmazonec2ConfigClient(client)
	client.ApiKey = newApiKeyClient(client)
	client.AuditLog = newAuditLogClient(client)
	client.AzureConfig = newAzureConfigClient(client)
	client.Azureadconfig = newAzureadconfigClient(client)
	client.Backup = newBackupClient(client)
	client.BackupTarget = newBackupTargetClient(client)
	client.BaseMachineConfig = newBaseMachineConfigClient(client)
	client.Binding = newBindingClient(client)
	client.BlkioDeviceOption = newBlkioDeviceOptionClient(client)
	client.CatalogTemplate = newCatalogTemplateClient(client)
	client.Certificate = newCertificateClient(client)
	client.ChangeSecretInput = newChangeSecretInputClient(client)
	client.ClusterMembership = newClusterMembershipClient(client)
	client.ComposeConfig = newComposeConfigClient(client)
	client.ComposeConfigInput = newComposeConfigInputClient(client)
	client.ComposeProject = newComposeProjectClient(client)
	client.ComposeService = newComposeServiceClient(client)
	client.ConfigItem = newConfigItemClient(client)
	client.ConfigItemStatus = newConfigItemStatusClient(client)
	client.Container = newContainerClient(client)
	client.ContainerEvent = newContainerEventClient(client)
	client.ContainerExec = newContainerExecClient(client)
	client.ContainerLogs = newContainerLogsClient(client)
	client.ContainerProxy = newContainerProxyClient(client)
	client.Credential = newCredentialClient(client)
	client.Databasechangelog = newDatabasechangelogClient(client)
	client.Databasechangeloglock = newDatabasechangeloglockClient(client)
	client.DefaultNetwork = newDefaultNetworkClient(client)
	client.DigitaloceanConfig = newDigitaloceanConfigClient(client)
	client.DnsService = newDnsServiceClient(client)
	client.DockerBuild = newDockerBuildClient(client)
	client.ExtensionImplementation = newExtensionImplementationClient(client)
	client.ExtensionPoint = newExtensionPointClient(client)
	client.ExternalDnsEvent = newExternalDnsEventClient(client)
	client.ExternalEvent = newExternalEventClient(client)
	client.ExternalHandler = newExternalHandlerClient(client)
	client.ExternalHandlerExternalHandlerProcessMap = newExternalHandlerExternalHandlerProcessMapClient(client)
	client.ExternalHandlerProcess = newExternalHandlerProcessClient(client)
	client.ExternalHandlerProcessConfig = newExternalHandlerProcessConfigClient(client)
	client.ExternalHostEvent = newExternalHostEventClient(client)
	client.ExternalService = newExternalServiceClient(client)
	client.ExternalServiceEvent = newExternalServiceEventClient(client)
	client.ExternalStoragePoolEvent = newExternalStoragePoolEventClient(client)
	client.ExternalVolumeEvent = newExternalVolumeEventClient(client)
	client.FieldDocumentation = newFieldDocumentationClient(client)
	client.GenericObject = newGenericObjectClient(client)
	client.HaConfig = newHaConfigClient(client)
	client.HaConfigInput = newHaConfigInputClient(client)
	client.HealthcheckInstanceHostMap = newHealthcheckInstanceHostMapClient(client)
	client.Host = newHostClient(client)
	client.HostAccess = newHostAccessClient(client)
	client.HostApiProxyToken = newHostApiProxyTokenClient(client)
	client.HostTemplate = newHostTemplateClient(client)
	client.Identity = newIdentityClient(client)
	client.Image = newImageClient(client)
	client.InServiceUpgradeStrategy = newInServiceUpgradeStrategyClient(client)
	client.Instance = newInstanceClient(client)
	client.InstanceConsole = newInstanceConsoleClient(client)
	client.InstanceConsoleInput = newInstanceConsoleInputClient(client)
	client.InstanceHealthCheck = newInstanceHealthCheckClient(client)
	client.InstanceLink = newInstanceLinkClient(client)
	client.InstanceStop = newInstanceStopClient(client)
	client.IpAddress = newIpAddressClient(client)
	client.KubernetesService = newKubernetesServiceClient(client)
	client.KubernetesStack = newKubernetesStackClient(client)
	client.KubernetesStackUpgrade = newKubernetesStackUpgradeClient(client)
	client.Label = newLabelClient(client)
	client.LaunchConfig = newLaunchConfigClient(client)
	client.LbConfig = newLbConfigClient(client)
	client.LbTargetConfig = newLbTargetConfigClient(client)
	client.LoadBalancerCookieStickinessPolicy = newLoadBalancerCookieStickinessPolicyClient(client)
	client.LoadBalancerService = newLoadBalancerServiceClient(client)
	client.LocalAuthConfig = newLocalAuthConfigClient(client)
	client.LogConfig = newLogConfigClient(client)
	client.Machine = newMachineClient(client)
	client.MachineDriver = newMachineDriverClient(client)
	client.Mount = newMountClient(client)
	client.MountEntry = newMountEntryClient(client)
	client.Network = newNetworkClient(client)
	client.NetworkDriver = newNetworkDriverClient(client)
	client.NetworkDriverService = newNetworkDriverServiceClient(client)
	client.NetworkPolicyRule = newNetworkPolicyRuleClient(client)
	client.NetworkPolicyRuleBetween = newNetworkPolicyRuleBetweenClient(client)
	client.NetworkPolicyRuleMember = newNetworkPolicyRuleMemberClient(client)
	client.NetworkPolicyRuleWithin = newNetworkPolicyRuleWithinClient(client)
	client.NfsConfig = newNfsConfigClient(client)
	client.Openldapconfig = newOpenldapconfigClient(client)
	client.PacketConfig = newPacketConfigClient(client)
	client.Password = newPasswordClient(client)
	client.PhysicalHost = newPhysicalHostClient(client)
	client.Port = newPortClient(client)
	client.PortRule = newPortRuleClient(client)
	client.ProcessDefinition = newProcessDefinitionClient(client)
	client.ProcessExecution = newProcessExecutionClient(client)
	client.ProcessInstance = newProcessInstanceClient(client)
	client.ProcessPool = newProcessPoolClient(client)
	client.ProcessSummary = newProcessSummaryClient(client)
	client.Project = newProjectClient(client)
	client.ProjectMember = newProjectMemberClient(client)
	client.ProjectTemplate = newProjectTemplateClient(client)
	client.PublicEndpoint = newPublicEndpointClient(client)
	client.Publish = newPublishClient(client)
	client.PullTask = newPullTaskClient(client)
	client.RecreateOnQuorumStrategyConfig = newRecreateOnQuorumStrategyConfigClient(client)
	client.Register = newRegisterClient(client)
	client.RegistrationToken = newRegistrationTokenClient(client)
	client.Registry = newRegistryClient(client)
	client.RegistryCredential = newRegistryCredentialClient(client)
	client.ResourceDefinition = newResourceDefinitionClient(client)
	client.RestartPolicy = newRestartPolicyClient(client)
	client.RestoreFromBackupInput = newRestoreFromBackupInputClient(client)
	client.RevertToSnapshotInput = newRevertToSnapshotInputClient(client)
	client.RollingRestartStrategy = newRollingRestartStrategyClient(client)
	client.ScalePolicy = newScalePolicyClient(client)
	client.ScheduledUpgrade = newScheduledUpgradeClient(client)
	client.SecondaryLaunchConfig = newSecondaryLaunchConfigClient(client)
	client.Secret = newSecretClient(client)
	client.SecretReference = newSecretReferenceClient(client)
	client.Service = newServiceClient(client)
	client.ServiceBinding = newServiceBindingClient(client)
	client.ServiceConsumeMap = newServiceConsumeMapClient(client)
	client.ServiceEvent = newServiceEventClient(client)
	client.ServiceExposeMap = newServiceExposeMapClient(client)
	client.ServiceLink = newServiceLinkClient(client)
	client.ServiceLog = newServiceLogClient(client)
	client.ServiceProxy = newServiceProxyClient(client)
	client.ServiceRestart = newServiceRestartClient(client)
	client.ServiceUpgrade = newServiceUpgradeClient(client)
	client.ServiceUpgradeStrategy = newServiceUpgradeStrategyClient(client)
	client.ServicesPortRange = newServicesPortRangeClient(client)
	client.SetProjectMembersInput = newSetProjectMembersInputClient(client)
	client.SetServiceLinksInput = newSetServiceLinksInputClient(client)
	client.Setting = newSettingClient(client)
	client.Snapshot = newSnapshotClient(client)
	client.SnapshotBackupInput = newSnapshotBackupInputClient(client)
	client.Stack = newStackClient(client)
	client.StackUpgrade = newStackUpgradeClient(client)
	client.StateTransition = newStateTransitionClient(client)
	client.StatsAccess = newStatsAccessClient(client)
	client.StorageDriver = newStorageDriverClient(client)
	client.StorageDriverService = newStorageDriverServiceClient(client)
	client.StoragePool = newStoragePoolClient(client)
	client.Subnet = newSubnetClient(client)
	client.TargetPortRule = newTargetPortRuleClient(client)
	client.Task = newTaskClient(client)
	client.TaskInstance = newTaskInstanceClient(client)
	client.ToServiceUpgradeStrategy = newToServiceUpgradeStrategyClient(client)
	client.TypeDocumentation = newTypeDocumentationClient(client)
	client.Ulimit = newUlimitClient(client)
	client.UserPreference = newUserPreferenceClient(client)
	client.VirtualMachine = newVirtualMachineClient(client)
	client.VirtualMachineDisk = newVirtualMachineDiskClient(client)
	client.Volume = newVolumeClient(client)
	client.VolumeActivateInput = newVolumeActivateInputClient(client)
	client.VolumeSnapshotInput = newVolumeSnapshotInputClient(client)
	client.VolumeTemplate = newVolumeTemplateClient(client)

	return client
}

func NewRancherClient(opts *ClientOpts) (*RancherClient, error) {
	rancherBaseClient := &RancherBaseClientImpl{
		Types: map[string]Schema{},
	}
	client := constructClient(rancherBaseClient)

	err := setupRancherBaseClient(rancherBaseClient, opts)
	if err != nil {
		return nil, err
	}

	return client, nil
}
