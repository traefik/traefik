/*
Copyright 2016 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bigtable

import (
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/bigtable/bttest"
	"golang.org/x/net/context"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

var legacyUseProd string
var integrationConfig IntegrationTestConfig

func init() {
	c := &integrationConfig

	flag.BoolVar(&c.UseProd, "it.use-prod", false, "Use remote bigtable instead of local emulator")
	flag.StringVar(&c.AdminEndpoint, "it.admin-endpoint", "", "Admin api host and port")
	flag.StringVar(&c.DataEndpoint, "it.data-endpoint", "", "Data api host and port")
	flag.StringVar(&c.Project, "it.project", "", "Project to use for integration test")
	flag.StringVar(&c.Instance, "it.instance", "", "Bigtable instance to use")
	flag.StringVar(&c.Cluster, "it.cluster", "", "Bigtable cluster to use")
	flag.StringVar(&c.Table, "it.table", "", "Bigtable table to create")

	// Backwards compat
	flag.StringVar(&legacyUseProd, "use_prod", "", `DEPRECATED: if set to "proj,instance,table", run integration test against production`)

}

// IntegrationTestConfig contains parameters to pick and setup a IntegrationEnv for testing
type IntegrationTestConfig struct {
	UseProd       bool
	AdminEndpoint string
	DataEndpoint  string
	Project       string
	Instance      string
	Cluster       string
	Table         string
}

// IntegrationEnv represents a testing environment.
// The environment can be implemented using production or an emulator
type IntegrationEnv interface {
	Config() IntegrationTestConfig
	NewAdminClient() (*AdminClient, error)
	NewClient() (*Client, error)
	Close()
}

// NewIntegrationEnv creates a new environment based on the command line args
func NewIntegrationEnv() (IntegrationEnv, error) {
	c := integrationConfig

	if legacyUseProd != "" {
		fmt.Println("WARNING: using legacy commandline arg -use_prod, please switch to -it.*")
		parts := strings.SplitN(legacyUseProd, ",", 3)
		c.UseProd = true
		c.Project = parts[0]
		c.Instance = parts[1]
		c.Table = parts[2]
	}

	if integrationConfig.UseProd {
		return NewProdEnv(c)
	} else {
		return NewEmulatedEnv(c)
	}
}

// EmulatedEnv encapsulates the state of an emulator
type EmulatedEnv struct {
	config IntegrationTestConfig
	server *bttest.Server
}

// NewEmulatedEnv builds and starts the emulator based environment
func NewEmulatedEnv(config IntegrationTestConfig) (*EmulatedEnv, error) {
	srv, err := bttest.NewServer("127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	if config.Project == "" {
		config.Project = "project"
	}
	if config.Instance == "" {
		config.Instance = "instance"
	}
	if config.Table == "" {
		config.Table = "mytable"
	}
	config.AdminEndpoint = srv.Addr
	config.DataEndpoint = srv.Addr

	env := &EmulatedEnv{
		config: config,
		server: srv,
	}
	return env, nil
}

// Close stops & cleans up the emulator
func (e *EmulatedEnv) Close() {
	e.server.Close()
}

// Config gets the config used to build this environment
func (e *EmulatedEnv) Config() IntegrationTestConfig {
	return e.config
}

// NewAdminClient builds a new connected admin client for this environment
func (e *EmulatedEnv) NewAdminClient() (*AdminClient, error) {
	timeout := 20 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	conn, err := grpc.Dial(e.server.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return NewAdminClient(ctx, e.config.Project, e.config.Instance, option.WithGRPCConn(conn))
}

// NewClient builds a new connected data client for this environment
func (e *EmulatedEnv) NewClient() (*Client, error) {
	timeout := 20 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	conn, err := grpc.Dial(e.server.Addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, e.config.Project, e.config.Instance, option.WithGRPCConn(conn))
}

// ProdEnv encapsulates the state necessary to connect to the external Bigtable service
type ProdEnv struct {
	config IntegrationTestConfig
}

// NewProdEnv builds the environment representation
func NewProdEnv(config IntegrationTestConfig) (*ProdEnv, error) {
	if config.Project == "" {
		return nil, errors.New("Project not set")
	}
	if config.Instance == "" {
		return nil, errors.New("Instance not set")
	}
	if config.Table == "" {
		return nil, errors.New("Table not set")
	}

	return &ProdEnv{config}, nil
}

// Close is a no-op for production environments
func (e *ProdEnv) Close() {}

// Config gets the config used to build this environment
func (e *ProdEnv) Config() IntegrationTestConfig {
	return e.config
}

// NewAdminClient builds a new connected admin client for this environment
func (e *ProdEnv) NewAdminClient() (*AdminClient, error) {
	timeout := 20 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	var clientOpts []option.ClientOption
	if endpoint := e.config.AdminEndpoint; endpoint != "" {
		clientOpts = append(clientOpts, option.WithEndpoint(endpoint))
	}
	return NewAdminClient(ctx, e.config.Project, e.config.Instance, clientOpts...)
}

// NewClient builds a connected data client for this environment
func (e *ProdEnv) NewClient() (*Client, error) {
	timeout := 20 * time.Second
	ctx, _ := context.WithTimeout(context.Background(), timeout)
	var clientOpts []option.ClientOption
	if endpoint := e.config.DataEndpoint; endpoint != "" {
		clientOpts = append(clientOpts, option.WithEndpoint(endpoint))
	}
	return NewClient(ctx, e.config.Project, e.config.Instance, clientOpts...)
}
