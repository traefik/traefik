package snapd

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/BurntSushi/ty/fun"
	"github.com/cenk/backoff"
	"github.com/containous/traefik/job"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/safe"
	"github.com/containous/traefik/types"
	"github.com/snapcore/snapd/client"
)

var _ provider.Provider = (*Provider)(nil)

// Provider holds configurations of the provider.
type Provider struct {
	provider.BaseProvider `mapstructure:",squash" export:"true"`
	Endpoint              string `description:"snapd server endpoint. Can be a tcp or a unix socket endpoint"`
	Domain                string `description:"Default domain used"`
	ExposedByDefault      bool   `description:"Expose snaps by default" export:"false"`
}

// Init the provider
func (p *Provider) Init(constraints types.Constraints) error {
	return p.BaseProvider.Init(constraints)
}

// snapData holds the needed data for the Provider p
type snapData struct {
	SnapName          string
	Properties        map[string]string
	SegmentProperties label.SegmentPropertyValues
	SegmentName       string
}

func (p *Provider) createClient() (*client.Client, error) {
	var config = &client.Config{} // TODO fill config
	return client.New(config), nil
}

// Provide allows the snapd provider to provide configurations to traefik
// using the given configuration channel.
func (p *Provider) Provide(configurationChan chan<- types.ConfigMessage, pool *safe.Pool) error {
	// TODO register this routine in pool, and watch for stop channel
	safe.Go(func() {
		operation := func() error {
			var err error

			snapdClient, err := p.createClient()
			if err != nil {
				log.Errorf("Failed to create a client for snapd, error: %s", err)
				return err
			}
			serverVersion, err := snapdClient.ServerVersion()
			if err != nil {
				log.Errorf("Failed to retrieve information of the snapd client and server host: %s", err)
				return err
			}
			log.Debugf("Provider connection established with snapd %s (Series %s)", serverVersion.Version, serverVersion.Series)
			var snapList []snapData
			snapList, err = listEnabledSnaps(snapdClient)
			if err != nil {
				log.Errorf("Failed to list snaps, error %s", err)
				return err
			}

			configuration := p.buildConfiguration(snapList)
			configurationChan <- types.ConfigMessage{
				ProviderName:  "snapd",
				Configuration: configuration,
			}
			if p.Watch {
				// TODO: poll snapd for enabled or disabled snaps
			}
			return nil
		}
		notify := func(err error, time time.Duration) {
			log.Errorf("Provider connection error %+v, retrying in %s", err, time)
		}
		err := backoff.RetryNotify(safe.OperationWithRecover(operation), job.NewBackOff(backoff.NewExponentialBackOff()), notify)
		if err != nil {
			log.Errorf("Cannot connect to snapd server %+v", err)
		}
	})

	return nil
}

func listEnabledSnaps(snapdClient *client.Client) ([]snapData, error) {
	snapList, err := snapdClient.List(nil, nil)
	if err != nil {
		return nil, err
	}

	var snapsInspected []snapData
	// get inspect snaps
	for _, snap := range snapList {
		if snap.Status == "active" {
			var configuration map[string]interface{}
			configuration, err = snapdClient.Conf(snap.Name, nil)
			if err != nil {
				return nil, err
			}
			properties := make(map[string]string, len(configuration))
			flattenNestedMaps("", &properties, configuration)
			sData := snapData{
				SnapName:   snap.Name,
				Properties: properties,
			}
			snapsInspected = append(snapsInspected, sData)
		}
	}
	return snapsInspected, nil
}

func flattenNestedMaps(prefix string, properties *map[string]string, nestedMap map[string]interface{}) {
	for k, v := range nestedMap {
		var name string
		if len(prefix) == 0 {
			name = k
		} else {
			name = prefix + "." + k
		}
		switch vv := v.(type) {
		case map[string]interface{}:
			flattenNestedMaps(name, properties, vv)
		case string:
			(*properties)[name] = vv
		case json.Number:
			(*properties)[name] = string(vv)
		case bool:
			(*properties)[name] = strconv.FormatBool(vv)
		default:
			fmt.Println(name, "is of a type I don't know how to handle")
		}
	}
}

