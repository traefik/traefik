package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/traefik/paerser/cli"
	"github.com/traefik/paerser/env"
	"github.com/traefik/paerser/file"
	"github.com/traefik/paerser/flag"
	"github.com/traefik/paerser/parser"
)

type RoutingConfigLoader struct{}

func (r RoutingConfigLoader) Load(arguments []string, _ *cli.Command) (bool, error) {
	if err := warnForRoutingConfig(arguments); err != nil {
		log.Debug().Err(err).Msg("Routing configuration warning analysis failed")
	}

	return false, nil
}

// warnForRoutingConfig prints warnings when routing configuration is found in install configuration.
func warnForRoutingConfig(arguments []string) error {
	// This part doesn't handle properly a flag defined like this: --accesslog true
	// where `true` could be considered as a new argument.
	// This is not really an issue with this loader since it will filter the unknown nodes later in this function.
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
		return fmt.Errorf("parsing arguments to labels: %w", err)
	}

	config, err := parseRoutingConfig(argsLabels)
	if err != nil {
		return fmt.Errorf("parsing routing config from args: %w", err)
	}

	// Check for routing configuration elements and warn.
	config.checkRoutingElements(log.With().Str("loader", "args").Logger())

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
		return fmt.Errorf("finding configuration file: %w", err)
	}

	if filePath != "" {
		// We don't rely on the Parser file loader here to avoid issues with unknown fields.
		// Parse file content into a generic map.
		var fileConfig map[string]interface{}
		if err := file.Decode(filePath, &fileConfig); err != nil {
			return fmt.Errorf("decoding configuration file %s: %w", filePath, err)
		}

		// Convert the file config to label format.
		fileLabels := make(map[string]string)
		flattenToLabels(fileConfig, "", fileLabels)

		config, err := parseRoutingConfig(fileLabels)
		if err != nil {
			return fmt.Errorf("parsing routing config from file: %w", err)
		}

		// Check if this is a self-reference scenario by examining the file provider configuration.
		if isSelfReference(config, filePath) {
			return nil
		}

		// Check for routing configuration elements and warn.
		config.checkRoutingElements(log.With().Str("loader", "file").Logger())

		return nil
	}

	// ENV
	vars := env.FindPrefixedEnvVars(os.Environ(), env.DefaultNamePrefix, &configuration{})
	if len(vars) > 0 {
		// We don't rely on the Parser env loader here to avoid issues with unknown fields.
		// Decode environment variables to a generic map.
		var envConfig map[string]interface{}
		if err := env.Decode(vars, env.DefaultNamePrefix, &envConfig); err != nil {
			return fmt.Errorf("decoding environment variables: %w", err)
		}

		// Convert the env config to label format.
		envLabels := make(map[string]string)
		flattenToLabels(envConfig, "", envLabels)

		config, err := parseRoutingConfig(envLabels)
		if err != nil {
			return fmt.Errorf("parsing routin config from environment variables: %w", err)
		}

		// Check for routing configuration elements and warn.
		config.checkRoutingElements(log.With().Str("loader", "env").Logger())
	}

	return nil
}

// parseRoutingConfig parses command-line arguments using the routing configuration struct,
// filtering unknown nodes and checking for routing options.
func parseRoutingConfig(labels map[string]string) (*routingConfiguration, error) {
	// If no config, we can return without error to allow other loaders to proceed.
	if len(labels) == 0 {
		return nil, nil
	}

	// Convert labels to node tree.
	node, err := parser.DecodeToNode(labels, "traefik")
	if err != nil {
		return nil, fmt.Errorf("decoding to node: %w", err)
	}

	// Filter unknown nodes.
	config := &routingConfiguration{}
	filterUnknownNodes(reflect.TypeOf(config), node)

	// If no config remains, we can return without error to allow other loaders to proceed.
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

// isSelfReference checks if providers.file.filename actually points to the same file being loaded.
// This is a legitimate use case where both install and routing config are in the same file.
func isSelfReference(config *routingConfiguration, actualConfigFile string) bool {
	if config == nil || config.Providers == nil || config.Providers.File == nil {
		return false
	}

	// No filename specified, cannot be a self-reference.
	if config.Providers.File.Filename == "" {
		return false
	}

	// Determine the actual config file being loaded.
	if actualConfigFile == "" {
		return false
	}

	// Compare resolved paths to see if they refer to the same file.
	return isSameFile(actualConfigFile, config.Providers.File.Filename)
}

// isSameFile compares two file paths to determine if they refer to the same file.
func isSameFile(path1, path2 string) bool {
	if path1 == "" || path2 == "" {
		return false
	}

	// Clean paths first to handle redundant separators and . and .. elements.
	clean1 := filepath.Clean(path1)
	clean2 := filepath.Clean(path2)

	// Convert cleaned paths to absolute paths for comparison.
	abs1, err1 := filepath.Abs(clean1)
	abs2, err2 := filepath.Abs(clean2)

	if err1 != nil || err2 != nil {
		// Fallback to cleaned path comparison if absolute path resolution fails.
		return clean1 == clean2
	}

	return abs1 == abs2
}

// routingConfiguration holds potential routing configuration elements that might be misplaced in the install config.
type routingConfiguration struct {
	HTTP      map[string]interface{} `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty"`
	TCP       map[string]interface{} `json:"tcp,omitempty" toml:"tcp,omitempty" yaml:"tcp,omitempty"`
	UDP       map[string]interface{} `json:"udp,omitempty" toml:"udp,omitempty" yaml:"udp,omitempty"`
	TLS       map[string]interface{} `json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty"`
	Providers *providersConfig       `json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty"`
}

// providersConfig holds provider configuration for self-reference detection.
type providersConfig struct {
	File *fileProviderConfig `json:"file,omitempty" toml:"file,omitempty" yaml:"file,omitempty"`
}

// fileProviderConfig holds file provider configuration.
type fileProviderConfig struct {
	Filename string `json:"filename,omitempty" toml:"filename,omitempty" yaml:"filename,omitempty"`
}

// checkRoutingElements checks if any routing configuration elements are present and logs warnings.
func (c *routingConfiguration) checkRoutingElements(logger zerolog.Logger) {
	if c == nil {
		return
	}

	if c.HTTP != nil {
		logger.Error().Msg("Found 'http' routing configuration in install configuration." +
			" Please note that this configuration will be ignored." +
			" See https://doc.traefik.io/traefik/getting-started/configuration-overview/")
	}

	if c.TCP != nil {
		logger.Error().Msg("Found 'tcp' routing configuration in install configuration." +
			" Please note that this configuration will be ignored." +
			" See https://doc.traefik.io/traefik/getting-started/configuration-overview/")
	}

	if c.UDP != nil {
		logger.Error().Msg("Found 'udp' routing configuration in install configuration." +
			" Please note that this configuration will be ignored." +
			" See https://doc.traefik.io/traefik/getting-started/configuration-overview/")
	}

	if c.TLS != nil {
		logger.Error().Msg("Found 'tls' routing configuration in install configuration." +
			" Please note that this configuration will be ignored." +
			" See https://doc.traefik.io/traefik/getting-started/configuration-overview/")
	}
}
