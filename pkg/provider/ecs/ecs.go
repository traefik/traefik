package ecs

import (
	"context"
	"fmt"
	"iter"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/cenkalti/backoff/v4"
	"github.com/patrickmn/go-cache"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/traefik/v3/pkg/config/dynamic"
	"github.com/traefik/traefik/v3/pkg/job"
	"github.com/traefik/traefik/v3/pkg/logs"
	"github.com/traefik/traefik/v3/pkg/provider"
	"github.com/traefik/traefik/v3/pkg/safe"
)

// Provider holds configurations of the provider.
type Provider struct {
	Constraints      string `description:"Constraints is an expression that Traefik matches against the container's labels to determine whether to create any route for that container." json:"constraints,omitempty" toml:"constraints,omitempty" yaml:"constraints,omitempty" export:"true"`
	ExposedByDefault bool   `description:"Expose services by default." json:"exposedByDefault,omitempty" toml:"exposedByDefault,omitempty" yaml:"exposedByDefault,omitempty" export:"true"`
	RefreshSeconds   int    `description:"Polling interval (in seconds)." json:"refreshSeconds,omitempty" toml:"refreshSeconds,omitempty" yaml:"refreshSeconds,omitempty" export:"true"`
	DefaultRule      string `description:"Default rule." json:"defaultRule,omitempty" toml:"defaultRule,omitempty" yaml:"defaultRule,omitempty"`

	// Provider lookup parameters.
	Clusters             []string `description:"ECS Cluster names." json:"clusters,omitempty" toml:"clusters,omitempty" yaml:"clusters,omitempty" export:"true"`
	AutoDiscoverClusters bool     `description:"Auto discover cluster." json:"autoDiscoverClusters,omitempty" toml:"autoDiscoverClusters,omitempty" yaml:"autoDiscoverClusters,omitempty" export:"true"`
	HealthyTasksOnly     bool     `description:"Determines whether to discover only healthy tasks." json:"healthyTasksOnly,omitempty" toml:"healthyTasksOnly,omitempty" yaml:"healthyTasksOnly,omitempty" export:"true"`
	ECSAnywhere          bool     `description:"Enable ECS Anywhere support." json:"ecsAnywhere,omitempty" toml:"ecsAnywhere,omitempty" yaml:"ecsAnywhere,omitempty" export:"true"`
	Region               string   `description:"AWS region to use for requests."  json:"region,omitempty" toml:"region,omitempty" yaml:"region,omitempty" export:"true"`
	AccessKeyID          string   `description:"AWS credentials access key ID to use for making requests." json:"accessKeyID,omitempty" toml:"accessKeyID,omitempty" yaml:"accessKeyID,omitempty" loggable:"false"`
	SecretAccessKey      string   `description:"AWS credentials access key to use for making requests." json:"secretAccessKey,omitempty" toml:"secretAccessKey,omitempty" yaml:"secretAccessKey,omitempty" loggable:"false"`
	defaultRuleTpl       *template.Template
}

type ecsInstance struct {
	Name                string
	ID                  string
	containerDefinition *ecstypes.ContainerDefinition
	machine             *machine
	Labels              map[string]string
	ExtraConf           configuration
}

type portMapping struct {
	containerPort int32
	hostPort      int32
	protocol      ecstypes.TransportProtocol
}

type machine struct {
	state        ec2types.InstanceStateName
	privateIP    string
	ports        []portMapping
	healthStatus ecstypes.HealthStatus
}

type awsClient struct {
	ecs *ecs.Client
	ec2 *ec2.Client
	ssm *ssm.Client
}

// DefaultTemplateRule The default template for the default rule.
const DefaultTemplateRule = "Host(`{{ normalize .Name }}`)"

var (
	_                    provider.Provider = (*Provider)(nil)
	existingTaskDefCache                   = cache.New(30*time.Minute, 5*time.Minute)
)

// SetDefaults sets the default values.
func (p *Provider) SetDefaults() {
	p.Clusters = []string{"default"}
	p.AutoDiscoverClusters = false
	p.HealthyTasksOnly = false
	p.ExposedByDefault = true
	p.RefreshSeconds = 15
	p.DefaultRule = DefaultTemplateRule
}

// Init the provider.
func (p *Provider) Init() error {
	defaultRuleTpl, err := provider.MakeDefaultRuleTemplate(p.DefaultRule, nil)
	if err != nil {
		return fmt.Errorf("error while parsing default rule: %w", err)
	}

	p.defaultRuleTpl = defaultRuleTpl
	return nil
}

