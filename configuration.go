package main

type GlobalConfiguration struct {
	Port              string
	GraceTimeOut      int64
	AccessLogsFile    string
	TraefikLogsFile   string
	TraefikLogsStdout bool
	CertFile, KeyFile string
	LogLevel          string
	Docker            *DockerProvider
	File              *FileProvider
	Web               *WebProvider
	Marathon          *MarathonProvider
	Consul            *ConsulProvider
}

func NewGlobalConfiguration() *GlobalConfiguration {
	globalConfiguration := new(GlobalConfiguration)
	// default values
	globalConfiguration.Port = ":80"
	globalConfiguration.GraceTimeOut = 10
	globalConfiguration.LogLevel = "ERROR"
	globalConfiguration.TraefikLogsStdout = true

	return globalConfiguration
}

type Backend struct {
	Servers map[string]Server
}

type Server struct {
	URL    string `json:"Url"`
	Weight int
}

type Route struct {
	Rule  string
	Value string
}

type Frontend struct {
	Backend string
	Routes  map[string]Route
}

type Configuration struct {
	Backends  map[string]Backend
	Frontends map[string]Frontend
}
