package ecs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/patrickmn/go-cache"
)

var _ provider.Provider = (*Provider)(nil)
var existingTaskDefCache = cache.New(30*time.Minute, 5*time.Minute)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`

	Domain           string `description:"Default domain used"`
	ExposedByDefault bool   `description:"Expose containers by default" export:"true"`
	RefreshSeconds   int    `description:"Polling interval (in seconds)" export:"true"`

	// Provider lookup parameters
	Clusters             Clusters `description:"ECS Clusters name"`
	Cluster              string   `description:"deprecated - ECS Cluster name"` // deprecated
	AutoDiscoverClusters bool     `description:"Auto discover cluster" export:"true"`
	Region               string   `description:"The AWS region to use for requests" export:"true"`
	AccessKeyID          string   `description:"The AWS credentials access key to use for making requests"`
	SecretAccessKey      string   `description:"The AWS credentials access key to use for making requests"`
}

type ecsInstance struct {
	Name                string
	ID                  string
	containerDefinition *ecs.ContainerDefinition
	machine             *machine
	TraefikLabels       map[string]string
	SegmentLabels       map[string]string
	SegmentName         string
}

type portMapping struct {
	containerPort int64
	hostPort      int64
}

type machine struct {
	name      string
	state     string
	privateIP string
	ports     []portMapping
}

type awsClient struct {
	ecs *ecs.ECS
	ec2 *ec2.EC2
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	return p.BaseProvider.Init(constraints)
}

func (p *Provider) createClient() (*awsClient, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}

	ec2meta := ec2metadata.New(sess)
	if p.Region == "" {
		log.Infoln("No EC2 region provided, querying instance metadata endpoint...")
		identity, err := ec2meta.GetInstanceIdentityDocument()
		if err != nil {
			return nil, err
		}
		p.Region = identity.Region
	}

	cfg := &aws.Config{
		Region: &p.Region,
		Credentials: credentials.NewChainCredentials(
			[]credentials.Provider{
				&credentials.StaticProvider{
					Value: credentials.Value{
						AccessKeyID:     p.AccessKeyID,
						SecretAccessKey: p.SecretAccessKey,
					},
				},
				&credentials.EnvProvider{},
				&credentials.SharedCredentialsProvider{},
				defaults.RemoteCredProvider(*(defaults.Config()), defaults.Handlers()),
			}),
	}

	if p.Trace {
		cfg.WithLogger(aws.LoggerFunc(func(args ...interface{}) {
			log.Debug(args...)
		}))
	}

	return &awsClient{
		ecs.New(sess, cfg),
		ec2.New(sess, cfg),
	}, nil
}

// Provide allows the ecs provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	handleCanceled := func(ctx context.Context, err error) error {
		if ctx.Err() == context.Canceled || err == context.Canceled {
			return nil
		}
		return err
	}

	pool.Go(func(stop chan bool) {
		ctx, cancel := context.WithCancel(context.Background())
		safe.Go(func() {
			<-stop
			cancel()
		})

		operation := func() error {
			awsClient, err := p.createClient()
			if err != nil {
				return err
			}

			configuration, err := p.loadECSConfig(ctx, awsClient)
			if err != nil {
				return handleCanceled(ctx, err)
			}

			configurationChan <- types.ConfigMessage{
				ProviderName:  "ecs",
				Configuration: configuration,
			}

			if p.Watch {
				reload := time.NewTicker(time.Second * time.Duration(p.RefreshSeconds))
				defer reload.Stop()
				for {
					select {
					case <-reload.C:
						configuration, err := p.loadECSConfig(ctx, awsClient)
						if err != nil {
							return handleCanceled(ctx, err)
						}

						configurationChan <- types.ConfigMessage{
							ProviderName:  "ecs",
							Configuration: configuration,
						}
					case <-ctx.Done():
						return handleCanceled(ctx, ctx.Err())
					}
				}
			}

			return nil
		}

		notify := func(err error, time time.Duration) {
			log.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to Provider api %+v", err)
		}
	})

	return nil
}

// Find all running Provider tasks in a cluster, also collect the task definitions (for docker labels)
// and the EC2 instance data
func (p *Provider) listInstances(ctx context.Context, client *awsClient) ([]ecsInstance, error) {
	var clustersArn []*string
	var clusters Clusters

	if p.AutoDiscoverClusters {
		input := &ecs.ListClustersInput{}
		for {
			result, err := client.ecs.ListClusters(input)
			if err != nil {
				return nil, err
			}
			if result != nil {
				clustersArn = append(clustersArn, result.ClusterArns...)
				input.NextToken = result.NextToken
				if result.NextToken == nil {
					break
				}
			} else {
				break
			}
		}
		for _, cArn := range clustersArn {
			clusters = append(clusters, *cArn)
		}
	} else if p.Cluster != "" {
		// TODO: Deprecated configuration - Need to be removed in the future
		clusters = Clusters{p.Cluster}
		log.Warn("Deprecated configuration found: ecs.cluster. Please use ecs.clusters instead.")
	} else {
		clusters = p.Clusters
	}

	var instances []ecsInstance

	log.Debugf("ECS Clusters: %s", clusters)

	for _, c := range clusters {

		input := &ecs.ListTasksInput{
			Cluster:       &c,
			DesiredStatus: aws.String(ecs.DesiredStatusRunning),
		}
		tasks := make(map[string]*ecs.Task)
		err := client.ecs.ListTasksPagesWithContext(ctx, input, func(page *ecs.ListTasksOutput, lastPage bool) bool {
			if len(page.TaskArns) > 0 {
				resp, err := client.ecs.DescribeTasksWithContext(ctx, &ecs.DescribeTasksInput{
					Tasks:   page.TaskArns,
					Cluster: &c,
				})
				if err != nil {
					log.Errorf("Unable to describe tasks for %v", page.TaskArns)
				} else {
					for _, t := range resp.Tasks {
						if aws.StringValue(t.LastStatus) == ecs.DesiredStatusRunning {
							tasks[aws.StringValue(t.TaskArn)] = t
						}
					}
				}
			}
			return !lastPage
		})

		if err != nil {
			log.Error("Unable to list tasks")
			return nil, err
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

		taskDefinitions, err := p.lookupTaskDefinitions(ctx, client, tasks)
		if err != nil {
			return nil, err
		}

		for key, task := range tasks {

			containerInstance := ec2Instances[aws.StringValue(task.ContainerInstanceArn)]
			taskDef := taskDefinitions[key]

			for _, container := range task.Containers {

				var containerDefinition *ecs.ContainerDefinition
				for _, def := range taskDef.ContainerDefinitions {
					if aws.StringValue(container.Name) == aws.StringValue(def.Name) {
						containerDefinition = def
						break
					}
				}

				if containerDefinition == nil {
					log.Debugf("Unable to find container definition for %s", aws.StringValue(container.Name))
					continue
				}

				var mach *machine
				if len(task.Attachments) != 0 {
					var ports []portMapping
					for _, mapping := range containerDefinition.PortMappings {
						if mapping != nil {
							ports = append(ports, portMapping{
								hostPort:      aws.Int64Value(mapping.HostPort),
								containerPort: aws.Int64Value(mapping.ContainerPort),
							})
						}
					}
					mach = &machine{
						privateIP: aws.StringValue(container.NetworkInterfaces[0].PrivateIpv4Address),
						ports:     ports,
						state:     aws.StringValue(task.LastStatus),
					}
				} else {
					var ports []portMapping
					for _, mapping := range container.NetworkBindings {
						if mapping != nil {
							ports = append(ports, portMapping{
								hostPort:      aws.Int64Value(mapping.HostPort),
								containerPort: aws.Int64Value(mapping.ContainerPort),
							})
						}
					}
					mach = &machine{
						privateIP: aws.StringValue(containerInstance.PrivateIpAddress),
						ports:     ports,
						state:     aws.StringValue(containerInstance.State.Name),
					}
				}

				instances = append(instances, ecsInstance{
					Name:                fmt.Sprintf("%s-%s", strings.Replace(aws.StringValue(task.Group), ":", "-", 1), *container.Name),
					ID:                  key[len(key)-12:],
					containerDefinition: containerDefinition,
					machine:             mach,
					TraefikLabels:       aws.StringValueMap(containerDefinition.DockerLabels),
				})
			}
		}
	}

	return instances, nil
}

func (p *Provider) lookupEc2Instances(ctx context.Context, client *awsClient, clusterName *string, ecsDatas map[string]*ecs.Task) (map[string]*ec2.Instance, error) {

	instanceIds := make(map[string]string)
	ec2Instances := make(map[string]*ec2.Instance)

	var containerInstancesArns []*string
	var instanceArns []*string

	for _, task := range ecsDatas {
		if task.ContainerInstanceArn != nil {
			containerInstancesArns = append(containerInstancesArns, task.ContainerInstanceArn)
		}
	}

	for _, arns := range p.chunkIDs(containerInstancesArns) {
		resp, err := client.ecs.DescribeContainerInstancesWithContext(ctx, &ecs.DescribeContainerInstancesInput{
			ContainerInstances: arns,
			Cluster:            clusterName,
		})

		if err != nil {
			log.Errorf("Unable to describe container instances: %v", err)
			return nil, err
		}

		for _, container := range resp.ContainerInstances {
			instanceIds[aws.StringValue(container.Ec2InstanceId)] = aws.StringValue(container.ContainerInstanceArn)
			instanceArns = append(instanceArns, container.Ec2InstanceId)
		}
	}

	if len(instanceArns) > 0 {
		for _, ids := range p.chunkIDs(instanceArns) {
			input := &ec2.DescribeInstancesInput{
				InstanceIds: ids,
			}

			err := client.ec2.DescribeInstancesPagesWithContext(ctx, input, func(page *ec2.DescribeInstancesOutput, lastPage bool) bool {
				if len(page.Reservations) > 0 {
					for _, r := range page.Reservations {
						for _, i := range r.Instances {
							if i.InstanceId != nil {
								ec2Instances[instanceIds[aws.StringValue(i.InstanceId)]] = i
							}
						}
					}
				}
				return !lastPage
			})

			if err != nil {
				log.Errorf("Unable to describe instances: %v", err)
				return nil, err
			}
		}
	}

	return ec2Instances, nil
}

func (p *Provider) lookupTaskDefinitions(ctx context.Context, client *awsClient, taskDefArns map[string]*ecs.Task) (map[string]*ecs.TaskDefinition, error) {
	taskDef := make(map[string]*ecs.TaskDefinition)
	for arn, task := range taskDefArns {
		if definition, ok := existingTaskDefCache.Get(arn); ok {
			taskDef[arn] = definition.(*ecs.TaskDefinition)
			log.Debugf("Found cached task definition for %s. Skipping the call", arn)
		} else {
			resp, err := client.ecs.DescribeTaskDefinitionWithContext(ctx, &ecs.DescribeTaskDefinitionInput{
				TaskDefinition: task.TaskDefinitionArn,
			})

			if err != nil {
				log.Errorf("Unable to describe task definition: %s", err)
				return nil, err
			}

			taskDef[arn] = resp.TaskDefinition
			existingTaskDefCache.Set(arn, resp.TaskDefinition, cache.DefaultExpiration)
		}
	}
	return taskDef, nil
}

func (p *Provider) loadECSConfig(ctx context.Context, client *awsClient) (*types.Configuration, error) {
	instances, err := p.listInstances(ctx, client)
	if err != nil {
		return nil, err
	}

	return p.buildConfiguration(instances)
}

// chunkIDs ECS expects no more than 100 parameters be passed to a API call;
// thus, pack each string into an array capped at 100 elements
func (p *Provider) chunkIDs(ids []*string) [][]*string {
	var chuncked [][]*string
	for i := 0; i < len(ids); i += 100 {
		sliceEnd := -1
		if i+100 < len(ids) {
			sliceEnd = i + 100
		} else {
			sliceEnd = len(ids)
		}
		chuncked = append(chuncked, ids[i:sliceEnd])
	}
	return chuncked
}
