/*
The MIT License (MIT)

Copyright (c) 2016-2020 Containous SAS

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1alpha1 "github.com/containous/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned/typed/traefik/v1alpha1"
	rest "k8s.io/client-go/rest"
	testing "k8s.io/client-go/testing"
)

type FakeTraefikV1alpha1 struct {
	*testing.Fake
}

func (c *FakeTraefikV1alpha1) IngressRoutes(namespace string) v1alpha1.IngressRouteInterface {
	return &FakeIngressRoutes{c, namespace}
}

func (c *FakeTraefikV1alpha1) IngressRouteTCPs(namespace string) v1alpha1.IngressRouteTCPInterface {
	return &FakeIngressRouteTCPs{c, namespace}
}

func (c *FakeTraefikV1alpha1) IngressRouteUDPs(namespace string) v1alpha1.IngressRouteUDPInterface {
	return &FakeIngressRouteUDPs{c, namespace}
}

func (c *FakeTraefikV1alpha1) Middlewares(namespace string) v1alpha1.MiddlewareInterface {
	return &FakeMiddlewares{c, namespace}
}

func (c *FakeTraefikV1alpha1) ServersTransports(namespace string) v1alpha1.ServersTransportInterface {
	return &FakeServersTransports{c, namespace}
}

func (c *FakeTraefikV1alpha1) TLSOptions(namespace string) v1alpha1.TLSOptionInterface {
	return &FakeTLSOptions{c, namespace}
}

func (c *FakeTraefikV1alpha1) TLSStores(namespace string) v1alpha1.TLSStoreInterface {
	return &FakeTLSStores{c, namespace}
}

func (c *FakeTraefikV1alpha1) TraefikServices(namespace string) v1alpha1.TraefikServiceInterface {
	return &FakeTraefikServices{c, namespace}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *FakeTraefikV1alpha1) RESTClient() rest.Interface {
	var ret *rest.RESTClient
	return ret
}
