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
	internalinterfaces "github.com/traefik/traefik/v3/pkg/provider/kubernetes/crd/generated/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// IngressRoutes returns a IngressRouteInformer.
	IngressRoutes() IngressRouteInformer
	// IngressRouteTCPs returns a IngressRouteTCPInformer.
	IngressRouteTCPs() IngressRouteTCPInformer
	// IngressRouteUDPs returns a IngressRouteUDPInformer.
	IngressRouteUDPs() IngressRouteUDPInformer
	// Middlewares returns a MiddlewareInformer.
	Middlewares() MiddlewareInformer
	// MiddlewareTCPs returns a MiddlewareTCPInformer.
	MiddlewareTCPs() MiddlewareTCPInformer
	// ServersTransports returns a ServersTransportInformer.
	ServersTransports() ServersTransportInformer
	// ServersTransportTCPs returns a ServersTransportTCPInformer.
	ServersTransportTCPs() ServersTransportTCPInformer
	// TLSOptions returns a TLSOptionInformer.
	TLSOptions() TLSOptionInformer
	// TLSStores returns a TLSStoreInformer.
	TLSStores() TLSStoreInformer
	// TraefikServices returns a TraefikServiceInformer.
	TraefikServices() TraefikServiceInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// IngressRoutes returns a IngressRouteInformer.
func (v *version) IngressRoutes() IngressRouteInformer {
	return &ingressRouteInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// IngressRouteTCPs returns a IngressRouteTCPInformer.
func (v *version) IngressRouteTCPs() IngressRouteTCPInformer {
	return &ingressRouteTCPInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// IngressRouteUDPs returns a IngressRouteUDPInformer.
func (v *version) IngressRouteUDPs() IngressRouteUDPInformer {
	return &ingressRouteUDPInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// Middlewares returns a MiddlewareInformer.
func (v *version) Middlewares() MiddlewareInformer {
	return &middlewareInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// MiddlewareTCPs returns a MiddlewareTCPInformer.
func (v *version) MiddlewareTCPs() MiddlewareTCPInformer {
	return &middlewareTCPInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ServersTransports returns a ServersTransportInformer.
func (v *version) ServersTransports() ServersTransportInformer {
	return &serversTransportInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// ServersTransportTCPs returns a ServersTransportTCPInformer.
func (v *version) ServersTransportTCPs() ServersTransportTCPInformer {
	return &serversTransportTCPInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// TLSOptions returns a TLSOptionInformer.
func (v *version) TLSOptions() TLSOptionInformer {
	return &tLSOptionInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// TLSStores returns a TLSStoreInformer.
func (v *version) TLSStores() TLSStoreInformer {
	return &tLSStoreInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// TraefikServices returns a TraefikServiceInformer.
func (v *version) TraefikServices() TraefikServiceInformer {
	return &traefikServiceInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