func (p *Provider) createClient(ctx context.Context, logger zerolog.Logger) (*awsClient, error) {
	optFns := []func(*config.LoadOptions) error{
		config.WithLogger(logs.NewAWSWrapper(logger)),
	}
	if p.Region != "" {
		optFns = append(optFns, config.WithRegion(p.Region))
	} else {
		logger.Info().Msg("No region provided, will retrieve region from the EC2 Metadata service")
		optFns = append(optFns, config.WithEC2IMDSRegion())
	}

	if p.AccessKeyID != "" && p.SecretAccessKey != "" {
		// From https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html#specify-credentials-programmatically:
		//   "If you explicitly provide credentials, as in this example, the SDK uses only those credentials."
		// this makes sure that user-defined credentials always have the highest priority
		staticCreds := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(p.AccessKeyID, p.SecretAccessKey, ""))
		optFns = append(optFns, config.WithCredentialsProvider(staticCreds))

		// If the access key and secret access key are not provided, config.LoadDefaultConfig
		// will look for the credentials in the default credential chain.
		// See https://docs.aws.amazon.com/sdk-for-go/v2/developer-guide/configure-gosdk.html#specifying-credentials.
	}

	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, err
	}

	return &awsClient{
		ecs.NewFromConfig(cfg),
		ec2.NewFromConfig(cfg),
		ssm.NewFromConfig(cfg),
	}, nil
}

// Provide configuration to traefik from ECS.
func (p *Provider) Provide(configurationChan chan<- dynamic.Message, pool *safe.Pool) error {
	pool.GoCtx(func(routineCtx context.Context) {
		logger := log.Ctx(routineCtx).With().Str(logs.ProviderName, "ecs").Logger()
		ctxLog := logger.WithContext(routineCtx)

		operation := func() error {
			awsClient, err := p.createClient(ctxLog, logger)
			if err != nil {
				return fmt.Errorf("unable to create AWS client: %w", err)
			}

			err = p.loadConfiguration(ctxLog, awsClient, configurationChan)
			if err != nil {
				return fmt.Errorf("failed to get ECS configuration: %w", err)
			}

			ticker := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					err = p.loadConfiguration(ctxLog, awsClient, configurationChan)
					if err != nil {
						return fmt.Errorf("failed to refresh ECS configuration: %w", err)
					}

				case <-routineCtx.Done():
					return nil
				}
			}
		}

		notify := func(err error, time time.Duration) {
			logger.Error().Err(err).Msgf("Provider error, retrying in %s", time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), backoff.WithContext(job.NewBackOff(backoff.NewExponentialBackOff()), routineCtx), notify)
		if err != nil {
			logger.Error().Err(err).Msg("Cannot retrieve data")
		}
	})

	return nil
}

func (p *Provider) loadConfiguration(ctx context.Context, client *awsClient, configurationChan chan<- dynamic.Message) error {
	instances, err := p.listInstances(ctx, client)
	if err != nil {
		return err
	}

	configurationChan <- dynamic.Message{
		ProviderName:  "ecs",
		Configuration: p.buildConfiguration(ctx, instances),
	}

	return nil
}

