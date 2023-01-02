/*
The MIT License (MIT)

Copyright (c) 2016-2020 Containous SAS; 2020-2023 Traefik Labs

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

package v1alpha1

import (
	"context"
	"time"

	scheme "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned/scheme"
	v1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// ServersTransportTCPsGetter has a method to return a ServersTransportTCPInterface.
// A group's client should implement this interface.
type ServersTransportTCPsGetter interface {
	ServersTransportTCPs(namespace string) ServersTransportTCPInterface
}

// ServersTransportTCPInterface has methods to work with ServersTransportTCP resources.
type ServersTransportTCPInterface interface {
	Create(ctx context.Context, serversTransportTCP *v1alpha1.ServersTransportTCP, opts v1.CreateOptions) (*v1alpha1.ServersTransportTCP, error)
	Update(ctx context.Context, serversTransportTCP *v1alpha1.ServersTransportTCP, opts v1.UpdateOptions) (*v1alpha1.ServersTransportTCP, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.ServersTransportTCP, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.ServersTransportTCPList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ServersTransportTCP, err error)
	ServersTransportTCPExpansion
}

// serversTransportTCPs implements ServersTransportTCPInterface
type serversTransportTCPs struct {
	client rest.Interface
	ns     string
}

// newServersTransportTCPs returns a ServersTransportTCPs
func newServersTransportTCPs(c *TraefikV1alpha1Client, namespace string) *serversTransportTCPs {
	return &serversTransportTCPs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the serversTransportTCP, and returns the corresponding serversTransportTCP object, and an error if there is any.
func (c *serversTransportTCPs) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.ServersTransportTCP, err error) {
	result = &v1alpha1.ServersTransportTCP{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serverstransporttcps").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of ServersTransportTCPs that match those selectors.
func (c *serversTransportTCPs) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.ServersTransportTCPList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.ServersTransportTCPList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("serverstransporttcps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested serversTransportTCPs.
func (c *serversTransportTCPs) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("serverstransporttcps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a serversTransportTCP and creates it.  Returns the server's representation of the serversTransportTCP, and an error, if there is any.
func (c *serversTransportTCPs) Create(ctx context.Context, serversTransportTCP *v1alpha1.ServersTransportTCP, opts v1.CreateOptions) (result *v1alpha1.ServersTransportTCP, err error) {
	result = &v1alpha1.ServersTransportTCP{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("serverstransporttcps").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(serversTransportTCP).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a serversTransportTCP and updates it. Returns the server's representation of the serversTransportTCP, and an error, if there is any.
func (c *serversTransportTCPs) Update(ctx context.Context, serversTransportTCP *v1alpha1.ServersTransportTCP, opts v1.UpdateOptions) (result *v1alpha1.ServersTransportTCP, err error) {
	result = &v1alpha1.ServersTransportTCP{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("serverstransporttcps").
		Name(serversTransportTCP.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(serversTransportTCP).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the serversTransportTCP and deletes it. Returns an error if one occurs.
func (c *serversTransportTCPs) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serverstransporttcps").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *serversTransportTCPs) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("serverstransporttcps").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched serversTransportTCP.
func (c *serversTransportTCPs) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.ServersTransportTCP, err error) {
	result = &v1alpha1.ServersTransportTCP{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("serverstransporttcps").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
