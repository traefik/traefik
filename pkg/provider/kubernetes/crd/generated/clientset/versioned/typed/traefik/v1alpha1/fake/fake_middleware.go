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

package fake

import (
	"context"

	v1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeMiddlewares implements MiddlewareInterface
type FakeMiddlewares struct {
	Fake *FakeTraefikV1alpha1
	ns   string
}

var middlewaresResource = schema.GroupVersionResource{Group: "traefik.containo.us", Version: "v1alpha1", Resource: "middlewares"}

var middlewaresKind = schema.GroupVersionKind{Group: "traefik.containo.us", Version: "v1alpha1", Kind: "Middleware"}

// Get takes name of the middleware, and returns the corresponding middleware object, and an error if there is any.
func (c *FakeMiddlewares) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Middleware, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(middlewaresResource, c.ns, name), &v1alpha1.Middleware{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Middleware), err
}

// List takes label and field selectors, and returns the list of Middlewares that match those selectors.
func (c *FakeMiddlewares) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.MiddlewareList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(middlewaresResource, middlewaresKind, c.ns, opts), &v1alpha1.MiddlewareList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.MiddlewareList{ListMeta: obj.(*v1alpha1.MiddlewareList).ListMeta}
	for _, item := range obj.(*v1alpha1.MiddlewareList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested middlewares.
func (c *FakeMiddlewares) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(middlewaresResource, c.ns, opts))

}

// Create takes the representation of a middleware and creates it.  Returns the server's representation of the middleware, and an error, if there is any.
func (c *FakeMiddlewares) Create(ctx context.Context, middleware *v1alpha1.Middleware, opts v1.CreateOptions) (result *v1alpha1.Middleware, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(middlewaresResource, c.ns, middleware), &v1alpha1.Middleware{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Middleware), err
}

// Update takes the representation of a middleware and updates it. Returns the server's representation of the middleware, and an error, if there is any.
func (c *FakeMiddlewares) Update(ctx context.Context, middleware *v1alpha1.Middleware, opts v1.UpdateOptions) (result *v1alpha1.Middleware, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(middlewaresResource, c.ns, middleware), &v1alpha1.Middleware{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Middleware), err
}

// Delete takes name of the middleware and deletes it. Returns an error if one occurs.
func (c *FakeMiddlewares) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(middlewaresResource, c.ns, name), &v1alpha1.Middleware{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeMiddlewares) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(middlewaresResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.MiddlewareList{})
	return err
}

// Patch applies the patch and returns the patched middleware.
func (c *FakeMiddlewares) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Middleware, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(middlewaresResource, c.ns, name, pt, data, subresources...), &v1alpha1.Middleware{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Middleware), err
}
