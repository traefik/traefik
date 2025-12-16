package cli

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/env"
	"github.com/traefik/paerser/file"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/paerser/parser"
)

type DeprecationLoader struct{}

func (d DeprecationLoader) Load(args []string, _ *cli.Command) (bool, error) {
	hasIncompatibleOptions, err := logDeprecations(args)
	if err != nil {
		log.Error().Err(err).Msg("Deprecated install configuration options analysis failed")
		return false, nil
	}

	if hasIncompatibleOptions {
		return true, errors.New("incompatible deprecated install configuration option found")
	}
	return false, nil
}

// logDeprecations prints deprecation hints and returns whether incompatible deprecated options need to be removed.
func logDeprecations(arguments []string) (bool, error) {
	// This part doesn't handle properly a flag defined like this: --accesslog true
	// where `true` could be considered as a new argument.
	// This is not really an issue with the deprecation loader since it will filter the unknown nodes later in this function.
	var args []string
	for _, arg := range arguments {
		if !strings.Contains(arg, "=") {
			args = append(args, arg+"=true")
			continue
		}
		args = append(args, arg)
	}

	// ARGS
	// Parse arguments to labels.
	argsLabels, err := flag.Parse(args, nil)
	if err != nil {
		return false, fmt.Errorf("parsing arguments to labels: %w", err)
	}

	config, err := parseDeprecatedConfig(argsLabels)
	if err != nil {
		return false, fmt.Errorf("parsing deprecated config from args: %w", err)
	}

	if config.deprecationNotice(log.With().Str("loader", "args").Logger()) {
		return true, nil
	}

	// FILE
	// Find the config file using the same logic as the normal file loader.
	finder := cli.Finder{
		BasePaths:  []string{"/etc/traefik/traefik", "$XDG_CONFIG_HOME/traefik", "$HOME/.config/traefik", "./traefik"},
		Extensions: []string{"toml", "yaml", "yml"},
	}

	configFile, ok := argsLabels["traefik.configfile"]
	if !ok {
		configFile = argsLabels["traefik.configFile"]
	}

	filePath, err := finder.Find(configFile)
	if err != nil {
		return false, fmt.Errorf("finding configuration file: %w", err)
	}

	if filePath != "" {
		// We don't rely on the Parser file loader here to avoid issues with unknown fields.
		// Parse file content into a generic map.
		var fileConfig map[string]interface{}
		if err := file.Decode(filePath, &fileConfig); err != nil {
			return false, fmt.Errorf("decoding configuration file %s: %w", filePath, err)
		}

		// Convert the file config to label format.
		fileLabels := make(map[string]string)
		flattenToLabels(fileConfig, "", fileLabels)

		config, err := parseDeprecatedConfig(fileLabels)
		if err != nil {
			return false, fmt.Errorf("parsing deprecated config from file: %w", err)
		}

		if config.deprecationNotice(log.With().Str("loader", "file").Logger()) {
			return true, nil
		}
	}

	// ENV
	vars := env.FindPrefixedEnvVars(os.Environ(), env.DefaultNamePrefix, &configuration{})
	if len(vars) > 0 {
		// We don't rely on the Parser env loader here to avoid issues with unknown fields.
		// Decode environment variables to a generic map.
		var envConfig map[string]interface{}
		if err := env.Decode(vars, env.DefaultNamePrefix, &envConfig); err != nil {
			return false, fmt.Errorf("decoding environment variables: %w", err)
		}

		// Convert the env config to label format.
		envLabels := make(map[string]string)
		flattenToLabels(envConfig, "", envLabels)

		config, err := parseDeprecatedConfig(envLabels)
		if err != nil {
			return false, fmt.Errorf("parsing deprecated config from environment variables: %w", err)
		}

		if config.deprecationNotice(log.With().Str("loader", "env").Logger()) {
			return true, nil
		}
	}

	return false, nil
}

// flattenToLabels recursively flattens a nested map into label key-value pairs.
// Example: {"experimental": {"http3": true}} -> {"traefik.experimental.http3": "true"}.
func flattenToLabels(config interface{}, currKey string, labels map[string]string) {
	switch v := config.(type) {
	case map[string]interface{}:
		for key, value := range v {
			newKey := key
			if currKey != "" {
				newKey = currKey + "." + key
			}
			flattenToLabels(value, newKey, labels)
		}
	case []interface{}:
		for i, item := range v {
			newKey := currKey + "[" + strconv.Itoa(i) + "]"
			flattenToLabels(item, newKey, labels)
		}
	default:
		// Convert value to string and create label with traefik prefix.
		labels["traefik."+currKey] = fmt.Sprintf("%v", v)
	}
}

