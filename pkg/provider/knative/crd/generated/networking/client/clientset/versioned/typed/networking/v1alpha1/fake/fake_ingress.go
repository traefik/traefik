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
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
	v1alpha1 "knative.dev/networking/pkg/apis/networking/v1alpha1"
)

// FakeIngresses implements IngressInterface
type FakeIngresses struct {
	Fake *FakeNetworkingV1alpha1
	ns   string
}

var ingressesResource = schema.GroupVersionResource{Group: "networking.internal.knative.dev", Version: "v1alpha1", Resource: "ingresses"}

var ingressesKind = schema.GroupVersionKind{Group: "networking.internal.knative.dev", Version: "v1alpha1", Kind: "Ingress"}

// Get takes name of the ingress, and returns the corresponding ingress object, and an error if there is any.
func (c *FakeIngresses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Ingress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(ingressesResource, c.ns, name), &v1alpha1.Ingress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Ingress), err
}

// List takes label and field selectors, and returns the list of Ingresses that match those selectors.
func (c *FakeIngresses) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.IngressList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(ingressesResource, ingressesKind, c.ns, opts), &v1alpha1.IngressList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.IngressList{ListMeta: obj.(*v1alpha1.IngressList).ListMeta}
	for _, item := range obj.(*v1alpha1.IngressList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested ingresses.
func (c *FakeIngresses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(ingressesResource, c.ns, opts))
}

// Create takes the representation of a ingress and creates it.  Returns the server's representation of the ingress, and an error, if there is any.
func (c *FakeIngresses) Create(ctx context.Context, ingress *v1alpha1.Ingress, opts v1.CreateOptions) (result *v1alpha1.Ingress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(ingressesResource, c.ns, ingress), &v1alpha1.Ingress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Ingress), err
}

// Update takes the representation of a ingress and updates it. Returns the server's representation of the ingress, and an error, if there is any.
func (c *FakeIngresses) Update(ctx context.Context, ingress *v1alpha1.Ingress, opts v1.UpdateOptions) (result *v1alpha1.Ingress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(ingressesResource, c.ns, ingress), &v1alpha1.Ingress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Ingress), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeIngresses) UpdateStatus(ctx context.Context, ingress *v1alpha1.Ingress, opts v1.UpdateOptions) (*v1alpha1.Ingress, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(ingressesResource, "status", c.ns, ingress), &v1alpha1.Ingress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Ingress), err
}

// Delete takes name of the ingress and deletes it. Returns an error if one occurs.
func (c *FakeIngresses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(ingressesResource, c.ns, name), &v1alpha1.Ingress{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeIngresses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(ingressesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.IngressList{})
	return err
}

// Patch applies the patch and returns the patched ingress.
func (c *FakeIngresses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Ingress, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(ingressesResource, c.ns, name, pt, data, subresources...), &v1alpha1.Ingress{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Ingress), err
}
