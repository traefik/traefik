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
	Url    string
	Weight int
}

type Rule struct {
	Category string
	Value    string
}

type Route struct {
	Backend string
	Rules   map[string]Rule
}

type Configuration struct {
	Backends map[string]Backend
	Routes   map[string]Route
}