func (p *Provider) buildConfiguration(snapsInspected []snapData) *types.Configuration {
	snapFuncMap := template.FuncMap{
		"getLabelValue": label.GetStringValue,
		"getSubDomain":  getSubDomain,
		"getDomain":     label.GetFuncString(label.TraefikDomain, p.Domain),

		// Backend functions
		"getIPAddress":      p.getIPAddress,
		"getServers":        p.getServers,
		"getMaxConn":        label.GetMaxConn,
		"getHealthCheck":    label.GetHealthCheck,
		"getBuffering":      label.GetBuffering,
		"getCircuitBreaker": label.GetCircuitBreaker,
		"getLoadBalancer":   label.GetLoadBalancer,

		// Frontend functions
		"getBackendName":       getBackendName,
		"getPriority":          label.GetFuncInt(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getPassHostHeader":    label.GetFuncBool(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPassTLSCert":       label.GetFuncBool(label.TraefikFrontendPassTLSCert, label.DefaultPassTLSCert),
		"getPassTLSClientCert": label.GetTLSClientCert,
		"getEntryPoints":       label.GetFuncSliceString(label.TraefikFrontendEntryPoints),
		"getBasicAuth":         label.GetFuncSliceString(label.TraefikFrontendAuthBasic), // Deprecated
		"getAuth":              label.GetAuth,
		"getFrontendRule":      p.getFrontendRule,
		"getRedirect":          label.GetRedirect,
		"getErrorPages":        label.GetErrorPages,
		"getRateLimit":         label.GetRateLimit,
		"getHeaders":           label.GetHeaders,
		"getWhiteList":         label.GetWhiteList,
	}

	// filter snaps
	filteredSnaps := fun.Filter(p.containerFilter, snapsInspected).([]snapData)

	frontends := make(map[string][]snapData)
	servers := make(map[string][]snapData)

	serviceNames := make(map[string]struct{})

	for idx, snap := range filteredSnaps {
		segmentLabels := label.ExtractTraefikLabels(snap.Properties)
		for segmentName, labels := range segmentLabels {
			snap.SegmentProperties = labels
			snap.SegmentName = segmentName

			serviceNamesKey := snap.SnapName + segmentName

			if _, exists := serviceNames[serviceNamesKey]; !exists {
				frontendName := p.getFrontendName(snap, idx)
				frontends[frontendName] = append(frontends[frontendName], snap)
				if len(serviceNamesKey) > 0 {
					serviceNames[serviceNamesKey] = struct{}{}
				}
			}

			// Backends
			backendName := getBackendName(snap)

			// Servers
			servers[backendName] = append(servers[backendName], snap)
		}
	}

	templateObjects := struct {
		Snaps     []snapData
		Frontends map[string][]snapData
		Servers   map[string][]snapData
		Domain    string
	}{
		Snaps:     filteredSnaps,
		Frontends: frontends,
		Servers:   servers,
		Domain:    p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/snapd.tmpl", snapFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func (p *Provider) containerFilter(snap snapData) bool {
	if !label.IsEnabled(snap.Properties, p.ExposedByDefault) {
		log.Debugf("Filtering disabled snap %s", snap.SnapName)
		return false
	}

	segmentLabels := label.ExtractTraefikLabels(snap.Properties)

	var errPort error
	for segmentName, labels := range segmentLabels {
		errPort = checkSegmentPort(labels, segmentName)

		if len(p.getFrontendRule(snap, labels)) == 0 {
			log.Debugf("Filtering snap with empty frontend rule %s %s", snap.SnapName, segmentName)
			return false
		}
	}

	if errPort != nil {
		log.Debugf("Filtering snap without port, %s: %v", snap.SnapName, errPort)
		return false
	}

	constraintTags := label.SplitAndTrimString(snap.Properties[label.TraefikTags], ",")
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Snap %s pruned by %q constraint", snap.SnapName, failingConstraint.String())
		}
		return false
	}

	return true
}

