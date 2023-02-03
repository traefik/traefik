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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/traefik/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// MiddlewareLister helps list Middlewares.
// All objects returned here must be treated as read-only.
type MiddlewareLister interface {
	// List lists all Middlewares in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Middleware, err error)
	// Middlewares returns an object that can list and get Middlewares.
	Middlewares(namespace string) MiddlewareNamespaceLister
	MiddlewareListerExpansion
}

// middlewareLister implements the MiddlewareLister interface.
type middlewareLister struct {
	indexer cache.Indexer
}

// NewMiddlewareLister returns a new MiddlewareLister.
func NewMiddlewareLister(indexer cache.Indexer) MiddlewareLister {
	return &middlewareLister{indexer: indexer}
}

// List lists all Middlewares in the indexer.
func (s *middlewareLister) List(selector labels.Selector) (ret []*v1alpha1.Middleware, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Middleware))
	})
	return ret, err
}

// Middlewares returns an object that can list and get Middlewares.
func (s *middlewareLister) Middlewares(namespace string) MiddlewareNamespaceLister {
	return middlewareNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// MiddlewareNamespaceLister helps list and get Middlewares.
// All objects returned here must be treated as read-only.
type MiddlewareNamespaceLister interface {
	// List lists all Middlewares in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.Middleware, err error)
	// Get retrieves the Middleware from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.Middleware, error)
	MiddlewareNamespaceListerExpansion
}

// middlewareNamespaceLister implements the MiddlewareNamespaceLister
// interface.
type middlewareNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all Middlewares in the indexer for a given namespace.
func (s middlewareNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.Middleware, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.Middleware))
	})
	return ret, err
}

// Get retrieves the Middleware from the indexer for a given namespace and name.
func (s middlewareNamespaceLister) Get(name string) (*v1alpha1.Middleware, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("middleware"), name)
	}
	return obj.(*v1alpha1.Middleware), nil
}
