package main


type Backend struct {
	Servers map[string]Server
}

type Server struct {
	Url string
}

type Rule struct {
	Category string
	Value    string
}

type Route struct {
	Backends []string
	Rules    map[string]Rule
}

type Service struct {
	Backends map[string]Backend
	Routes   map[string]Route
}