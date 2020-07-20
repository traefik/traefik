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

// FakeRealms implements RealmInterface
type FakeRealms struct {
	Fake *FakeNetworkingV1alpha1
}

var realmsResource = schema.GroupVersionResource{Group: "networking.internal.knative.dev", Version: "v1alpha1", Resource: "realms"}

var realmsKind = schema.GroupVersionKind{Group: "networking.internal.knative.dev", Version: "v1alpha1", Kind: "Realm"}

// Get takes name of the realm, and returns the corresponding realm object, and an error if there is any.
func (c *FakeRealms) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.Realm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(realmsResource, name), &v1alpha1.Realm{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Realm), err
}

// List takes label and field selectors, and returns the list of Realms that match those selectors.
func (c *FakeRealms) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.RealmList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(realmsResource, realmsKind, opts), &v1alpha1.RealmList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.RealmList{ListMeta: obj.(*v1alpha1.RealmList).ListMeta}
	for _, item := range obj.(*v1alpha1.RealmList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested realms.
func (c *FakeRealms) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(realmsResource, opts))
}

// Create takes the representation of a realm and creates it.  Returns the server's representation of the realm, and an error, if there is any.
func (c *FakeRealms) Create(ctx context.Context, realm *v1alpha1.Realm, opts v1.CreateOptions) (result *v1alpha1.Realm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(realmsResource, realm), &v1alpha1.Realm{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Realm), err
}

// Update takes the representation of a realm and updates it. Returns the server's representation of the realm, and an error, if there is any.
func (c *FakeRealms) Update(ctx context.Context, realm *v1alpha1.Realm, opts v1.UpdateOptions) (result *v1alpha1.Realm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(realmsResource, realm), &v1alpha1.Realm{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Realm), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeRealms) UpdateStatus(ctx context.Context, realm *v1alpha1.Realm, opts v1.UpdateOptions) (*v1alpha1.Realm, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(realmsResource, "status", realm), &v1alpha1.Realm{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Realm), err
}

// Delete takes name of the realm and deletes it. Returns an error if one occurs.
func (c *FakeRealms) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(realmsResource, name), &v1alpha1.Realm{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeRealms) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(realmsResource, listOpts)

	_, err := c.Fake.Invokes(action, &v1alpha1.RealmList{})
	return err
}

// Patch applies the patch and returns the patched realm.
func (c *FakeRealms) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.Realm, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(realmsResource, name, pt, data, subresources...), &v1alpha1.Realm{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.Realm), err
}
