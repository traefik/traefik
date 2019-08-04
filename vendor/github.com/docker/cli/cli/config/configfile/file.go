package configfile

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/config/credentials"
	"github.com/docker/cli/opts"
	"github.com/docker/docker/api/types"
	"github.com/pkg/errors"
)

const (
	// This constant is only used for really old config files when the
	// URL wasn't saved as part of the config file and it was just
	// assumed to be this value.
	defaultIndexServer = "https://index.docker.io/v1/"
)

// ConfigFile ~/.docker/config.json file info
type ConfigFile struct {
	AuthConfigs          map[string]types.AuthConfig `json:"auths"`
	HTTPHeaders          map[string]string           `json:"HttpHeaders,omitempty"`
	PsFormat             string                      `json:"psFormat,omitempty"`
	ImagesFormat         string                      `json:"imagesFormat,omitempty"`
	NetworksFormat       string                      `json:"networksFormat,omitempty"`
	PluginsFormat        string                      `json:"pluginsFormat,omitempty"`
	VolumesFormat        string                      `json:"volumesFormat,omitempty"`
	StatsFormat          string                      `json:"statsFormat,omitempty"`
	DetachKeys           string                      `json:"detachKeys,omitempty"`
	CredentialsStore     string                      `json:"credsStore,omitempty"`
	CredentialHelpers    map[string]string           `json:"credHelpers,omitempty"`
	Filename             string                      `json:"-"` // Note: for internal use only
	ServiceInspectFormat string                      `json:"serviceInspectFormat,omitempty"`
	ServicesFormat       string                      `json:"servicesFormat,omitempty"`
	TasksFormat          string                      `json:"tasksFormat,omitempty"`
	SecretFormat         string                      `json:"secretFormat,omitempty"`
	ConfigFormat         string                      `json:"configFormat,omitempty"`
	NodesFormat          string                      `json:"nodesFormat,omitempty"`
	PruneFilters         []string                    `json:"pruneFilters,omitempty"`
	Proxies              map[string]ProxyConfig      `json:"proxies,omitempty"`
	Experimental         string                      `json:"experimental,omitempty"`
	StackOrchestrator    string                      `json:"stackOrchestrator,omitempty"`
	Kubernetes           *KubernetesConfig           `json:"kubernetes,omitempty"`
}

// ProxyConfig contains proxy configuration settings
type ProxyConfig struct {
	HTTPProxy  string `json:"httpProxy,omitempty"`
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	NoProxy    string `json:"noProxy,omitempty"`
	FTPProxy   string `json:"ftpProxy,omitempty"`
}

// KubernetesConfig contains Kubernetes orchestrator settings
type KubernetesConfig struct {
	AllNamespaces string `json:"allNamespaces,omitempty"`
}

// New initializes an empty configuration file for the given filename 'fn'
func New(fn string) *ConfigFile {
	return &ConfigFile{
		AuthConfigs: make(map[string]types.AuthConfig),
		HTTPHeaders: make(map[string]string),
		Filename:    fn,
	}
}

// LegacyLoadFromReader reads the non-nested configuration data given and sets up the
// auth config information with given directory and populates the receiver object
func (configFile *ConfigFile) LegacyLoadFromReader(configData io.Reader) error {
	b, err := ioutil.ReadAll(configData)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &configFile.AuthConfigs); err != nil {
		arr := strings.Split(string(b), "\n")
		if len(arr) < 2 {
			return errors.Errorf("The Auth config file is empty")
		}
		authConfig := types.AuthConfig{}
		origAuth := strings.Split(arr[0], " = ")
		if len(origAuth) != 2 {
			return errors.Errorf("Invalid Auth config file")
		}
		authConfig.Username, authConfig.Password, err = decodeAuth(origAuth[1])
		if err != nil {
			return err
		}
		authConfig.ServerAddress = defaultIndexServer
		configFile.AuthConfigs[defaultIndexServer] = authConfig
	} else {
		for k, authConfig := range configFile.AuthConfigs {
			authConfig.Username, authConfig.Password, err = decodeAuth(authConfig.Auth)
			if err != nil {
				return err
			}
			authConfig.Auth = ""
			authConfig.ServerAddress = k
			configFile.AuthConfigs[k] = authConfig
		}
	}
	return nil
}