// Find all running Provider tasks in a cluster, also collect the task definitions (for docker labels)
// and the EC2 instance data.
func (p *Provider) listInstances(ctx context.Context, client *awsClient) ([]ecsInstance, error) {
	logger := log.Ctx(ctx)

	var clusters []string

	if p.AutoDiscoverClusters {
		input := &ecs.ListClustersInput{}

		paginator := ecs.NewListClustersPaginator(client.ecs, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, err
			}

			clusters = append(clusters, page.ClusterArns...)
		}
	} else {
		clusters = p.Clusters
	}

	var instances []ecsInstance

	logger.Debug().Msgf("ECS Clusters: %s", clusters)
	for _, c := range clusters {
		input := &ecs.ListTasksInput{
			Cluster:       &c,
			DesiredStatus: ecstypes.DesiredStatusRunning,
		}

		tasks := make(map[string]ecstypes.Task)

		paginator := ecs.NewListTasksPaginator(client.ecs, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				return nil, fmt.Errorf("listing tasks: %w", err)
			}
			if len(page.TaskArns) > 0 {
				resp, err := client.ecs.DescribeTasks(ctx, &ecs.DescribeTasksInput{
					Tasks:   page.TaskArns,
					Cluster: &c,
				})
				if err != nil {
					logger.Error().Msgf("Unable to describe tasks for %v", page.TaskArns)
				} else {
					for _, t := range resp.Tasks {
						if p.HealthyTasksOnly && t.HealthStatus != ecstypes.HealthStatusHealthy {
							logger.Debug().Msgf("Skipping unhealthy task %s", aws.ToString(t.TaskArn))
							continue
						}

						tasks[aws.ToString(t.TaskArn)] = t
					}
				}
			}
		}

		// Skip to the next cluster if there are no tasks found on
		// this cluster.
		if len(tasks) == 0 {
			continue
		}

		ec2Instances, err := p.lookupEc2Instances(ctx, client, &c, tasks)
		if err != nil {
			return nil, err
		}

		miInstances := make(map[string]ssmtypes.InstanceInformation)
		if p.ECSAnywhere {
			// Try looking up for instances on ECS Anywhere
			miInstances, err = p.lookupMiInstances(ctx, client, &c, tasks)
			if err != nil {
				return nil, err
			}
		}

		taskDefinitions, err := p.lookupTaskDefinitions(ctx, client, tasks)
		if err != nil {
			return nil, err
		}

		for key, task := range tasks {
			containerInstance, hasContainerInstance := ec2Instances[aws.ToString(task.ContainerInstanceArn)]
			taskDef := taskDefinitions[key]

			for _, container := range task.Containers {
				var containerDefinition *ecstypes.ContainerDefinition
				for _, def := range taskDef.ContainerDefinitions {
					if aws.ToString(container.Name) == aws.ToString(def.Name) {
						containerDefinition = &def
						break
					}
				}

				if containerDefinition == nil {
					logger.Debug().Msgf("Unable to find container definition for %s", aws.ToString(container.Name))
					continue
				}

				var mach *machine
				if taskDef.NetworkMode == ecstypes.NetworkModeAwsvpc && len(task.Attachments) != 0 {
					if len(container.NetworkInterfaces) == 0 {
						logger.Error().Msgf("Skip container %s: no network interfaces", aws.ToString(container.Name))
						continue
					}

					var ports []portMapping
					for _, mapping := range containerDefinition.PortMappings {
						ports = append(ports, portMapping{
							hostPort:      aws.ToInt32(mapping.HostPort),
							containerPort: aws.ToInt32(mapping.ContainerPort),
							protocol:      mapping.Protocol,
						})
					}
					mach = &machine{
						privateIP:    aws.ToString(container.NetworkInterfaces[0].PrivateIpv4Address),
						ports:        ports,
						state:        ec2types.InstanceStateName(strings.ToLower(aws.ToString(task.LastStatus))),
						healthStatus: task.HealthStatus,
					}
				} else {
					miContainerInstance, hasMiContainerInstance := miInstances[aws.ToString(task.ContainerInstanceArn)]
					if !hasContainerInstance && !hasMiContainerInstance {
						logger.Error().Msgf("Unable to find container instance information for %s", aws.ToString(container.Name))
						continue
					}

					var ports []portMapping
					for _, mapping := range container.NetworkBindings {
						ports = append(ports, portMapping{
							hostPort:      aws.ToInt32(mapping.HostPort),
							containerPort: aws.ToInt32(mapping.ContainerPort),
							protocol:      mapping.Protocol,
						})
					}
					var privateIPAddress string
					var stateName ec2types.InstanceStateName
					if hasContainerInstance {
						privateIPAddress = aws.ToString(containerInstance.PrivateIpAddress)
						stateName = containerInstance.State.Name
					} else if hasMiContainerInstance {
						privateIPAddress = aws.ToString(miContainerInstance.IPAddress)
						stateName = ec2types.InstanceStateName(strings.ToLower(aws.ToString(task.LastStatus)))
					}

					mach = &machine{
						privateIP: privateIPAddress,
						ports:     ports,
						state:     stateName,
					}
				}

				instance := ecsInstance{
					Name:                fmt.Sprintf("%s-%s", strings.Replace(aws.ToString(task.Group), ":", "-", 1), aws.ToString(container.Name)),
					ID:                  key[len(key)-12:],
					containerDefinition: containerDefinition,
					machine:             mach,
					Labels:              containerDefinition.DockerLabels,
				}

				extraConf, err := p.getConfiguration(instance)
				if err != nil {
					logger.Error().Err(err).Msgf("Skip container %s", getServiceName(instance))
					continue
				}
				instance.ExtraConf = extraConf

				instances = append(instances, instance)
			}
		}
	}

	return instances, nil
}