func checkSegmentPort(labels map[string]string, segmentName string) error {
	if port, ok := labels[label.TraefikPort]; ok {
		_, err := strconv.Atoi(port)
		if err != nil {
			return fmt.Errorf("invalid port value %q for the segment %q: %v", port, segmentName, err)
		}
	} else {
		return fmt.Errorf("port label is missing, please use %s as default value or define port label for all segments ('traefik.<segment_name>.port')", label.TraefikPort)
	}
	return nil
}

func (p *Provider) getFrontendName(snap snapData, idx int) string {
	var name string
	if len(snap.SegmentName) > 0 {
		name = snap.SegmentName + "-" + getBackendName(snap)
	} else {
		name = p.getFrontendRule(snap, snap.SegmentProperties) + "-" + strconv.Itoa(idx)
	}

	return provider.Normalize(name)
}

func (p *Provider) getFrontendRule(snap snapData, segmentLabels map[string]string) string {
	if value := label.GetStringValue(segmentLabels, label.TraefikFrontendRule, ""); len(value) != 0 {
		return value
	}

	domain := label.GetStringValue(segmentLabels, label.TraefikDomain, p.Domain)

	if len(domain) > 0 {
		return "Host:" + getSubDomain(snap.SnapName) + "." + domain
	}

	return ""
}

// Escape beginning slash "/", convert all others to dash "-", and convert underscores "_" to dash "-"
func getSubDomain(name string) string {
	return strings.Replace(strings.Replace(strings.TrimPrefix(name, "/"), "/", "-", -1), "_", "-", -1)
}

func getBackendName(snap snapData) string {
	if len(snap.SegmentName) > 0 {
		return getSegmentBackendName(snap)
	}

	return getDefaultBackendName(snap)
}

func getSegmentBackendName(snap snapData) string {
	serviceName := snap.SnapName
	if value := label.GetStringValue(snap.SegmentProperties, label.TraefikBackend, ""); len(value) > 0 {
		return provider.Normalize(serviceName + "-" + value)
	}

	return provider.Normalize(serviceName + "-" + snap.SegmentName)
}

func getDefaultBackendName(snap snapData) string {
	if value := label.GetStringValue(snap.SegmentProperties, label.TraefikBackend, ""); len(value) != 0 {
		return provider.Normalize(value)
	}

	return provider.Normalize(snap.SnapName)
}

func (p *Provider) getIPAddress(snap snapData) string {
	return "127.0.0.1"
}

func (p *Provider) getServers(snaps []snapData) map[string]types.Server {
	var servers map[string]types.Server

	for _, snap := range snaps {
		port := label.GetStringValue(snap.Properties, label.TraefikPort, "")
		if len(port) == 0 {
			log.Warnf("No port defined for %q.", snap.SnapName)
			continue
		}

		if servers == nil {
			servers = make(map[string]types.Server)
		}

		protocol := label.GetStringValue(snap.SegmentProperties, label.TraefikProtocol, label.DefaultProtocol)

		serverURL := fmt.Sprintf("%s://%s", protocol, net.JoinHostPort("127.0.0.1", port))

		serverName := getServerName(snap.SnapName, serverURL)
		if _, exist := servers[serverName]; exist {
			log.Debugf("Skipping server %q with the same URL.", serverName)
			continue
		}

		servers[serverName] = types.Server{
			URL:    serverURL,
			Weight: label.GetIntValue(snap.SegmentProperties, label.TraefikWeight, label.DefaultWeight),
		}
	}

	return servers
}

func getServerName(snapName, url string) string {
	hash := md5.New()
	_, err := hash.Write([]byte(url))
	if err != nil {
		// Impossible case
		log.Errorf("Fail to hash server URL %q", url)
	}

	return provider.Normalize("server-" + snapName + "-" + hex.EncodeToString(hash.Sum(nil)))
}
