package main

type GlobalConfiguration struct {
	Port string
	GraceTimeOut int64
	Docker *DockerProvider
	File   *FileProvider
	Web    *WebProvider
	Marathon *MarathonProvider
}

func NewGlobalConfiguration() *GlobalConfiguration {
	globalConfiguration := new(GlobalConfiguration)
	// default values
	globalConfiguration.Port = ":8080"
	globalConfiguration.GraceTimeOut = 10

	return globalConfiguration
}

type Backend struct {
	Servers map[string]Server
}

type Server struct {
	Url string
	Weight int
}

type Rule struct {
	Category string
	Value    string
}

type Route struct {
	Backend string
	Rules    map[string]Rule
}

type Configuration struct {
	Backends map[string]Backend
	Routes   map[string]Route
}