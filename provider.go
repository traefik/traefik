package main

type Provider interface {
	Provide(configurationChan chan<- *Configuration)
}
