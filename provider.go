package main

type Provider interface {
	Provide(chan<- *Configuration)
}
