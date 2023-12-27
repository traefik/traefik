/*
The MIT License (MIT)

Copyright (c) 2016-2020 Containous SAS; 2020-2024 Traefik Labs

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

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefikio/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// IngressRouteTCPLister helps list IngressRouteTCPs.
// All objects returned here must be treated as read-only.
type IngressRouteTCPLister interface {
	// List lists all IngressRouteTCPs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.IngressRouteTCP, err error)
	// IngressRouteTCPs returns an object that can list and get IngressRouteTCPs.
	IngressRouteTCPs(namespace string) IngressRouteTCPNamespaceLister
	IngressRouteTCPListerExpansion
}

// ingressRouteTCPLister implements the IngressRouteTCPLister interface.
type ingressRouteTCPLister struct {
	indexer cache.Indexer
}

// NewIngressRouteTCPLister returns a new IngressRouteTCPLister.
func NewIngressRouteTCPLister(indexer cache.Indexer) IngressRouteTCPLister {
	return &ingressRouteTCPLister{indexer: indexer}
}

// List lists all IngressRouteTCPs in the indexer.
func (s *ingressRouteTCPLister) List(selector labels.Selector) (ret []*v1.IngressRouteTCP, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.IngressRouteTCP))
	})
	return ret, err
}

// IngressRouteTCPs returns an object that can list and get IngressRouteTCPs.
func (s *ingressRouteTCPLister) IngressRouteTCPs(namespace string) IngressRouteTCPNamespaceLister {
	return ingressRouteTCPNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// IngressRouteTCPNamespaceLister helps list and get IngressRouteTCPs.
// All objects returned here must be treated as read-only.
type IngressRouteTCPNamespaceLister interface {
	// List lists all IngressRouteTCPs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1.IngressRouteTCP, err error)
	// Get retrieves the IngressRouteTCP from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1.IngressRouteTCP, error)
	IngressRouteTCPNamespaceListerExpansion
}

// ingressRouteTCPNamespaceLister implements the IngressRouteTCPNamespaceLister
// interface.
type ingressRouteTCPNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all IngressRouteTCPs in the indexer for a given namespace.
func (s ingressRouteTCPNamespaceLister) List(selector labels.Selector) (ret []*v1.IngressRouteTCP, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.IngressRouteTCP))
	})
	return ret, err
}

// Get retrieves the IngressRouteTCP from the indexer for a given namespace and name.
func (s ingressRouteTCPNamespaceLister) Get(name string) (*v1.IngressRouteTCP, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1.Resource("ingressroutetcp"), name)
	}
	return obj.(*v1.IngressRouteTCP), nil
}
