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

package v1alpha1

import (
	"context"
	"time"

	scheme "github.com/containous/traefik/v2/pkg/provider/knative/crd/generated/networking/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
	v1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

// DomainsGetter has a method to return a DomainInterface.
// A group's client should implement this interface.
type DomainsGetter interface {
	Domains() DomainInterface
}

// DomainInterface has methods to work with Domain resources.
type DomainInterface interface {
	Create(ctx context.Context, domain *v1alpha1.Domain, opts v1.CreateOptions) (*v1alpha1.Domain, error)
	Update(ctx context.Context, domain *v1alpha1.Domain, opts v1.UpdateOptions) (*v1alpha1.Domain, error)
	UpdateStatus(ctx context.Context, domain *v1alpha1.Domain, opts v1.UpdateOptions) (*v1alpha1.Domain, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.Domain, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.DomainList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Domain, err error)
	DomainExpansion
}

// domains implements DomainInterface
type domains struct {
	client rest.Interface
}

// newDomains returns a Domains
func newDomains(c *NetworkingV1alpha1Client) *domains {
	return &domains{
		client: c.RESTClient(),
	}
}

// Get takes name of the domain, and returns the corresponding domain object, and an error if there is any.
func (c *domains) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Domain, err error) {
	result = &v1alpha1.Domain{}
	err = c.client.Get().
		Resource("domains").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of Domains that match those selectors.
func (c *domains) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.DomainList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.DomainList{}
	err = c.client.Get().
		Resource("domains").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested domains.
func (c *domains) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("domains").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a domain and creates it.  Returns the server's representation of the domain, and an error, if there is any.
func (c *domains) Create(ctx context.Context, domain *v1alpha1.Domain, opts v1.CreateOptions) (result *v1alpha1.Domain, err error) {
	result = &v1alpha1.Domain{}
	err = c.client.Post().
		Resource("domains").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(domain).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a domain and updates it. Returns the server's representation of the domain, and an error, if there is any.
func (c *domains) Update(ctx context.Context, domain *v1alpha1.Domain, opts v1.UpdateOptions) (result *v1alpha1.Domain, err error) {
	result = &v1alpha1.Domain{}
	err = c.client.Put().
		Resource("domains").
		Name(domain.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(domain).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *domains) UpdateStatus(ctx context.Context, domain *v1alpha1.Domain, opts v1.UpdateOptions) (result *v1alpha1.Domain, err error) {
	result = &v1alpha1.Domain{}
	err = c.client.Put().
		Resource("domains").
		Name(domain.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(domain).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the domain and deletes it. Returns an error if one occurs.
func (c *domains) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("domains").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *domains) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("domains").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched domain.
func (c *domains) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Domain, err error) {
	result = &v1alpha1.Domain{}
	err = c.client.Patch(pt).
		Resource("domains").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