// LoadFromReader reads the configuration data given and sets up the auth config
// information with given directory and populates the receiver object
func (configFile *ConfigFile) LoadFromReader(configData io.Reader) error {
	if err := json.NewDecoder(configData).Decode(&configFile); err != nil {
		return err
	}
	var err error
	for addr, ac := range configFile.AuthConfigs {
		ac.Username, ac.Password, err = decodeAuth(ac.Auth)
		if err != nil {
			return err
		}
		ac.Auth = ""
		ac.ServerAddress = addr
		configFile.AuthConfigs[addr] = ac
	}
	return checkKubernetesConfiguration(configFile.Kubernetes)
}

// ContainsAuth returns whether there is authentication configured
// in this file or not.
func (configFile *ConfigFile) ContainsAuth() bool {
	return configFile.CredentialsStore != "" ||
		len(configFile.CredentialHelpers) > 0 ||
		len(configFile.AuthConfigs) > 0
}

// GetAuthConfigs returns the mapping of repo to auth configuration
func (configFile *ConfigFile) GetAuthConfigs() map[string]types.AuthConfig {
	return configFile.AuthConfigs
}

// SaveToWriter encodes and writes out all the authorization information to
// the given writer
func (configFile *ConfigFile) SaveToWriter(writer io.Writer) error {
	// Encode sensitive data into a new/temp struct
	tmpAuthConfigs := make(map[string]types.AuthConfig, len(configFile.AuthConfigs))
	for k, authConfig := range configFile.AuthConfigs {
		authCopy := authConfig
		// encode and save the authstring, while blanking out the original fields
		authCopy.Auth = encodeAuth(&authCopy)
		authCopy.Username = ""
		authCopy.Password = ""
		authCopy.ServerAddress = ""
		tmpAuthConfigs[k] = authCopy
	}

	saveAuthConfigs := configFile.AuthConfigs
	configFile.AuthConfigs = tmpAuthConfigs
	defer func() { configFile.AuthConfigs = saveAuthConfigs }()

	data, err := json.MarshalIndent(configFile, "", "\t")
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

// Save encodes and writes out all the authorization information
func (configFile *ConfigFile) Save() error {
	if configFile.Filename == "" {
		return errors.Errorf("Can't save config with empty filename")
	}

	if err := os.MkdirAll(filepath.Dir(configFile.Filename), 0700); err != nil {
		return err
	}
	f, err := os.OpenFile(configFile.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	return configFile.SaveToWriter(f)
}

// ParseProxyConfig computes proxy configuration by retrieving the config for the provided host and
// then checking this against any environment variables provided to the container
func (configFile *ConfigFile) ParseProxyConfig(host string, runOpts []string) map[string]*string {
	var cfgKey string

	if _, ok := configFile.Proxies[host]; !ok {
		cfgKey = "default"
	} else {
		cfgKey = host
	}

	config := configFile.Proxies[cfgKey]
	permitted := map[string]*string{
		"HTTP_PROXY":  &config.HTTPProxy,
		"HTTPS_PROXY": &config.HTTPSProxy,
		"NO_PROXY":    &config.NoProxy,
		"FTP_PROXY":   &config.FTPProxy,
	}
	m := opts.ConvertKVStringsToMapWithNil(runOpts)
	for k := range permitted {
		if *permitted[k] == "" {
			continue
		}
		if _, ok := m[k]; !ok {
			m[k] = permitted[k]
		}
		if _, ok := m[strings.ToLower(k)]; !ok {
			m[strings.ToLower(k)] = permitted[k]
		}
	}
	return m
}

// encodeAuth creates a base64 encoded string to containing authorization information
func encodeAuth(authConfig *types.AuthConfig) string {
	if authConfig.Username == "" && authConfig.Password == "" {
		return ""
	}

	authStr := authConfig.Username + ":" + authConfig.Password
	msg := []byte(authStr)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(msg)))
	base64.StdEncoding.Encode(encoded, msg)
	return string(encoded)
}

