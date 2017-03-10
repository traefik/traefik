package project

import (
	"golang.org/x/net/context"
)

// Container defines what a libcompose container provides.
type Container interface {
	ID() (string, error)
	Name() string
	Port(ctx context.Context, port string) (string, error)
	IsRunning(ctx context.Context) (bool, error)
}