// parseDeprecatedConfig parses command-line arguments using the deprecation configuration struct,
// filtering unknown nodes and checking for deprecated options.
// Returns true if incompatible deprecated options are found.
func parseDeprecatedConfig(labels map[string]string) (*configuration, error) {
	// If no config, we can return without error to allow other loaders to proceed.
	if len(labels) == 0 {
		return nil, nil
	}

	// Convert labels to node tree.
	node, err := parser.DecodeToNode(labels, "traefik")
	if err != nil {
		return nil, fmt.Errorf("decoding to node: %w", err)
	}

	// Filter unknown nodes and check for deprecated options.
	config := &configuration{}
	filterUnknownNodes(reflect.TypeOf(config), node)

	// If no config remains we can return without error, to allow other loaders to proceed.
	if node == nil || len(node.Children) == 0 {
		return nil, nil
	}

	// Telling parser to look for the label struct tag to allow empty values.
	err = parser.AddMetadata(config, node, parser.MetadataOpts{TagName: "label"})
	if err != nil {
		return nil, fmt.Errorf("adding metadata to node: %w", err)
	}

	err = parser.Fill(config, node, parser.FillerOpts{})
	if err != nil {
		return nil, fmt.Errorf("filling configuration: %w", err)
	}

	return config, nil
}

// filterUnknownNodes removes from the node tree all nodes that do not correspond to any field in the given type.
func filterUnknownNodes(fType reflect.Type, node *parser.Node) bool {
	var children []*parser.Node
	for _, child := range node.Children {
		if hasKnownNodes(fType, child) {
			children = append(children, child)
		}
	}

	node.Children = children
	return len(node.Children) > 0
}

// hasKnownNodes checks whether the given node corresponds to a known field in the given type.
func hasKnownNodes(rootType reflect.Type, node *parser.Node) bool {
	rType := rootType
	if rootType.Kind() == reflect.Pointer {
		rType = rootType.Elem()
	}

	// unstructured type fitting anything, considering the current node as known.
	if rType.Kind() == reflect.Map && rType.Elem().Kind() == reflect.Interface {
		return true
	}

	// unstructured type fitting anything, considering the current node as known.
	if rType.Kind() == reflect.Interface {
		return true
	}

	// find matching field in struct type.
	field, b := findTypedField(rType, node)
	if !b {
		return b
	}

	if len(node.Children) > 0 {
		return filterUnknownNodes(field.Type, node)
	}

	return true
}

func findTypedField(rType reflect.Type, node *parser.Node) (reflect.StructField, bool) {
	// avoid panicking.
	if rType.Kind() != reflect.Struct {
		return reflect.StructField{}, false
	}

	for i := range rType.NumField() {
		cField := rType.Field(i)

		// ignore unexported fields.
		if cField.PkgPath == "" {
			if strings.EqualFold(cField.Name, node.Name) {
				node.FieldName = cField.Name
				return cField, true
			}
		}
	}

	return reflect.StructField{}, false
}

// configuration holds the install configuration removed/deprecated options.
type configuration struct {
	Core         *core          `json:"core,omitempty" toml:"core,omitempty" yaml:"core,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Experimental *experimental  `json:"experimental,omitempty" toml:"experimental,omitempty" yaml:"experimental,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Pilot        map[string]any `json:"pilot,omitempty" toml:"pilot,omitempty" yaml:"pilot,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Providers    *providers     `json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Tracing      *tracing       `json:"tracing,omitempty" toml:"tracing,omitempty" yaml:"tracing,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (c *configuration) deprecationNotice(logger zerolog.Logger) bool {
	if c == nil {
		return false
	}

	var incompatible bool
	if c.Pilot != nil {
		incompatible = true
		logger.Error().Msg("Pilot configuration has been removed in v3, please remove all Pilot-related install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#pilot")
	}

	incompatibleCore := c.Core.deprecationNotice(logger)
	incompatibleExperimental := c.Experimental.deprecationNotice(logger)
	incompatibleProviders := c.Providers.deprecationNotice(logger)
	incompatibleTracing := c.Tracing.deprecationNotice(logger)
	return incompatible || incompatibleCore || incompatibleExperimental || incompatibleProviders || incompatibleTracing
}

