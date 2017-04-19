package ecs

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash"`

	Domain           string `description:"Default domain used"`
	ExposedByDefault bool   `description:"Expose containers by default"`
	RefreshSeconds   int    `description:"Polling interval (in seconds)"`

	// Provider lookup parameters
	Cluster         string `description:"ECS Cluster Name"`
	Region          string `description:"The AWS region to use for requests"`
	AccessKeyID     string `description:"The AWS credentials access key to use for making requests"`
	SecretAccessKey string `description:"The AWS credentials access key to use for making requests"`
}

type ecsInstance struct {
	Name                string
	ID                  string
	task                *ecs.Task
	taskDefinition      *ecs.TaskDefinition
	container           *ecs.Container
	containerDefinition *ecs.ContainerDefinition
	machine             *ec2.Instance
}

type awsClient struct {
	ecs *ecs.ECS
	ec2 *ec2.EC2
}

func (p *Provider) createClient() (*awsClient, error) {
	sess := session.New()
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

	return &awsClient{
		ecs.New(sess, cfg),
		ec2.New(sess, cfg),
	}, nil
}

// Provide allows the ecs provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool, constraints types.Constraints) error {

	p.Constraints = append(p.Constraints, constraints...)

	handleCanceled := func(ctx context.Context, err error) error {
		if ctx.Err() == context.Canceled || err == context.Canceled {
			return nil
		}
		return err
	}

	pool.Go(func(stop chan bool) {
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			select {
			case <-stop:
				cancel()
			}
		}()

		operation := func() error {
			aws, err := p.createClient()
			if err != nil {
				return err
			}

			configuration, err := p.loadECSConfig(ctx, aws)
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
						configuration, err := p.loadECSConfig(ctx, aws)
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

func wrapAws(ctx context.Context, req *request.Request) error {
	req.HTTPRequest = req.HTTPRequest.WithContext(ctx)
	return req.Send()
}

func (p *Provider) loadECSConfig(ctx context.Context, client *awsClient) (*types.Configuration, error) {
	var ecsFuncMap = template.FuncMap{
		"filterFrontends": p.filterFrontends,
		"getFrontendRule": p.getFrontendRule,
	}

	instances, err := p.listInstances(ctx, client)
	if err != nil {
		return nil, err
	}

	instances = fun.Filter(p.filterInstance, instances).([]ecsInstance)

	return p.GetConfiguration("templates/ecs.tmpl", ecsFuncMap, struct {
		Instances []ecsInstance
	}{
		instances,
	})
}

// Find all running Provider tasks in a cluster, also collect the task definitions (for docker labels)
// and the EC2 instance data
func (p *Provider) listInstances(ctx context.Context, client *awsClient) ([]ecsInstance, error) {
	var taskArns []*string
	req, _ := client.ecs.ListTasksRequest(&ecs.ListTasksInput{
		Cluster:       &p.Cluster,
		DesiredStatus: aws.String(ecs.DesiredStatusRunning),
	})

	for ; req != nil; req = req.NextPage() {
		if err := wrapAws(ctx, req); err != nil {
			return nil, err
		}

		taskArns = append(taskArns, req.Data.(*ecs.ListTasksOutput).TaskArns...)
	}

	// Early return: if we can't list tasks we have nothing to
	// describe below - likely empty cluster/permissions are bad.  This
	// stops the AWS API from returning a 401 when you DescribeTasks
	// with no input.
	if len(taskArns) == 0 {
		return []ecsInstance{}, nil
	}

	chunkedTaskArns := p.chunkedTaskArns(taskArns)
	var tasks []*ecs.Task

	for _, arns := range chunkedTaskArns {
		req, taskResp := client.ecs.DescribeTasksRequest(&ecs.DescribeTasksInput{
			Tasks:   arns,
			Cluster: &p.Cluster,
		})

		if err := wrapAws(ctx, req); err != nil {
			return nil, err
		}
		tasks = append(tasks, taskResp.Tasks...)

	}

	containerInstanceArns := make([]*string, 0)
	byContainerInstance := make(map[string]int)

	taskDefinitionArns := make([]*string, 0)
	byTaskDefinition := make(map[string]int)

	for _, task := range tasks {
		if _, found := byContainerInstance[*task.ContainerInstanceArn]; !found {
			byContainerInstance[*task.ContainerInstanceArn] = len(containerInstanceArns)
			containerInstanceArns = append(containerInstanceArns, task.ContainerInstanceArn)
		}
		if _, found := byTaskDefinition[*task.TaskDefinitionArn]; !found {
			byTaskDefinition[*task.TaskDefinitionArn] = len(taskDefinitionArns)
			taskDefinitionArns = append(taskDefinitionArns, task.TaskDefinitionArn)
		}
	}

	machines, err := p.lookupEc2Instances(ctx, client, containerInstanceArns)
	if err != nil {
		return nil, err
	}

	taskDefinitions, err := p.lookupTaskDefinitions(ctx, client, taskDefinitionArns)
	if err != nil {
		return nil, err
	}

	var instances []ecsInstance
	for _, task := range tasks {

		machineIdx := byContainerInstance[*task.ContainerInstanceArn]
		taskDefIdx := byTaskDefinition[*task.TaskDefinitionArn]

		for _, container := range task.Containers {

			taskDefinition := taskDefinitions[taskDefIdx]
			var containerDefinition *ecs.ContainerDefinition
			for _, def := range taskDefinition.ContainerDefinitions {
				if *container.Name == *def.Name {
					containerDefinition = def
					break
				}
			}

			instances = append(instances, ecsInstance{
				fmt.Sprintf("%s-%s", strings.Replace(*task.Group, ":", "-", 1), *container.Name),
				(*task.TaskArn)[len(*task.TaskArn)-12:],
				task,
				taskDefinition,
				container,
				containerDefinition,
				machines[machineIdx],
			})
		}
	}

	return instances, nil
}

func (p *Provider) lookupEc2Instances(ctx context.Context, client *awsClient, containerArns []*string) ([]*ec2.Instance, error) {

	order := make(map[string]int)
	instanceIds := make([]*string, len(containerArns))
	instances := make([]*ec2.Instance, len(containerArns))
	for i, arn := range containerArns {
		order[*arn] = i
	}

	req, _ := client.ecs.DescribeContainerInstancesRequest(&ecs.DescribeContainerInstancesInput{
		ContainerInstances: containerArns,
		Cluster:            &p.Cluster,
	})

	for ; req != nil; req = req.NextPage() {
		if err := wrapAws(ctx, req); err != nil {
			return nil, err
		}

		containerResp := req.Data.(*ecs.DescribeContainerInstancesOutput)
		for i, container := range containerResp.ContainerInstances {
			order[*container.Ec2InstanceId] = order[*container.ContainerInstanceArn]
			instanceIds[i] = container.Ec2InstanceId
		}
	}

	req, _ = client.ec2.DescribeInstancesRequest(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})

	for ; req != nil; req = req.NextPage() {
		if err := wrapAws(ctx, req); err != nil {
			return nil, err
		}

		instancesResp := req.Data.(*ec2.DescribeInstancesOutput)
		for _, r := range instancesResp.Reservations {
			for _, i := range r.Instances {
				if i.InstanceId != nil {
					instances[order[*i.InstanceId]] = i
				}
			}
		}
	}
	return instances, nil
}

func (p *Provider) lookupTaskDefinitions(ctx context.Context, client *awsClient, taskDefArns []*string) ([]*ecs.TaskDefinition, error) {
	taskDefinitions := make([]*ecs.TaskDefinition, len(taskDefArns))
	for i, arn := range taskDefArns {

		req, resp := client.ecs.DescribeTaskDefinitionRequest(&ecs.DescribeTaskDefinitionInput{
			TaskDefinition: arn,
		})

		if err := wrapAws(ctx, req); err != nil {
			return nil, err
		}

		taskDefinitions[i] = resp.TaskDefinition
	}
	return taskDefinitions, nil
}

func (i ecsInstance) label(k string) string {
	if v, found := i.containerDefinition.DockerLabels[k]; found {
		return *v
	}
	return ""
}

func (p *Provider) filterInstance(i ecsInstance) bool {
	if len(i.container.NetworkBindings) == 0 {
		log.Debugf("Filtering ecs instance without port %s (%s)", i.Name, i.ID)
		return false
	}

	if i.machine == nil ||
		i.machine.State == nil ||
		i.machine.State.Name == nil {
		log.Debugf("Filtering ecs instance in an missing ec2 information %s (%s)", i.Name, i.ID)
		return false
	}

	if *i.machine.State.Name != ec2.InstanceStateNameRunning {
		log.Debugf("Filtering ecs instance in an incorrect state %s (%s) (state = %s)", i.Name, i.ID, *i.machine.State.Name)
		return false
	}

	if i.machine.PrivateIpAddress == nil {
		log.Debugf("Filtering ecs instance without an ip address %s (%s)", i.Name, i.ID)
		return false
	}

	label := i.label("traefik.enable")
	enabled := p.ExposedByDefault && label != "false" || label == "true"
	if !enabled {
		log.Debugf("Filtering disabled ecs instance %s (%s) (traefik.enabled = '%s')", i.Name, i.ID, label)
		return false
	}

	return true
}

func (p *Provider) filterFrontends(instances []ecsInstance) []ecsInstance {
	byName := make(map[string]bool)

	return fun.Filter(func(i ecsInstance) bool {
		if _, found := byName[i.Name]; !found {
			byName[i.Name] = true
			return true
		}

		return false
	}, instances).([]ecsInstance)
}

func (p *Provider) getFrontendRule(i ecsInstance) string {
	if label := i.label("traefik.frontend.rule"); label != "" {
		return label
	}
	return "Host:" + strings.ToLower(strings.Replace(i.Name, "_", "-", -1)) + "." + p.Domain
}

// Provider expects no more than 100 parameters be passed to a DescribeTask call; thus, pack
// each string into an array capped at 100 elements
func (p *Provider) chunkedTaskArns(tasks []*string) [][]*string {
	var chunkedTasks [][]*string
	for i := 0; i < len(tasks); i += 100 {
		sliceEnd := -1
		if i+100 < len(tasks) {
			sliceEnd = i + 100
		} else {
			sliceEnd = len(tasks)
		}
		chunkedTasks = append(chunkedTasks, tasks[i:sliceEnd])
	}
	return chunkedTasks
}

func (i ecsInstance) Protocol() string {
	if label := i.label("traefik.protocol"); label != "" {
		return label
	}
	return "http"
}

func (i ecsInstance) Host() string {
	return *i.machine.PrivateIpAddress
}

func (i ecsInstance) Port() string {
	return strconv.FormatInt(*i.container.NetworkBindings[0].HostPort, 10)
}

func (i ecsInstance) Weight() string {
	if label := i.label("traefik.weight"); label != "" {
		return label
	}
	return "0"
}

func (i ecsInstance) PassHostHeader() string {
	if label := i.label("traefik.frontend.passHostHeader"); label != "" {
		return label
	}
	return "true"
}

func (i ecsInstance) Priority() string {
	if label := i.label("traefik.frontend.priority"); label != "" {
		return label
	}
	return "0"
}

func (i ecsInstance) EntryPoints() []string {
	if label := i.label("traefik.frontend.entryPoints"); label != "" {
		return strings.Split(label, ",")
	}
	return []string{}
}
