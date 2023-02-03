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

	v1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeTLSStores implements TLSStoreInterface
type FakeTLSStores struct {
	Fake *FakeTraefikV1alpha1
	ns   string
}

var tlsstoresResource = schema.GroupVersionResource{Group: "traefik.containo.us", Version: "v1alpha1", Resource: "tlsstores"}

var tlsstoresKind = schema.GroupVersionKind{Group: "traefik.containo.us", Version: "v1alpha1", Kind: "TLSStore"}

// Get takes name of the tLSStore, and returns the corresponding tLSStore object, and an error if there is any.
func (c *FakeTLSStores) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.TLSStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(tlsstoresResource, c.ns, name), &v1alpha1.TLSStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TLSStore), err
}

// List takes label and field selectors, and returns the list of TLSStores that match those selectors.
func (c *FakeTLSStores) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.TLSStoreList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(tlsstoresResource, tlsstoresKind, c.ns, opts), &v1alpha1.TLSStoreList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.TLSStoreList{ListMeta: obj.(*v1alpha1.TLSStoreList).ListMeta}
	for _, item := range obj.(*v1alpha1.TLSStoreList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested tLSStores.
func (c *FakeTLSStores) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(tlsstoresResource, c.ns, opts))

}

// Create takes the representation of a tLSStore and creates it.  Returns the server's representation of the tLSStore, and an error, if there is any.
func (c *FakeTLSStores) Create(ctx context.Context, tLSStore *v1alpha1.TLSStore, opts v1.CreateOptions) (result *v1alpha1.TLSStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(tlsstoresResource, c.ns, tLSStore), &v1alpha1.TLSStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TLSStore), err
}

// Update takes the representation of a tLSStore and updates it. Returns the server's representation of the tLSStore, and an error, if there is any.
func (c *FakeTLSStores) Update(ctx context.Context, tLSStore *v1alpha1.TLSStore, opts v1.UpdateOptions) (result *v1alpha1.TLSStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(tlsstoresResource, c.ns, tLSStore), &v1alpha1.TLSStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TLSStore), err
}

// Delete takes name of the tLSStore and deletes it. Returns an error if one occurs.
func (c *FakeTLSStores) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(tlsstoresResource, c.ns, name), &v1alpha1.TLSStore{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeTLSStores) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(tlsstoresResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.TLSStoreList{})
	return err
}

// Patch applies the patch and returns the patched tLSStore.
func (c *FakeTLSStores) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.TLSStore, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(tlsstoresResource, c.ns, name, pt, data, subresources...), &v1alpha1.TLSStore{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.TLSStore), err
}
