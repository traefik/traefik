/*
Copyright 2014 Rohith All rights reserved.

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

package main

import (
	"flag"
	"os"

	marathon "github.com/gambol99/go-marathon"
	"github.com/golang/glog"
)

var marathon_url string

func init() {
	flag.StringVar(&marathon_url, "url", "http://127.0.0.1:8080", "the url for the marathon endpoint")
}

func Assert(err error) {
	if err != nil {
		glog.Fatalf("Failed, error: %s", err)
	}
}

func main() {
	flag.Parse()
	config := marathon.NewDefaultConfig()
	config.URL = marathon_url
	config.LogOutput = os.Stdout
	client, err := marathon.NewClient(config)
	if err != nil {
		glog.Fatalf("Failed to create a client for marathon, error: %s", err)
	}

	glog.Infof("Retrieving a list of groups")
	if groups, err := client.Groups(); err != nil {
		glog.Errorf("Failed to retrieve the groups from maratho, error: %s", err)
	} else {
		for _, group := range groups.Groups {
			glog.Infof("Found group: %s", group.ID)
		}
	}

	GROUP_NAME := "/product/group"

	found, err := client.HasGroup(GROUP_NAME)
	Assert(err)
	if found {
		glog.Infof("Deleting the grouy: %s, as it already exists", GROUP_NAME)
		id, err := client.DeleteGroup(GROUP_NAME)
		Assert(err)
		err = client.WaitOnDeployment(id.DeploymentID, 0)
		Assert(err)
	}

	/* step: the frontend app */
	frontend := marathon.NewDockerApplication()
	frontend.Name("/product/group/frontend")
	frontend.CPU(0.1).Memory(64).Storage(0.0).Count(2)
	frontend.Arg("/usr/sbin/apache2ctl").Arg("-D").Arg("FOREGROUND")
	frontend.AddEnv("NAME", "frontend_http")
	frontend.AddEnv("SERVICE_80_NAME", "frontend_http")
	frontend.AddEnv("SERVICE_443_NAME", "frontend_https")
	frontend.AddEnv("BACKEND_MYSQL", "/product/group/mysql/3306;3306")
	frontend.AddEnv("BACKEND_CACHE", "/product/group/cache/6379;6379")
	frontend.DependsOn("/product/group/cache")
	frontend.DependsOn("/product/group/mysql")
	frontend.Container.Docker.Container("quay.io/gambol99/apache-php:latest").Expose(80).Expose(443)
	_, err = frontend.CheckHTTP("/hostname.php", 80, 10)
	Assert(err)

	mysql := marathon.NewDockerApplication()
	mysql.Name("/product/group/mysql")
	mysql.CPU(0.1).Memory(128).Storage(0.0).Count(1)
	mysql.AddEnv("NAME", "group_cache")
	mysql.AddEnv("SERVICE_3306_NAME", "mysql")
	mysql.AddEnv("MYSQL_PASS", "mysql")
	mysql.Container.Docker.Container("tutum/mysql").Expose(3306)
	_, err = mysql.CheckTCP(3306, 10)
	Assert(err)

	redis := marathon.NewDockerApplication()
	redis.Name("/product/group/cache")
	redis.CPU(0.1).Memory(64).Storage(0.0).Count(2)
	redis.AddEnv("NAME", "group_cache")
	redis.AddEnv("SERVICE_6379_NAME", "redis")
	redis.Container.Docker.Container("redis:latest").Expose(6379)
	_, err = redis.CheckTCP(6379, 10)
	Assert(err)

	group := marathon.NewApplicationGroup(GROUP_NAME)
	group.App(frontend).App(redis).App(mysql)

	Assert(client.CreateGroup(group, true))
	glog.Infof("Successfully created the group: %s", group.ID)

	glog.Infof("Updating the group paramaters")
	frontend.Count(4)

	id, err := client.UpdateGroup(GROUP_NAME, group)
	Assert(err)
	glog.Infof("Successfully updated the group: %s, version: %s", group.ID, id.DeploymentID)
	Assert(client.WaitOnGroup(GROUP_NAME, 0))
}
