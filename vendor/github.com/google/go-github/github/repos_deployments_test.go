// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestRepositoriesService_ListDeployments(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/deployments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"environment": "test"})
		fmt.Fprint(w, `[{"id":1}, {"id":2}]`)
	})

	opt := &DeploymentsListOptions{Environment: "test"}
	deployments, _, err := client.Repositories.ListDeployments(context.Background(), "o", "r", opt)
	if err != nil {
		t.Errorf("Repositories.ListDeployments returned error: %v", err)
	}

	want := []*Deployment{{ID: Int(1)}, {ID: Int(2)}}
	if !reflect.DeepEqual(deployments, want) {
		t.Errorf("Repositories.ListDeployments returned %+v, want %+v", deployments, want)
	}
}

func TestRepositoriesService_GetDeployment(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/deployments/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":3}`)
	})

	deployment, _, err := client.Repositories.GetDeployment(context.Background(), "o", "r", 3)
	if err != nil {
		t.Errorf("Repositories.GetDeployment returned error: %v", err)
	}

	want := &Deployment{ID: Int(3)}

	if !reflect.DeepEqual(deployment, want) {
		t.Errorf("Repositories.GetDeployment returned %+v, want %+v", deployment, want)
	}
}

func TestRepositoriesService_CreateDeployment(t *testing.T) {
	setup()
	defer teardown()

	input := &DeploymentRequest{Ref: String("1111"), Task: String("deploy"), TransientEnvironment: Bool(true)}

	mux.HandleFunc("/repos/o/r/deployments", func(w http.ResponseWriter, r *http.Request) {
		v := new(DeploymentRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeDeploymentStatusPreview)
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"ref": "1111", "task": "deploy"}`)
	})

	deployment, _, err := client.Repositories.CreateDeployment(context.Background(), "o", "r", input)
	if err != nil {
		t.Errorf("Repositories.CreateDeployment returned error: %v", err)
	}

	want := &Deployment{Ref: String("1111"), Task: String("deploy")}
	if !reflect.DeepEqual(deployment, want) {
		t.Errorf("Repositories.CreateDeployment returned %+v, want %+v", deployment, want)
	}
}

func TestRepositoriesService_ListDeploymentStatuses(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/deployments/1/statuses", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testFormValues(t, r, values{"page": "2"})
		fmt.Fprint(w, `[{"id":1}, {"id":2}]`)
	})

	opt := &ListOptions{Page: 2}
	statutses, _, err := client.Repositories.ListDeploymentStatuses(context.Background(), "o", "r", 1, opt)
	if err != nil {
		t.Errorf("Repositories.ListDeploymentStatuses returned error: %v", err)
	}

	want := []*DeploymentStatus{{ID: Int(1)}, {ID: Int(2)}}
	if !reflect.DeepEqual(statutses, want) {
		t.Errorf("Repositories.ListDeploymentStatuses returned %+v, want %+v", statutses, want)
	}
}

func TestRepositoriesService_GetDeploymentStatus(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/repos/o/r/deployments/3/statuses/4", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypeDeploymentStatusPreview)
		fmt.Fprint(w, `{"id":4}`)
	})

	deploymentStatus, _, err := client.Repositories.GetDeploymentStatus(context.Background(), "o", "r", 3, 4)
	if err != nil {
		t.Errorf("Repositories.GetDeploymentStatus returned error: %v", err)
	}

	want := &DeploymentStatus{ID: Int(4)}
	if !reflect.DeepEqual(deploymentStatus, want) {
		t.Errorf("Repositories.GetDeploymentStatus returned %+v, want %+v", deploymentStatus, want)
	}
}

func TestRepositoriesService_CreateDeploymentStatus(t *testing.T) {
	setup()
	defer teardown()

	input := &DeploymentStatusRequest{State: String("inactive"), Description: String("deploy"), AutoInactive: Bool(false)}

	mux.HandleFunc("/repos/o/r/deployments/1/statuses", func(w http.ResponseWriter, r *http.Request) {
		v := new(DeploymentStatusRequest)
		json.NewDecoder(r.Body).Decode(v)

		testMethod(t, r, "POST")
		testHeader(t, r, "Accept", mediaTypeDeploymentStatusPreview)
		if !reflect.DeepEqual(v, input) {
			t.Errorf("Request body = %+v, want %+v", v, input)
		}

		fmt.Fprint(w, `{"state": "inactive", "description": "deploy"}`)
	})

	deploymentStatus, _, err := client.Repositories.CreateDeploymentStatus(context.Background(), "o", "r", 1, input)
	if err != nil {
		t.Errorf("Repositories.CreateDeploymentStatus returned error: %v", err)
	}

	want := &DeploymentStatus{State: String("inactive"), Description: String("deploy")}
	if !reflect.DeepEqual(deploymentStatus, want) {
		t.Errorf("Repositories.CreateDeploymentStatus returned %+v, want %+v", deploymentStatus, want)
	}
}
