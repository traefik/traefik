package dynamodb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/traefik/traefik/v2/pkg/config/dynamic"
	"github.com/traefik/traefik/v2/pkg/log"
	"github.com/traefik/traefik/v2/pkg/provider"
)

// buildConfiguration retrieves items from dynamodb and converts them into Backends and Frontends in a Configuration
func (p *Provider) buildConfiguration(ctx context.Context, client *dynamoClient) (*dynamic.Configuration, error) {
	logger := log.FromContext(ctx)
	configurations := make(map[string]*dynamic.Configuration)

	items, err := p.scanTable(client, logger)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Number of Items retrieved from Provider: %d", len(items))

	// unmarshal dynamoAttributes into Backends and Frontends
	for i, item := range items {
		logger.Debugf("Provider Item: %d\n%v", i, item)
		name, exists := item["name"]
		if !exists || name.S == nil || *name.S == "" {
			logger.Warnf("Item %v in dynamodb table does not have a name, skipping", item)
		}

		conf := &dynamic.Configuration{
			HTTP: &dynamic.HTTPConfiguration{},
			TCP:  &dynamic.TCPConfiguration{},
			UDP:  &dynamic.UDPConfiguration{},
		}

		itemName := *name.S
		normalizedName := provider.Normalize(*name.S)
		atLeastOne := false

		// verify the type of each item by checking to see if it has
		// the corresponding type, service, router or middleware map
		if service, exists := item["service"]; exists {
			logger.Debug("Unmarshaling service from Provider...")
			err := p.buildServiceConfiguration(service, conf, itemName)
			if err != nil {
				logger.Errorf(err.Error())
				continue
			}

			atLeastOne = true
		}

		if router, exists := item["router"]; exists {
			logger.Debug("Unmarshaling router from Provider...")
			err := p.buildRouterConfiguration(router, conf, itemName)
			if err != nil {
				logger.Errorf(err.Error())
				continue
			}

			atLeastOne = true
		}

		if router, exists := item["middleware"]; exists {
			logger.Debug("Unmarshaling middleware from Provider...")
			err := p.buildMiddlewareConfiguration(router, conf, normalizedName)
			if err != nil {
				logger.Errorf(err.Error())
				continue
			}

			atLeastOne = true
		}

		if atLeastOne {
			configurations[normalizedName] = conf
		} else {
			logger.Warnf("Error in format of Provider Item: %v", item)
		}
	}

	return provider.Merge(ctx, configurations), nil
}

func (p *Provider) buildServiceConfiguration(value *dynamodb.AttributeValue, conf *dynamic.Configuration, name string) error {
	serviceType, exists := value.M["type"]

	if !exists || *serviceType.S == "http" {
		if exists {
			delete(value.M, "type")
		}

		tmpService := &dynamic.Service{}
		err := dynamodbattribute.Unmarshal(value, tmpService)
		if err != nil {
			return err
		}

		if len(conf.HTTP.Services) == 0 {
			conf.HTTP.Services = make(map[string]*dynamic.Service)
		}

		conf.HTTP.Services[name] = tmpService
	} else if *serviceType.S == "tcp" {
		delete(value.M, "type")

		tmpService := &dynamic.TCPService{}
		err := dynamodbattribute.Unmarshal(value, tmpService)
		if err != nil {
			return err
		}

		if len(conf.TCP.Services) == 0 {
			conf.TCP.Services = make(map[string]*dynamic.TCPService)
		}

		conf.TCP.Services[name] = tmpService
	} else if *serviceType.S == "udp" {
		delete(value.M, "type")

		tmpService := &dynamic.UDPService{}
		err := dynamodbattribute.Unmarshal(value, tmpService)
		if err != nil {
			return err
		}

		if len(conf.UDP.Services) == 0 {
			conf.UDP.Services = make(map[string]*dynamic.UDPService)
		}

		conf.UDP.Services[name] = tmpService
	} else {
		return fmt.Errorf("unknown service type \"%s\"", *serviceType.S)
	}

	return nil
}

func (p *Provider) buildRouterConfiguration(value *dynamodb.AttributeValue, conf *dynamic.Configuration, name string) error {
	routerType, exists := value.M["type"]

	if !exists || *routerType.S == "http" {
		if exists {
			delete(value.M, "type")
		}

		tmpRouter := &dynamic.Router{}
		err := dynamodbattribute.Unmarshal(value, tmpRouter)
		if err != nil {
			return err
		}

		if len(conf.HTTP.Routers) == 0 {
			conf.HTTP.Routers = make(map[string]*dynamic.Router)
		}

		conf.HTTP.Routers[name] = tmpRouter
	} else if *routerType.S == "tcp" {
		delete(value.M, "type")

		tmpRouter := &dynamic.TCPRouter{}
		err := dynamodbattribute.Unmarshal(value, tmpRouter)
		if err != nil {
			return err
		}

		if len(conf.TCP.Routers) == 0 {
			conf.TCP.Routers = make(map[string]*dynamic.TCPRouter)
		}

		conf.TCP.Routers[name] = tmpRouter
	} else if *routerType.S == "udp" {
		delete(value.M, "type")

		tmpRouter := &dynamic.UDPRouter{}
		err := dynamodbattribute.Unmarshal(value, tmpRouter)
		if err != nil {
			return err
		}

		if len(conf.UDP.Routers) == 0 {
			conf.UDP.Routers = make(map[string]*dynamic.UDPRouter)
		}

		conf.UDP.Routers[name] = tmpRouter
	} else {
		return fmt.Errorf("unknown router type \"%s\"", *routerType.S)
	}

	return nil
}

func (p *Provider) buildMiddlewareConfiguration(value *dynamodb.AttributeValue, conf *dynamic.Configuration, name string) error {
	middlewareType, exists := value.M["type"]

	if !exists || *middlewareType.S == "http" {
		if exists {
			delete(value.M, "type")
		}

		tmpMiddleware := &dynamic.Middleware{}
		err := dynamodbattribute.Unmarshal(value, tmpMiddleware)
		if err != nil {
			return err
		}

		if len(conf.HTTP.Middlewares) == 0 {
			conf.HTTP.Middlewares = make(map[string]*dynamic.Middleware)
		}

		conf.HTTP.Middlewares[name] = tmpMiddleware
	} else if *middlewareType.S == "tcp" {
		delete(value.M, "type")

		tmpMiddleware := &dynamic.TCPMiddleware{}
		err := dynamodbattribute.Unmarshal(value, tmpMiddleware)
		if err != nil {
			return err
		}

		if len(conf.TCP.Routers) == 0 {
			conf.TCP.Middlewares = make(map[string]*dynamic.TCPMiddleware)
		}

		conf.TCP.Middlewares[name] = tmpMiddleware
	} else {
		return fmt.Errorf("unknown middleware type \"%s\"", *middlewareType.S)
	}

	return nil
}
