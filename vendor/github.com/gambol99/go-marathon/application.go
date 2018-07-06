/*
Copyright 2014 The go-marathon Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package marathon

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"
)

var (
	// ErrNoApplicationContainer is thrown when a container has been specified yet
	ErrNoApplicationContainer = errors.New("you have not specified a docker container yet")
)

// Applications is a collection of applications
type Applications struct {
	Apps []Application `json:"apps"`
}

// IPAddressPerTask is used by IP-per-task functionality https://mesosphere.github.io/marathon/docs/ip-per-task.html
type IPAddressPerTask struct {
	Groups      *[]string          `json:"groups,omitempty"`
	Labels      *map[string]string `json:"labels,omitempty"`
	Discovery   *Discovery         `json:"discovery,omitempty"`
	NetworkName string             `json:"networkName,omitempty"`
}

// Discovery provides info about ports expose by IP-per-task functionality
type Discovery struct {
	Ports *[]Port `json:"ports,omitempty"`
}

// Port provides info about ports used by IP-per-task
type Port struct {
	Number   int    `json:"number,omitempty"`
	Name     string `json:"name,omitempty"`
	Protocol string `json:"protocol,omitempty"`
}

// Application is the definition for an application in marathon
type Application struct {
	ID          string        `json:"id,omitempty"`
	Cmd         *string       `json:"cmd,omitempty"`
	Args        *[]string     `json:"args,omitempty"`
	Constraints *[][]string   `json:"constraints,omitempty"`
	Container   *Container    `json:"container,omitempty"`
	CPUs        float64       `json:"cpus,omitempty"`
	GPUs        *float64      `json:"gpus,omitempty"`
	Disk        *float64      `json:"disk,omitempty"`
	Networks    *[]PodNetwork `json:"networks,omitempty"`

	// Contains non-secret environment variables. Secrets environment variables are part of the Secrets map.
	Env                        *map[string]string  `json:"-"`
	Executor                   *string             `json:"executor,omitempty"`
	HealthChecks               *[]HealthCheck      `json:"healthChecks,omitempty"`
	ReadinessChecks            *[]ReadinessCheck   `json:"readinessChecks,omitempty"`
	Instances                  *int                `json:"instances,omitempty"`
	Mem                        *float64            `json:"mem,omitempty"`
	Tasks                      []*Task             `json:"tasks,omitempty"`
	Ports                      []int               `json:"ports"`
	PortDefinitions            *[]PortDefinition   `json:"portDefinitions,omitempty"`
	RequirePorts               *bool               `json:"requirePorts,omitempty"`
	BackoffSeconds             *float64            `json:"backoffSeconds,omitempty"`
	BackoffFactor              *float64            `json:"backoffFactor,omitempty"`
	MaxLaunchDelaySeconds      *float64            `json:"maxLaunchDelaySeconds,omitempty"`
	TaskKillGracePeriodSeconds *float64            `json:"taskKillGracePeriodSeconds,omitempty"`
	Deployments                []map[string]string `json:"deployments,omitempty"`
	// Available when embedding readiness information through query parameter.
	ReadinessCheckResults *[]ReadinessCheckResult `json:"readinessCheckResults,omitempty"`
	Dependencies          []string                `json:"dependencies"`
	TasksRunning          int                     `json:"tasksRunning,omitempty"`
	TasksStaged           int                     `json:"tasksStaged,omitempty"`
	TasksHealthy          int                     `json:"tasksHealthy,omitempty"`
	TasksUnhealthy        int                     `json:"tasksUnhealthy,omitempty"`
	TaskStats             map[string]TaskStats    `json:"taskStats,omitempty"`
	User                  string                  `json:"user,omitempty"`
	UpgradeStrategy       *UpgradeStrategy        `json:"upgradeStrategy,omitempty"`
	UnreachableStrategy   *UnreachableStrategy    `json:"unreachableStrategy,omitempty"`
	KillSelection         string                  `json:"killSelection,omitempty"`
	Uris                  *[]string               `json:"uris,omitempty"`
	Version               string                  `json:"version,omitempty"`
	VersionInfo           *VersionInfo            `json:"versionInfo,omitempty"`
	Labels                *map[string]string      `json:"labels,omitempty"`
	AcceptedResourceRoles []string                `json:"acceptedResourceRoles,omitempty"`
	LastTaskFailure       *LastTaskFailure        `json:"lastTaskFailure,omitempty"`
	Fetch                 *[]Fetch                `json:"fetch,omitempty"`
	IPAddressPerTask      *IPAddressPerTask       `json:"ipAddress,omitempty"`
	Residency             *Residency              `json:"residency,omitempty"`
	Secrets               *map[string]Secret      `json:"-"`
}

// ApplicationVersions is a collection of application versions for a specific app in marathon
type ApplicationVersions struct {
	Versions []string `json:"versions"`
}

// ApplicationVersion is the application version response from marathon
type ApplicationVersion struct {
	Version string `json:"version"`
}

// VersionInfo is the application versioning details from marathon
type VersionInfo struct {
	LastScalingAt      string `json:"lastScalingAt,omitempty"`
	LastConfigChangeAt string `json:"lastConfigChangeAt,omitempty"`
}

// Fetch will download URI before task starts
type Fetch struct {
	URI        string `json:"uri"`
	Executable bool   `json:"executable"`
	Extract    bool   `json:"extract"`
	Cache      bool   `json:"cache"`
}

// GetAppOpts contains a payload for Application method
//		embed:	Embeds nested resources that match the supplied path.
// 				You can specify this parameter multiple times with different values
type GetAppOpts struct {
	Embed []string `url:"embed,omitempty"`
}

// DeleteAppOpts contains a payload for DeleteApplication method
//		force:		overrides a currently running deployment.
type DeleteAppOpts struct {
	Force bool `url:"force,omitempty"`
}

// TaskStats is a container for Stats
type TaskStats struct {
	Stats Stats `json:"stats"`
}

// Stats is a collection of aggregate statistics about an application's tasks
type Stats struct {
	Counts   map[string]int     `json:"counts"`
	LifeTime map[string]float64 `json:"lifeTime"`
}

// Secret is the environment variable and secret store path associated with a secret.
// The value for EnvVar is populated from the env field, and Source is populated from
// the secrets field of the application json.
type Secret struct {
	EnvVar string
	Source string
}

// SetIPAddressPerTask defines that the application will have a IP address defines by a external agent.
// This configuration is not allowed to be used with Port or PortDefinitions. Thus, the implementation
// clears both.
func (r *Application) SetIPAddressPerTask(ipAddressPerTask IPAddressPerTask) *Application {
	r.Ports = make([]int, 0)
	r.EmptyPortDefinitions()
	r.IPAddressPerTask = &ipAddressPerTask

	return r
}

// NewDockerApplication creates a default docker application
func NewDockerApplication() *Application {
	application := new(Application)
	application.Container = NewDockerContainer()
	return application
}

// Name sets the name / ID of the application i.e. the identifier for this application
func (r *Application) Name(id string) *Application {
	r.ID = validateID(id)
	return r
}

// Command sets the cmd of the application
func (r *Application) Command(cmd string) *Application {
	r.Cmd = &cmd
	return r
}

// CPU set the amount of CPU shares per instance which is assigned to the application
//		cpu:	the CPU shared (check Docker docs) per instance
func (r *Application) CPU(cpu float64) *Application {
	r.CPUs = cpu
	return r
}

// SetGPUs set the amount of GPU per instance which is assigned to the application
//		gpu:	the GPU (check MESOS docs) per instance
func (r *Application) SetGPUs(gpu float64) *Application {
	r.GPUs = &gpu
	return r
}

// EmptyGPUs explicitly empties GPUs -- use this if you need to empty
// gpus of an application that already has gpus set (setting port definitions to nil will
// keep the current value)
func (r *Application) EmptyGPUs() *Application {
	g := 0.0
	r.GPUs = &g
	return r
}

// Storage sets the amount of disk space the application is assigned, which for docker
// application I don't believe is relevant
//		disk:	the disk space in MB
func (r *Application) Storage(disk float64) *Application {
	r.Disk = &disk
	return r
}

// AllTaskRunning checks to see if all the application tasks are running, i.e. the instances is equal
// to the number of running tasks
func (r *Application) AllTaskRunning() bool {
	if r.Instances == nil || *r.Instances == 0 {
		return true
	}
	if r.Tasks == nil {
		return false
	}
	if r.TasksRunning == *r.Instances {
		return true
	}
	return false
}

// DependsOn adds one or more dependencies for this application. Note, if you want to wait for
// an application dependency to actually be UP, i.e. not just deployed, you need a health check
// on the dependant app.
//		names:	the application id(s) this application depends on
func (r *Application) DependsOn(names ...string) *Application {
	if r.Dependencies == nil {
		r.Dependencies = make([]string, 0)
	}
	r.Dependencies = append(r.Dependencies, names...)

	return r
}

// Memory sets he amount of memory the application can consume per instance
//		memory:	the amount of MB to assign
func (r *Application) Memory(memory float64) *Application {
	r.Mem = &memory

	return r
}

// AddPortDefinition adds a port definition. Port definitions are used to define ports that
// should be considered part of a resource. They are necessary when you are using HOST
// networking and no port mappings are specified.
func (r *Application) AddPortDefinition(portDefinition PortDefinition) *Application {
	if r.PortDefinitions == nil {
		r.EmptyPortDefinitions()
	}

	portDefinitions := *r.PortDefinitions
	portDefinitions = append(portDefinitions, portDefinition)
	r.PortDefinitions = &portDefinitions
	return r
}

// EmptyPortDefinitions explicitly empties port definitions -- use this if you need to empty
// port definitions of an application that already has port definitions set (setting port definitions to nil will
// keep the current value)
func (r *Application) EmptyPortDefinitions() *Application {
	r.PortDefinitions = &[]PortDefinition{}

	return r
}

// Count sets the number of instances of the application to run
//		count:	the number of instances to run
func (r *Application) Count(count int) *Application {
	r.Instances = &count

	return r
}

// SetTaskKillGracePeriod sets the number of seconds between escalating from SIGTERM to SIGKILL
// when signalling tasks to terminate. Using this grace period, tasks should perform orderly shut down
// immediately upon receiving SIGTERM.
//		seconds:	the number of seconds
func (r *Application) SetTaskKillGracePeriod(seconds float64) *Application {
	r.TaskKillGracePeriodSeconds = &seconds

	return r
}

// AddArgs adds one or more arguments to the applications
//		arguments:	the argument(s) you are adding
func (r *Application) AddArgs(arguments ...string) *Application {
	if r.Args == nil {
		r.EmptyArgs()
	}

	args := *r.Args
	args = append(args, arguments...)
	r.Args = &args

	return r
}

// EmptyArgs explicitly empties arguments -- use this if you need to empty
// arguments of an application that already has arguments set (setting args to nil will
// keep the current value)
func (r *Application) EmptyArgs() *Application {
	r.Args = &[]string{}

	return r
}

// AddConstraint adds a new constraint
//		constraints:	the constraint definition, one constraint per array element
func (r *Application) AddConstraint(constraints ...string) *Application {
	if r.Constraints == nil {
		r.EmptyConstraints()
	}

	c := *r.Constraints
	c = append(c, constraints)
	r.Constraints = &c

	return r
}

// EmptyConstraints explicitly empties constraints -- use this if you need to empty
// constraints of an application that already has constraints set (setting constraints to nil will
// keep the current value)
func (r *Application) EmptyConstraints() *Application {
	r.Constraints = &[][]string{}

	return r
}

// AddLabel adds a label to the application
//		name:	the name of the label
//		value: value for this label
func (r *Application) AddLabel(name, value string) *Application {
	if r.Labels == nil {
		r.EmptyLabels()
	}
	(*r.Labels)[name] = value

	return r
}

// EmptyLabels explicitly empties the labels -- use this if you need to empty
// the labels of an application that already has labels set (setting labels to nil will
// keep the current value)
func (r *Application) EmptyLabels() *Application {
	r.Labels = &map[string]string{}

	return r
}

// AddEnv adds an environment variable to the application
// name:	the name of the variable
// value:	go figure, the value associated to the above
func (r *Application) AddEnv(name, value string) *Application {
	if r.Env == nil {
		r.EmptyEnvs()
	}
	(*r.Env)[name] = value

	return r
}

// EmptyEnvs explicitly empties the envs -- use this if you need to empty
// the environments of an application that already has environments set (setting env to nil will
// keep the current value)
func (r *Application) EmptyEnvs() *Application {
	r.Env = &map[string]string{}

	return r
}

// AddSecret adds a secret declaration
// envVar: the name of the environment variable
// name:	the name of the secret
// source:	the source ID of the secret
func (r *Application) AddSecret(envVar, name, source string) *Application {
	if r.Secrets == nil {
		r.EmptySecrets()
	}
	(*r.Secrets)[name] = Secret{EnvVar: envVar, Source: source}

	return r
}

// EmptySecrets explicitly empties the secrets -- use this if you need to empty
// the secrets of an application that already has secrets set (setting secrets to nil will
// keep the current value)
func (r *Application) EmptySecrets() *Application {
	r.Secrets = &map[string]Secret{}

	return r
}

// SetExecutor sets the executor
func (r *Application) SetExecutor(executor string) *Application {
	r.Executor = &executor

	return r
}

// AddHealthCheck adds a health check
// 	healthCheck the health check that should be added
func (r *Application) AddHealthCheck(healthCheck HealthCheck) *Application {
	if r.HealthChecks == nil {
		r.EmptyHealthChecks()
	}

	healthChecks := *r.HealthChecks
	healthChecks = append(healthChecks, healthCheck)
	r.HealthChecks = &healthChecks

	return r
}

// EmptyHealthChecks explicitly empties health checks -- use this if you need to empty
// health checks of an application that already has health checks set (setting health checks to nil will
// keep the current value)
func (r *Application) EmptyHealthChecks() *Application {
	r.HealthChecks = &[]HealthCheck{}

	return r
}

// HasHealthChecks is a helper method, used to check if an application has health checks
func (r *Application) HasHealthChecks() bool {
	return r.HealthChecks != nil && len(*r.HealthChecks) > 0
}

// AddReadinessCheck adds a readiness check.
func (r *Application) AddReadinessCheck(readinessCheck ReadinessCheck) *Application {
	if r.ReadinessChecks == nil {
		r.EmptyReadinessChecks()
	}

	readinessChecks := *r.ReadinessChecks
	readinessChecks = append(readinessChecks, readinessCheck)
	r.ReadinessChecks = &readinessChecks

	return r
}

// EmptyReadinessChecks empties the readiness checks.
func (r *Application) EmptyReadinessChecks() *Application {
	r.ReadinessChecks = &[]ReadinessCheck{}

	return r
}

// DeploymentIDs retrieves the application deployments IDs
func (r *Application) DeploymentIDs() []*DeploymentID {
	var deployments []*DeploymentID

	if r.Deployments == nil {
		return deployments
	}

	// step: extract the deployment id from the result
	for _, deploy := range r.Deployments {
		if id, found := deploy["id"]; found {
			deployment := &DeploymentID{
				Version:      r.Version,
				DeploymentID: id,
			}
			deployments = append(deployments, deployment)
		}
	}

	return deployments
}

// CheckHTTP adds a HTTP check to an application
//		port: 		the port the check should be checking
// 		interval:	the interval in seconds the check should be performed
func (r *Application) CheckHTTP(path string, port, interval int) (*Application, error) {
	if r.Container == nil || r.Container.Docker == nil {
		return nil, ErrNoApplicationContainer
	}
	// step: get the port index
	portIndex, err := r.Container.Docker.ServicePortIndex(port)
	if err != nil {
		portIndex, err = r.Container.ServicePortIndex(port)
		if err != nil {
			return nil, err
		}
	}
	health := NewDefaultHealthCheck()
	health.IntervalSeconds = interval
	*health.Path = path
	*health.PortIndex = portIndex
	// step: add to the checks
	r.AddHealthCheck(*health)

	return r, nil
}

// CheckTCP adds a TCP check to an application; note the port mapping must already exist, or an
// error will thrown
//		port: 		the port the check should, err, check
// 		interval:	the interval in seconds the check should be performed
func (r *Application) CheckTCP(port, interval int) (*Application, error) {
	if r.Container == nil || r.Container.Docker == nil {
		return nil, ErrNoApplicationContainer
	}
	// step: get the port index
	portIndex, err := r.Container.Docker.ServicePortIndex(port)
	if err != nil {
		portIndex, err = r.Container.ServicePortIndex(port)
		if err != nil {
			return nil, err
		}
	}
	health := NewDefaultHealthCheck()
	health.Protocol = "TCP"
	health.IntervalSeconds = interval
	*health.PortIndex = portIndex
	// step: add to the checks
	r.AddHealthCheck(*health)

	return r, nil
}

// AddUris adds one or more uris to the applications
//		arguments:	the uri(s) you are adding
func (r *Application) AddUris(newUris ...string) *Application {
	if r.Uris == nil {
		r.EmptyUris()
	}

	uris := *r.Uris
	uris = append(uris, newUris...)
	r.Uris = &uris

	return r
}

// EmptyUris explicitly empties uris -- use this if you need to empty
// uris of an application that already has uris set (setting uris to nil will
// keep the current value)
func (r *Application) EmptyUris() *Application {
	r.Uris = &[]string{}

	return r
}

// AddFetchURIs adds one or more fetch URIs to the application.
//		fetchURIs:	the fetch URI(s) to add.
func (r *Application) AddFetchURIs(fetchURIs ...Fetch) *Application {
	if r.Fetch == nil {
		r.EmptyFetchURIs()
	}

	fetch := *r.Fetch
	fetch = append(fetch, fetchURIs...)
	r.Fetch = &fetch

	return r
}

// EmptyFetchURIs explicitly empties fetch URIs -- use this if you need to empty
// fetch URIs of an application that already has fetch URIs set.
// Setting fetch URIs to nil will keep the current value.
func (r *Application) EmptyFetchURIs() *Application {
	r.Fetch = &[]Fetch{}

	return r
}

// SetUpgradeStrategy sets the upgrade strategy.
func (r *Application) SetUpgradeStrategy(us UpgradeStrategy) *Application {
	r.UpgradeStrategy = &us
	return r
}

// EmptyUpgradeStrategy explicitly empties the upgrade strategy -- use this if
// you need to empty the upgrade strategy of an application that already has
// the upgrade strategy set (setting it to nil will keep the current value).
func (r *Application) EmptyUpgradeStrategy() *Application {
	r.UpgradeStrategy = &UpgradeStrategy{}
	return r
}

// SetUnreachableStrategy sets the unreachable strategy.
func (r *Application) SetUnreachableStrategy(us UnreachableStrategy) *Application {
	r.UnreachableStrategy = &us
	return r
}

// EmptyUnreachableStrategy explicitly empties the unreachable strategy -- use this if
// you need to empty the unreachable strategy of an application that already has
// the unreachable strategy set (setting it to nil will keep the current value).
func (r *Application) EmptyUnreachableStrategy() *Application {
	r.UnreachableStrategy = &UnreachableStrategy{}
	return r
}

// SetResidency sets behavior for resident applications, an application is resident when
// it has local persistent volumes set
func (r *Application) SetResidency(whenLost TaskLostBehaviorType) *Application {
	r.Residency = &Residency{
		TaskLostBehavior: whenLost,
	}
	return r
}

// EmptyResidency explicitly empties the residency -- use this if
// you need to empty the residency of an application that already has
// the residency set (setting it to nil will keep the current value).
func (r *Application) EmptyResidency() *Application {
	r.Residency = &Residency{}
	return r
}

// String returns the json representation of this application
func (r *Application) String() string {
	s, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "error decoding type into json: %s"}`, err)
	}

	return string(s)
}

// Applications retrieves an array of all the applications which are running in marathon
func (r *marathonClient) Applications(v url.Values) (*Applications, error) {
	query := v.Encode()
	if query != "" {
		query = "?" + query
	}

	applications := new(Applications)
	err := r.apiGet(marathonAPIApps+query, nil, applications)
	if err != nil {
		return nil, err
	}

	return applications, nil
}

// ListApplications retrieves an array of the application names currently running in marathon
func (r *marathonClient) ListApplications(v url.Values) ([]string, error) {
	applications, err := r.Applications(v)
	if err != nil {
		return nil, err
	}
	var list []string
	for _, application := range applications.Apps {
		list = append(list, application.ID)
	}

	return list, nil
}

// HasApplicationVersion checks to see if the application version exists in Marathon
// 		name: 		the id used to identify the application
//		version: 	the version (normally a timestamp) your looking for
func (r *marathonClient) HasApplicationVersion(name, version string) (bool, error) {
	id := trimRootPath(name)
	versions, err := r.ApplicationVersions(id)
	if err != nil {
		return false, err
	}

	return contains(versions.Versions, version), nil
}

// ApplicationVersions is a list of versions which has been deployed with marathon for a specific application
//		name:		the id used to identify the application
func (r *marathonClient) ApplicationVersions(name string) (*ApplicationVersions, error) {
	path := fmt.Sprintf("%s/versions", buildPath(name))
	versions := new(ApplicationVersions)
	if err := r.apiGet(path, nil, versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// SetApplicationVersion changes the version of the application
// 		name: 		the id used to identify the application
//		version: 	the version (normally a timestamp) you wish to change to
func (r *marathonClient) SetApplicationVersion(name string, version *ApplicationVersion) (*DeploymentID, error) {
	path := buildPath(name)
	deploymentID := new(DeploymentID)
	if err := r.apiPut(path, version, deploymentID); err != nil {
		return nil, err
	}

	return deploymentID, nil
}

// Application retrieves the application configuration from marathon
// 		name: 		the id used to identify the application
func (r *marathonClient) Application(name string) (*Application, error) {
	var wrapper struct {
		Application *Application `json:"app"`
	}

	if err := r.apiGet(buildPath(name), nil, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.Application, nil
}

// ApplicationBy retrieves the application configuration from marathon
// 		name: 		the id used to identify the application
//		opts:		GetAppOpts request payload
func (r *marathonClient) ApplicationBy(name string, opts *GetAppOpts) (*Application, error) {
	path, err := addOptions(buildPath(name), opts)
	if err != nil {
		return nil, err
	}
	var wrapper struct {
		Application *Application `json:"app"`
	}

	if err := r.apiGet(path, nil, &wrapper); err != nil {
		return nil, err
	}

	return wrapper.Application, nil
}

// ApplicationByVersion retrieves the application configuration from marathon
// 		name: 		the id used to identify the application
// 		version:  the version of the configuration you would like to receive
func (r *marathonClient) ApplicationByVersion(name, version string) (*Application, error) {
	app := new(Application)

	path := fmt.Sprintf("%s/versions/%s", buildPath(name), version)
	if err := r.apiGet(path, nil, app); err != nil {
		return nil, err
	}

	return app, nil
}

// ApplicationOK validates that the application, or more appropriately it's tasks have passed all the health checks.
// If no health checks exist, we simply return true
// 		name: 		the id used to identify the application
func (r *marathonClient) ApplicationOK(name string) (bool, error) {
	// step: get the application
	application, err := r.Application(name)
	if err != nil {
		return false, err
	}

	// step: check if all the tasks are running?
	if !application.AllTaskRunning() {
		return false, nil
	}

	// step: if the application has not health checks, just return true
	if application.HealthChecks == nil || len(*application.HealthChecks) == 0 {
		return true, nil
	}

	// step: iterate the application checks and look for false
	for _, task := range application.Tasks {
		// Health check results may not be available immediately. Assume
		// non-healthiness if they are missing for any task.
		if task.HealthCheckResults == nil {
			return false, nil
		}

		for _, check := range task.HealthCheckResults {
			//When a task is flapping in Marathon, this is sometimes nil
			if check == nil || !check.Alive {
				return false, nil
			}
		}
	}

	return true, nil
}

// ApplicationDeployments retrieves an array of Deployment IDs for an application
//       name:       the id used to identify the application
func (r *marathonClient) ApplicationDeployments(name string) ([]*DeploymentID, error) {
	application, err := r.Application(name)
	if err != nil {
		return nil, err
	}

	return application.DeploymentIDs(), nil
}

// CreateApplication creates a new application in Marathon
// 		application:		the structure holding the application configuration
func (r *marathonClient) CreateApplication(application *Application) (*Application, error) {
	result := new(Application)
	if err := r.apiPost(marathonAPIApps, application, result); err != nil {
		return nil, err
	}

	return result, nil
}

// WaitOnApplication waits for an application to be deployed
//		name:		the id of the application
//		timeout:	a duration of time to wait for an application to deploy
func (r *marathonClient) WaitOnApplication(name string, timeout time.Duration) error {
	return r.wait(name, timeout, r.appExistAndRunning)
}

func (r *marathonClient) appExistAndRunning(name string) bool {
	app, err := r.Application(name)
	if apiErr, ok := err.(*APIError); ok && apiErr.ErrCode == ErrCodeNotFound {
		return false
	}
	if err == nil && app.AllTaskRunning() {
		return true
	}
	return false
}

// DeleteApplication deletes an application from marathon
// 		name: 		the id used to identify the application
//		force:		used to force the delete operation in case of blocked deployment
func (r *marathonClient) DeleteApplication(name string, force bool) (*DeploymentID, error) {
	path := buildPathWithForceParam(name, force)
	// step: check of the application already exists
	deployID := new(DeploymentID)
	if err := r.apiDelete(path, nil, deployID); err != nil {
		return nil, err
	}

	return deployID, nil
}

// RestartApplication performs a rolling restart of marathon application
// 		name: 		the id used to identify the application
func (r *marathonClient) RestartApplication(name string, force bool) (*DeploymentID, error) {
	deployment := new(DeploymentID)
	var options struct{}
	path := buildPathWithForceParam(fmt.Sprintf("%s/restart", name), force)
	if err := r.apiPost(path, &options, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

// ScaleApplicationInstances changes the number of instance an application is running
// 		name: 		the id used to identify the application
// 		instances:	the number of instances you wish to change to
//    force: used to force the scale operation in case of blocked deployment
func (r *marathonClient) ScaleApplicationInstances(name string, instances int, force bool) (*DeploymentID, error) {
	changes := new(Application)
	changes.ID = validateID(name)
	changes.Instances = &instances
	path := buildPathWithForceParam(name, force)
	deployID := new(DeploymentID)
	if err := r.apiPut(path, changes, deployID); err != nil {
		return nil, err
	}

	return deployID, nil
}

// UpdateApplication updates an application in Marathon
// 		application:		the structure holding the application configuration
func (r *marathonClient) UpdateApplication(application *Application, force bool) (*DeploymentID, error) {
	result := new(DeploymentID)
	path := buildPathWithForceParam(application.ID, force)
	if err := r.apiPut(path, application, result); err != nil {
		return nil, err
	}
	return result, nil
}

func buildPathWithForceParam(rootPath string, force bool) string {
	path := buildPath(rootPath)
	if force {
		path += "?force=true"
	}
	return path
}

func buildPath(path string) string {
	return fmt.Sprintf("%s/%s", marathonAPIApps, trimRootPath(path))
}

// EmptyLabels explicitly empties labels -- use this if you need to empty
// labels of an application that already has IP per task with labels defined
func (i *IPAddressPerTask) EmptyLabels() *IPAddressPerTask {
	i.Labels = &map[string]string{}
	return i
}

// AddLabel adds a label to an IPAddressPerTask
//    name: The label name
//   value: The label value
func (i *IPAddressPerTask) AddLabel(name, value string) *IPAddressPerTask {
	if i.Labels == nil {
		i.EmptyLabels()
	}
	(*i.Labels)[name] = value
	return i
}

// EmptyGroups explicitly empties groups -- use this if you need to empty
// groups of an application that already has IP per task with groups defined
func (i *IPAddressPerTask) EmptyGroups() *IPAddressPerTask {
	i.Groups = &[]string{}
	return i
}

// AddGroup adds a group to an IPAddressPerTask
//  group: The group name
func (i *IPAddressPerTask) AddGroup(group string) *IPAddressPerTask {
	if i.Groups == nil {
		i.EmptyGroups()
	}

	groups := *i.Groups
	groups = append(groups, group)
	i.Groups = &groups

	return i
}

// SetDiscovery define the discovery to an IPAddressPerTask
//  discovery: The discovery struct
func (i *IPAddressPerTask) SetDiscovery(discovery Discovery) *IPAddressPerTask {
	i.Discovery = &discovery
	return i
}

// EmptyPorts explicitly empties discovey port -- use this if you need to empty
// discovey port of an application that already has IP per task with discovey ports
// defined
func (d *Discovery) EmptyPorts() *Discovery {
	d.Ports = &[]Port{}
	return d
}

// AddPort adds a port to the discovery info of a IP per task applicable
//   port: The discovery port
func (d *Discovery) AddPort(port Port) *Discovery {
	if d.Ports == nil {
		d.EmptyPorts()
	}
	ports := *d.Ports
	ports = append(ports, port)
	d.Ports = &ports
	return d
}

// EmptyNetworks explicitly empties networks
func (r *Application) EmptyNetworks() *Application {
	r.Networks = &[]PodNetwork{}
	return r
}

// SetNetwork sets the networking mode
func (r *Application) SetNetwork(name string, mode PodNetworkMode) *Application {
	if r.Networks == nil {
		r.EmptyNetworks()
	}

	network := PodNetwork{Name: name, Mode: mode}
	networks := *r.Networks
	networks = append(networks, network)
	r.Networks = &networks
	return r
}