// decodeAuth decodes a base64 encoded string and returns username and password
func decodeAuth(authStr string) (string, string, error) {
	if authStr == "" {
		return "", "", nil
	}

	decLen := base64.StdEncoding.DecodedLen(len(authStr))
	decoded := make([]byte, decLen)
	authByte := []byte(authStr)
	n, err := base64.StdEncoding.Decode(decoded, authByte)
	if err != nil {
		return "", "", err
	}
	if n > decLen {
		return "", "", errors.Errorf("Something went wrong decoding auth config")
	}
	arr := strings.SplitN(string(decoded), ":", 2)
	if len(arr) != 2 {
		return "", "", errors.Errorf("Invalid auth configuration file")
	}
	password := strings.Trim(arr[1], "\x00")
	return arr[0], password, nil
}

// GetCredentialsStore returns a new credentials store from the settings in the
// configuration file
func (configFile *ConfigFile) GetCredentialsStore(registryHostname string) credentials.Store {
	if helper := getConfiguredCredentialStore(configFile, registryHostname); helper != "" {
		return newNativeStore(configFile, helper)
	}
	return credentials.NewFileStore(configFile)
}

// var for unit testing.
var newNativeStore = func(configFile *ConfigFile, helperSuffix string) credentials.Store {
	return credentials.NewNativeStore(configFile, helperSuffix)
}

// GetAuthConfig for a repository from the credential store
func (configFile *ConfigFile) GetAuthConfig(registryHostname string) (types.AuthConfig, error) {
	return configFile.GetCredentialsStore(registryHostname).Get(registryHostname)
}

// getConfiguredCredentialStore returns the credential helper configured for the
// given registry, the default credsStore, or the empty string if neither are
// configured.
func getConfiguredCredentialStore(c *ConfigFile, registryHostname string) string {
	if c.CredentialHelpers != nil && registryHostname != "" {
		if helper, exists := c.CredentialHelpers[registryHostname]; exists {
			return helper
		}
	}
	return c.CredentialsStore
}

// GetAllCredentials returns all of the credentials stored in all of the
// configured credential stores.
func (configFile *ConfigFile) GetAllCredentials() (map[string]types.AuthConfig, error) {
	auths := make(map[string]types.AuthConfig)
	addAll := func(from map[string]types.AuthConfig) {
		for reg, ac := range from {
			auths[reg] = ac
		}
	}

	defaultStore := configFile.GetCredentialsStore("")
	newAuths, err := defaultStore.GetAll()
	if err != nil {
		return nil, err
	}
	addAll(newAuths)

	// Auth configs from a registry-specific helper should override those from the default store.
	for registryHostname := range configFile.CredentialHelpers {
		newAuth, err := configFile.GetAuthConfig(registryHostname)
		if err != nil {
			return nil, err
		}
		auths[registryHostname] = newAuth
	}
	return auths, nil
}

// GetFilename returns the file name that this config file is based on.
func (configFile *ConfigFile) GetFilename() string {
	return configFile.Filename
}

func checkKubernetesConfiguration(kubeConfig *KubernetesConfig) error {
	if kubeConfig == nil {
		return nil
	}
	switch kubeConfig.AllNamespaces {
	case "":
	case "enabled":
	case "disabled":
	default:
		return fmt.Errorf("invalid 'kubernetes.allNamespaces' value, should be 'enabled' or 'disabled': %s", kubeConfig.AllNamespaces)
	}
	return nil
}