type core struct {
	DefaultRuleSyntax string `json:"defaultRuleSyntax,omitempty" toml:"defaultRuleSyntax,omitempty" yaml:"defaultRuleSyntax,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (c *core) deprecationNotice(logger zerolog.Logger) bool {
	if c != nil && c.DefaultRuleSyntax != "" {
		logger.Error().Msg("`Core.DefaultRuleSyntax` option has been deprecated in v3.4, and will be removed in the next major version." +
			" Please consider migrating all router rules to v3 syntax." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v3/#rule-syntax")
	}

	return false
}

type providers struct {
	Docker            *docker        `json:"docker,omitempty" toml:"docker,omitempty" yaml:"docker,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Swarm             *swarm         `json:"swarm,omitempty" toml:"swarm,omitempty" yaml:"swarm,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Consul            *consul        `json:"consul,omitempty" toml:"consul,omitempty" yaml:"consul,omitempty" label:"allowEmpty" file:"allowEmpty"`
	ConsulCatalog     *consulCatalog `json:"consulCatalog,omitempty" toml:"consulCatalog,omitempty" yaml:"consulCatalog,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Nomad             *nomad         `json:"nomad,omitempty" toml:"nomad,omitempty" yaml:"nomad,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Marathon          map[string]any `json:"marathon,omitempty" toml:"marathon,omitempty" yaml:"marathon,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Rancher           map[string]any `json:"rancher,omitempty" toml:"rancher,omitempty" yaml:"rancher,omitempty" label:"allowEmpty" file:"allowEmpty"`
	ETCD              *etcd          `json:"etcd,omitempty" toml:"etcd,omitempty" yaml:"etcd,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Redis             *redis         `json:"redis,omitempty" toml:"redis,omitempty" yaml:"redis,omitempty" label:"allowEmpty" file:"allowEmpty"`
	HTTP              *http          `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" label:"allowEmpty" file:"allowEmpty"`
	KubernetesIngress *ingress       `json:"kubernetesIngress,omitempty" toml:"kubernetesIngress,omitempty" yaml:"kubernetesIngress,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (p *providers) deprecationNotice(logger zerolog.Logger) bool {
	if p == nil {
		return false
	}

	var incompatible bool

	if p.Marathon != nil {
		incompatible = true
		logger.Error().Msg("Marathon provider has been removed in v3, please remove all Marathon-related install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#marathon-provider")
	}

	if p.Rancher != nil {
		incompatible = true
		logger.Error().Msg("Rancher provider has been removed in v3, please remove all Rancher-related install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#rancher-v1-provider")
	}

	dockerIncompatible := p.Docker.deprecationNotice(logger)
	consulIncompatible := p.Consul.deprecationNotice(logger)
	consulCatalogIncompatible := p.ConsulCatalog.deprecationNotice(logger)
	nomadIncompatible := p.Nomad.deprecationNotice(logger)
	swarmIncompatible := p.Swarm.deprecationNotice(logger)
	etcdIncompatible := p.ETCD.deprecationNotice(logger)
	redisIncompatible := p.Redis.deprecationNotice(logger)
	httpIncompatible := p.HTTP.deprecationNotice(logger)
	p.KubernetesIngress.deprecationNotice(logger)
	return incompatible ||
		dockerIncompatible ||
		consulIncompatible ||
		consulCatalogIncompatible ||
		nomadIncompatible ||
		swarmIncompatible ||
		etcdIncompatible ||
		redisIncompatible ||
		httpIncompatible
}

type tls struct {
	CAOptional *bool `json:"caOptional,omitempty" toml:"caOptional,omitempty" yaml:"caOptional,omitempty"`
}

type docker struct {
	SwarmMode *bool `json:"swarmMode,omitempty" toml:"swarmMode,omitempty" yaml:"swarmMode,omitempty"`
	TLS       *tls  `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (d *docker) deprecationNotice(logger zerolog.Logger) bool {
	if d == nil {
		return false
	}

	var incompatible bool

	if d.SwarmMode != nil {
		incompatible = true
		logger.Error().Msg("Docker provider `swarmMode` option has been removed in v3, please use the Swarm Provider instead." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#docker-docker-swarm")
	}

	if d.TLS != nil && d.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Docker provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tlscaoptional")
	}

	return incompatible
}

type swarm struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (s *swarm) deprecationNotice(logger zerolog.Logger) bool {
	if s == nil {
		return false
	}

	var incompatible bool

	if s.TLS != nil && s.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Swarm provider `tls.CAOptional` option does not exist, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start.")
	}

	return incompatible
}

