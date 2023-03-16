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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	versioned "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/clientset/versioned"
	internalinterfaces "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/generated/listers/traefikcontainous/v1alpha1"
	traefikcontainousv1alpha1 "github.com/traefik/traefik/v2/pkg/provider/kubernetes/crd/traefikcontainous/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ServersTransportInformer provides access to a shared informer and lister for
// ServersTransports.
type ServersTransportInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ServersTransportLister
}

type serversTransportInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewServersTransportInformer constructs a new informer for ServersTransport type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewServersTransportInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredServersTransportInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredServersTransportInformer constructs a new informer for ServersTransport type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredServersTransportInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TraefikContainousV1alpha1().ServersTransports(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.TraefikContainousV1alpha1().ServersTransports(namespace).Watch(context.TODO(), options)
			},
		},
		&traefikcontainousv1alpha1.ServersTransport{},
		resyncPeriod,
		indexers,
	)
}

func (f *serversTransportInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredServersTransportInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *serversTransportInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&traefikcontainousv1alpha1.ServersTransport{}, f.defaultInformer)
}

func (f *serversTransportInformer) Lister() v1alpha1.ServersTransportLister {
	return v1alpha1.NewServersTransportLister(f.Informer().GetIndexer())
}