func (p *Provider) lookupMiInstances(ctx context.Context, client *awsClient, clusterName *string, ecsDatas map[string]ecstypes.Task) (map[string]ssmtypes.InstanceInformation, error) {
	instanceIDs := make(map[string]string)
	miInstances := make(map[string]ssmtypes.InstanceInformation)

	var containerInstancesArns []string
	var instanceArns []string

	for _, task := range ecsDatas {
		if task.ContainerInstanceArn != nil {
			containerInstancesArns = append(containerInstancesArns, *task.ContainerInstanceArn)
		}
	}

	for arns := range chunkIDs(containerInstancesArns) {
		resp, err := client.ecs.DescribeContainerInstances(ctx, &ecs.DescribeContainerInstancesInput{
			ContainerInstances: arns,
			Cluster:            clusterName,
		})
		if err != nil {
			return nil, fmt.Errorf("describing container instances: %w", err)
		}

		for _, container := range resp.ContainerInstances {
			instanceIDs[aws.ToString(container.Ec2InstanceId)] = aws.ToString(container.ContainerInstanceArn)

			// Disallow EC2 Instance IDs
			// This prevents considering EC2 instances in ECS
			// and getting InvalidInstanceID.Malformed error when calling the describe-instances endpoint.
			if strings.HasPrefix(aws.ToString(container.Ec2InstanceId), "mi-") {
				instanceArns = append(instanceArns, *container.Ec2InstanceId)
			}
		}
	}

	if len(instanceArns) > 0 {
		for ids := range chunkIDs(instanceArns) {
			input := &ssm.DescribeInstanceInformationInput{
				Filters: []ssmtypes.InstanceInformationStringFilter{
					{
						Key:    aws.String("InstanceIds"),
						Values: ids,
					},
				},
			}

			paginator := ssm.NewDescribeInstanceInformationPaginator(client.ssm, input)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if err != nil {
					return nil, fmt.Errorf("describing instances: %w", err)
				}

				for _, i := range page.InstanceInformationList {
					if i.InstanceId != nil {
						miInstances[instanceIDs[aws.ToString(i.InstanceId)]] = i
					}
				}
			}
		}
	}

	return miInstances, nil
}

func (p *Provider) lookupEc2Instances(ctx context.Context, client *awsClient, clusterName *string, ecsDatas map[string]ecstypes.Task) (map[string]ec2types.Instance, error) {
	instanceIDs := make(map[string]string)
	ec2Instances := make(map[string]ec2types.Instance)

	var containerInstancesArns []string
	var instanceArns []string

	for _, task := range ecsDatas {
		if task.ContainerInstanceArn != nil {
			containerInstancesArns = append(containerInstancesArns, *task.ContainerInstanceArn)
		}
	}

	for arns := range chunkIDs(containerInstancesArns) {
		resp, err := client.ecs.DescribeContainerInstances(ctx, &ecs.DescribeContainerInstancesInput{
			ContainerInstances: arns,
			Cluster:            clusterName,
		})
		if err != nil {
			return nil, fmt.Errorf("describing container instances: %w", err)
		}

		for _, container := range resp.ContainerInstances {
			instanceIDs[aws.ToString(container.Ec2InstanceId)] = aws.ToString(container.ContainerInstanceArn)
			// Disallow Instance IDs of the form mi-*
			// This prevents considering external instances in ECS Anywhere setups
			// and getting InvalidInstanceID.Malformed error when calling the describe-instances endpoint.
			if strings.HasPrefix(aws.ToString(container.Ec2InstanceId), "mi-") {
				continue
			}
			if container.Ec2InstanceId != nil {
				instanceArns = append(instanceArns, *container.Ec2InstanceId)
			}
		}
	}

	if len(instanceArns) > 0 {
		for ids := range chunkIDs(instanceArns) {
			input := &ec2.DescribeInstancesInput{
				InstanceIds: ids,
			}

			paginator := ec2.NewDescribeInstancesPaginator(client.ec2, input)
			for paginator.HasMorePages() {
				page, err := paginator.NextPage(ctx)
				if err != nil {
					return nil, fmt.Errorf("describing instances: %w", err)
				}
				for _, r := range page.Reservations {
					for _, i := range r.Instances {
						if i.InstanceId != nil {
							ec2Instances[instanceIDs[aws.ToString(i.InstanceId)]] = i
						}
					}
				}
			}
		}
	}

	return ec2Instances, nil
}

func (p *Provider) lookupTaskDefinitions(ctx context.Context, client *awsClient, taskDefArns map[string]ecstypes.Task) (map[string]*ecstypes.TaskDefinition, error) {
	logger := log.Ctx(ctx)
	taskDef := make(map[string]*ecstypes.TaskDefinition)

	for arn, task := range taskDefArns {
		if definition, ok := existingTaskDefCache.Get(arn); ok {
			taskDef[arn] = definition.(*ecstypes.TaskDefinition)
			logger.Debug().Msgf("Found cached task definition for %s. Skipping the call", arn)
		} else {
			resp, err := client.ecs.DescribeTaskDefinition(ctx, &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: task.TaskDefinitionArn,
			})
			if err != nil {
				return nil, fmt.Errorf("describing task definition: %w", err)
			}

			taskDef[arn] = resp.TaskDefinition
			existingTaskDefCache.Set(arn, resp.TaskDefinition, cache.DefaultExpiration)
		}
	}
	return taskDef, nil
}

// chunkIDs ECS expects no more than 100 parameters be passed to a API call;
// thus, pack each string into an array capped at 100 elements.
func chunkIDs(ids []string) iter.Seq[[]string] {
	return slices.Chunk(ids, 100)
}