type etcd struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (e *etcd) deprecationNotice(logger zerolog.Logger) bool {
	if e == nil {
		return false
	}

	var incompatible bool

	if e.TLS != nil && e.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("ETCD provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tlscaoptional_3")
	}

	return incompatible
}

type redis struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (r *redis) deprecationNotice(logger zerolog.Logger) bool {
	if r == nil {
		return false
	}

	var incompatible bool

	if r.TLS != nil && r.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Redis provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tlscaoptional_4")
	}

	return incompatible
}

type consul struct {
	Namespace *string `json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	TLS       *tls    `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (c *consul) deprecationNotice(logger zerolog.Logger) bool {
	if c == nil {
		return false
	}

	var incompatible bool

	if c.Namespace != nil {
		incompatible = true
		logger.Error().Msg("Consul provider `namespace` option has been removed, please use the `namespaces` option instead." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#consul-provider")
	}

	if c.TLS != nil && c.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Consul provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tlscaoptional_1")
	}

	return incompatible
}

type consulCatalog struct {
	Namespace *string         `json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	Endpoint  *endpointConfig `json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

type endpointConfig struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
}

func (c *consulCatalog) deprecationNotice(logger zerolog.Logger) bool {
	if c == nil {
		return false
	}

	var incompatible bool

	if c.Namespace != nil {
		incompatible = true
		logger.Error().Msg("ConsulCatalog provider `namespace` option has been removed, please use the `namespaces` option instead." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#consulcatalog-provider")
	}

	if c.Endpoint != nil && c.Endpoint.TLS != nil && c.Endpoint.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("ConsulCatalog provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#endpointtlscaoptional")
	}

	return incompatible
}

type nomad struct {
	Namespace *string         `json:"namespace,omitempty" toml:"namespace,omitempty" yaml:"namespace,omitempty"`
	Endpoint  *endpointConfig `json:"endpoint,omitempty" toml:"endpoint,omitempty" yaml:"endpoint,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (n *nomad) deprecationNotice(logger zerolog.Logger) bool {
	if n == nil {
		return false
	}

	var incompatible bool

	if n.Namespace != nil {
		incompatible = true
		logger.Error().Msg("Nomad provider `namespace` option has been removed, please use the `namespaces` option instead." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#nomad-provider")
	}

	if n.Endpoint != nil && n.Endpoint.TLS != nil && n.Endpoint.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("Nomad provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#endpointtlscaoptional_1")
	}

	return incompatible
}

type http struct {
	TLS *tls `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (h *http) deprecationNotice(logger zerolog.Logger) bool {
	if h == nil {
		return false
	}

	var incompatible bool

	if h.TLS != nil && h.TLS.CAOptional != nil {
		incompatible = true
		logger.Error().Msg("HTTP provider `tls.CAOptional` option has been removed in v3, as TLS client authentication is a server side option (see https://github.com/golang/go/blob/740a490f71d026bb7d2d13cb8fa2d6d6e0572b70/src/crypto/tls/common.go#L634)." +
			" Please remove all occurrences from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tlscaoptional_2")
	}

	return incompatible
}

type ingress struct {
	DisableIngressClassLookup *bool `json:"disableIngressClassLookup,omitempty" toml:"disableIngressClassLookup,omitempty" yaml:"disableIngressClassLookup,omitempty"`
}

func (i *ingress) deprecationNotice(logger zerolog.Logger) {
	if i == nil {
		return
	}

	if i.DisableIngressClassLookup != nil {
		logger.Error().Msg("Kubernetes Ingress provider `disableIngressClassLookup` option has been deprecated in v3.1, and will be removed in the next major version." +
			" Please use the `disableClusterScopeResources` option instead." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v3/#ingressclasslookup")
	}
}

type experimental struct {
	HTTP3                  *bool `json:"http3,omitempty" toml:"http3,omitempty" yaml:"http3,omitempty"`
	KubernetesGateway      *bool `json:"kubernetesGateway,omitempty" toml:"kubernetesGateway,omitempty" yaml:"kubernetesGateway,omitempty"`
	KubernetesIngressNGINX *bool `json:"kubernetesIngressNGINX,omitempty" toml:"kubernetesIngressNGINX,omitempty" yaml:"kubernetesIngressNGINX,omitempty"`
}

func (e *experimental) deprecationNotice(logger zerolog.Logger) bool {
	if e == nil {
		return false
	}

	if e.HTTP3 != nil {
		logger.Error().Msg("HTTP3 is not an experimental feature in v3 and the associated enablement has been removed." +
			" Please remove its usage from the install configuration for Traefik to start." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3-details/#http3")

		return true
	}

	if e.KubernetesGateway != nil {
		logger.Error().Msg("KubernetesGateway provider is not an experimental feature starting with v3.1." +
			" Please remove its usage from the install configuration." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v3/#gateway-api-kubernetesgateway-provider")
	}

	if e.KubernetesIngressNGINX != nil {
		logger.Error().Msg("KubernetesIngressNGINX provider is not an experimental feature starting with v3.6.2." +
			" Please remove its usage from the install configuration." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v3/#ingress-nginx-provider")
	}

	return false
}

//

type tracing struct {
	SpanNameLimit    *int              `json:"spanNameLimit,omitempty" toml:"spanNameLimit,omitempty" yaml:"spanNameLimit,omitempty"`
	GlobalAttributes map[string]string `json:"globalAttributes,omitempty" toml:"globalAttributes,omitempty" yaml:"globalAttributes,omitempty" export:"true"`
	Jaeger           map[string]any    `json:"jaeger,omitempty" toml:"jaeger,omitempty" yaml:"jaeger,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Zipkin           map[string]any    `json:"zipkin,omitempty" toml:"zipkin,omitempty" yaml:"zipkin,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Datadog          map[string]any    `json:"datadog,omitempty" toml:"datadog,omitempty" yaml:"datadog,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Instana          map[string]any    `json:"instana,omitempty" toml:"instana,omitempty" yaml:"instana,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Haystack         map[string]any    `json:"haystack,omitempty" toml:"haystack,omitempty" yaml:"haystack,omitempty" label:"allowEmpty" file:"allowEmpty"`
	Elastic          map[string]any    `json:"elastic,omitempty" toml:"elastic,omitempty" yaml:"elastic,omitempty" label:"allowEmpty" file:"allowEmpty"`
}

func (t *tracing) deprecationNotice(logger zerolog.Logger) bool {
	if t == nil {
		return false
	}
	var incompatible bool
	if t.SpanNameLimit != nil {
		incompatible = true
		logger.Error().Msg("SpanNameLimit option for Tracing has been removed in v3, as Span names are now of a fixed length." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tracing")
	}

	if t.GlobalAttributes != nil {
		log.Warn().Msgf("tracing.globalAttributes option is now deprecated, please use tracing.resourceAttributes instead.")

		logger.Error().Msg("`tracing.globalAttributes` option has been deprecated in v3.3, and will be removed in the next major version." +
			" Please use the `tracing.resourceAttributes` option instead." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v3/#tracing-global-attributes")
	}

	if t.Jaeger != nil {
		incompatible = true
		logger.Error().Msg("Jaeger Tracing backend has been removed in v3, please remove all Jaeger-related Tracing install configuration for Traefik to start." +
			" In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tracing")
	}

	if t.Zipkin != nil {
		incompatible = true
		logger.Error().Msg("Zipkin Tracing backend has been removed in v3, please remove all Zipkin-related Tracing install configuration for Traefik to start." +
			" In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tracing")
	}

	if t.Datadog != nil {
		incompatible = true
		logger.Error().Msg("Datadog Tracing backend has been removed in v3, please remove all Datadog-related Tracing install configuration for Traefik to start." +
			" In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tracing")
	}

	if t.Instana != nil {
		incompatible = true
		logger.Error().Msg("Instana Tracing backend has been removed in v3, please remove all Instana-related Tracing install configuration for Traefik to start." +
			" In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tracing")
	}

	if t.Haystack != nil {
		incompatible = true
		logger.Error().Msg("Haystack Tracing backend has been removed in v3, please remove all Haystack-related Tracing install configuration for Traefik to start." +
			" In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tracing")
	}

	if t.Elastic != nil {
		incompatible = true
		logger.Error().Msg("Elastic Tracing backend has been removed in v3, please remove all Elastic-related Tracing install configuration for Traefik to start." +
			" In v3, Open Telemetry replaces specific tracing backend implementations, and an collector/exporter can be used to export metrics in a vendor specific format." +
			" For more information please read the migration guide: https://doc.traefik.io/traefik/v3.6/migration/v2-to-v3/#tracing")
	}

	return incompatible
}
