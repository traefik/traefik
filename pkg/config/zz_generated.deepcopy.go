// +build !ignore_autogenerated

/*
The MIT License (MIT)

Copyright (c) 2016-2019 Containous SAS

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package config

import (
	tls "github.com/containous/traefik/pkg/tls"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AddPrefix) DeepCopyInto(out *AddPrefix) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AddPrefix.
func (in *AddPrefix) DeepCopy() *AddPrefix {
	if in == nil {
		return nil
	}
	out := new(AddPrefix)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Auth) DeepCopyInto(out *Auth) {
	*out = *in
	if in.Basic != nil {
		in, out := &in.Basic, &out.Basic
		*out = new(BasicAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.Digest != nil {
		in, out := &in.Digest, &out.Digest
		*out = new(DigestAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.Forward != nil {
		in, out := &in.Forward, &out.Forward
		*out = new(ForwardAuth)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Auth.
func (in *Auth) DeepCopy() *Auth {
	if in == nil {
		return nil
	}
	out := new(Auth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BasicAuth) DeepCopyInto(out *BasicAuth) {
	*out = *in
	if in.Users != nil {
		in, out := &in.Users, &out.Users
		*out = make(Users, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BasicAuth.
func (in *BasicAuth) DeepCopy() *BasicAuth {
	if in == nil {
		return nil
	}
	out := new(BasicAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Buffering) DeepCopyInto(out *Buffering) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Buffering.
func (in *Buffering) DeepCopy() *Buffering {
	if in == nil {
		return nil
	}
	out := new(Buffering)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Chain) DeepCopyInto(out *Chain) {
	*out = *in
	if in.Middlewares != nil {
		in, out := &in.Middlewares, &out.Middlewares
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Chain.
func (in *Chain) DeepCopy() *Chain {
	if in == nil {
		return nil
	}
	out := new(Chain)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CircuitBreaker) DeepCopyInto(out *CircuitBreaker) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CircuitBreaker.
func (in *CircuitBreaker) DeepCopy() *CircuitBreaker {
	if in == nil {
		return nil
	}
	out := new(CircuitBreaker)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ClientTLS) DeepCopyInto(out *ClientTLS) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ClientTLS.
func (in *ClientTLS) DeepCopy() *ClientTLS {
	if in == nil {
		return nil
	}
	out := new(ClientTLS)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Compress) DeepCopyInto(out *Compress) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Compress.
func (in *Compress) DeepCopy() *Compress {
	if in == nil {
		return nil
	}
	out := new(Compress)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Configuration) DeepCopyInto(out *Configuration) {
	*out = *in
	if in.HTTP != nil {
		in, out := &in.HTTP, &out.HTTP
		*out = new(HTTPConfiguration)
		(*in).DeepCopyInto(*out)
	}
	if in.TCP != nil {
		in, out := &in.TCP, &out.TCP
		*out = new(TCPConfiguration)
		(*in).DeepCopyInto(*out)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(TLSConfiguration)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Configuration.
func (in *Configuration) DeepCopy() *Configuration {
	if in == nil {
		return nil
	}
	out := new(Configuration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Configurations) DeepCopyInto(out *Configurations) {
	{
		in := &in
		*out = make(Configurations, len(*in))
		for key, val := range *in {
			var outVal *Configuration
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Configuration)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Configurations.
func (in Configurations) DeepCopy() Configurations {
	if in == nil {
		return nil
	}
	out := new(Configurations)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DigestAuth) DeepCopyInto(out *DigestAuth) {
	*out = *in
	if in.Users != nil {
		in, out := &in.Users, &out.Users
		*out = make(Users, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DigestAuth.
func (in *DigestAuth) DeepCopy() *DigestAuth {
	if in == nil {
		return nil
	}
	out := new(DigestAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ErrorPage) DeepCopyInto(out *ErrorPage) {
	*out = *in
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ErrorPage.
func (in *ErrorPage) DeepCopy() *ErrorPage {
	if in == nil {
		return nil
	}
	out := new(ErrorPage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ForwardAuth) DeepCopyInto(out *ForwardAuth) {
	*out = *in
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(ClientTLS)
		**out = **in
	}
	if in.AuthResponseHeaders != nil {
		in, out := &in.AuthResponseHeaders, &out.AuthResponseHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ForwardAuth.
func (in *ForwardAuth) DeepCopy() *ForwardAuth {
	if in == nil {
		return nil
	}
	out := new(ForwardAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HTTPConfiguration) DeepCopyInto(out *HTTPConfiguration) {
	*out = *in
	if in.Routers != nil {
		in, out := &in.Routers, &out.Routers
		*out = make(map[string]*Router, len(*in))
		for key, val := range *in {
			var outVal *Router
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Router)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.Middlewares != nil {
		in, out := &in.Middlewares, &out.Middlewares
		*out = make(map[string]*Middleware, len(*in))
		for key, val := range *in {
			var outVal *Middleware
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Middleware)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make(map[string]*Service, len(*in))
		for key, val := range *in {
			var outVal *Service
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Service)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HTTPConfiguration.
func (in *HTTPConfiguration) DeepCopy() *HTTPConfiguration {
	if in == nil {
		return nil
	}
	out := new(HTTPConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Headers) DeepCopyInto(out *Headers) {
	*out = *in
	if in.CustomRequestHeaders != nil {
		in, out := &in.CustomRequestHeaders, &out.CustomRequestHeaders
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.CustomResponseHeaders != nil {
		in, out := &in.CustomResponseHeaders, &out.CustomResponseHeaders
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.AccessControlAllowHeaders != nil {
		in, out := &in.AccessControlAllowHeaders, &out.AccessControlAllowHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AccessControlAllowMethods != nil {
		in, out := &in.AccessControlAllowMethods, &out.AccessControlAllowMethods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AccessControlExposeHeaders != nil {
		in, out := &in.AccessControlExposeHeaders, &out.AccessControlExposeHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowedHosts != nil {
		in, out := &in.AllowedHosts, &out.AllowedHosts
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.HostsProxyHeaders != nil {
		in, out := &in.HostsProxyHeaders, &out.HostsProxyHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.SSLProxyHeaders != nil {
		in, out := &in.SSLProxyHeaders, &out.SSLProxyHeaders
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Headers.
func (in *Headers) DeepCopy() *Headers {
	if in == nil {
		return nil
	}
	out := new(Headers)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HealthCheck) DeepCopyInto(out *HealthCheck) {
	*out = *in
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HealthCheck.
func (in *HealthCheck) DeepCopy() *HealthCheck {
	if in == nil {
		return nil
	}
	out := new(HealthCheck)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPStrategy) DeepCopyInto(out *IPStrategy) {
	*out = *in
	if in.ExcludedIPs != nil {
		in, out := &in.ExcludedIPs, &out.ExcludedIPs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPStrategy.
func (in *IPStrategy) DeepCopy() *IPStrategy {
	if in == nil {
		return nil
	}
	out := new(IPStrategy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IPWhiteList) DeepCopyInto(out *IPWhiteList) {
	*out = *in
	if in.SourceRange != nil {
		in, out := &in.SourceRange, &out.SourceRange
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.IPStrategy != nil {
		in, out := &in.IPStrategy, &out.IPStrategy
		*out = new(IPStrategy)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IPWhiteList.
func (in *IPWhiteList) DeepCopy() *IPWhiteList {
	if in == nil {
		return nil
	}
	out := new(IPWhiteList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LoadBalancerService) DeepCopyInto(out *LoadBalancerService) {
	*out = *in
	if in.Stickiness != nil {
		in, out := &in.Stickiness, &out.Stickiness
		*out = new(Stickiness)
		**out = **in
	}
	if in.Servers != nil {
		in, out := &in.Servers, &out.Servers
		*out = make([]Server, len(*in))
		copy(*out, *in)
	}
	if in.HealthCheck != nil {
		in, out := &in.HealthCheck, &out.HealthCheck
		*out = new(HealthCheck)
		(*in).DeepCopyInto(*out)
	}
	if in.ResponseForwarding != nil {
		in, out := &in.ResponseForwarding, &out.ResponseForwarding
		*out = new(ResponseForwarding)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LoadBalancerService.
func (in *LoadBalancerService) DeepCopy() *LoadBalancerService {
	if in == nil {
		return nil
	}
	out := new(LoadBalancerService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MaxConn) DeepCopyInto(out *MaxConn) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MaxConn.
func (in *MaxConn) DeepCopy() *MaxConn {
	if in == nil {
		return nil
	}
	out := new(MaxConn)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Message) DeepCopyInto(out *Message) {
	*out = *in
	if in.Configuration != nil {
		in, out := &in.Configuration, &out.Configuration
		*out = new(Configuration)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Message.
func (in *Message) DeepCopy() *Message {
	if in == nil {
		return nil
	}
	out := new(Message)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Middleware) DeepCopyInto(out *Middleware) {
	*out = *in
	if in.AddPrefix != nil {
		in, out := &in.AddPrefix, &out.AddPrefix
		*out = new(AddPrefix)
		**out = **in
	}
	if in.StripPrefix != nil {
		in, out := &in.StripPrefix, &out.StripPrefix
		*out = new(StripPrefix)
		(*in).DeepCopyInto(*out)
	}
	if in.StripPrefixRegex != nil {
		in, out := &in.StripPrefixRegex, &out.StripPrefixRegex
		*out = new(StripPrefixRegex)
		(*in).DeepCopyInto(*out)
	}
	if in.ReplacePath != nil {
		in, out := &in.ReplacePath, &out.ReplacePath
		*out = new(ReplacePath)
		**out = **in
	}
	if in.ReplacePathRegex != nil {
		in, out := &in.ReplacePathRegex, &out.ReplacePathRegex
		*out = new(ReplacePathRegex)
		**out = **in
	}
	if in.Chain != nil {
		in, out := &in.Chain, &out.Chain
		*out = new(Chain)
		(*in).DeepCopyInto(*out)
	}
	if in.IPWhiteList != nil {
		in, out := &in.IPWhiteList, &out.IPWhiteList
		*out = new(IPWhiteList)
		(*in).DeepCopyInto(*out)
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = new(Headers)
		(*in).DeepCopyInto(*out)
	}
	if in.Errors != nil {
		in, out := &in.Errors, &out.Errors
		*out = new(ErrorPage)
		(*in).DeepCopyInto(*out)
	}
	if in.RateLimit != nil {
		in, out := &in.RateLimit, &out.RateLimit
		*out = new(RateLimit)
		(*in).DeepCopyInto(*out)
	}
	if in.RedirectRegex != nil {
		in, out := &in.RedirectRegex, &out.RedirectRegex
		*out = new(RedirectRegex)
		**out = **in
	}
	if in.RedirectScheme != nil {
		in, out := &in.RedirectScheme, &out.RedirectScheme
		*out = new(RedirectScheme)
		**out = **in
	}
	if in.BasicAuth != nil {
		in, out := &in.BasicAuth, &out.BasicAuth
		*out = new(BasicAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.DigestAuth != nil {
		in, out := &in.DigestAuth, &out.DigestAuth
		*out = new(DigestAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.ForwardAuth != nil {
		in, out := &in.ForwardAuth, &out.ForwardAuth
		*out = new(ForwardAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.MaxConn != nil {
		in, out := &in.MaxConn, &out.MaxConn
		*out = new(MaxConn)
		**out = **in
	}
	if in.Buffering != nil {
		in, out := &in.Buffering, &out.Buffering
		*out = new(Buffering)
		**out = **in
	}
	if in.CircuitBreaker != nil {
		in, out := &in.CircuitBreaker, &out.CircuitBreaker
		*out = new(CircuitBreaker)
		**out = **in
	}
	if in.Compress != nil {
		in, out := &in.Compress, &out.Compress
		*out = new(Compress)
		**out = **in
	}
	if in.PassTLSClientCert != nil {
		in, out := &in.PassTLSClientCert, &out.PassTLSClientCert
		*out = new(PassTLSClientCert)
		(*in).DeepCopyInto(*out)
	}
	if in.Retry != nil {
		in, out := &in.Retry, &out.Retry
		*out = new(Retry)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Middleware.
func (in *Middleware) DeepCopy() *Middleware {
	if in == nil {
		return nil
	}
	out := new(Middleware)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PassTLSClientCert) DeepCopyInto(out *PassTLSClientCert) {
	*out = *in
	if in.Info != nil {
		in, out := &in.Info, &out.Info
		*out = new(TLSClientCertificateInfo)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PassTLSClientCert.
func (in *PassTLSClientCert) DeepCopy() *PassTLSClientCert {
	if in == nil {
		return nil
	}
	out := new(PassTLSClientCert)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rate) DeepCopyInto(out *Rate) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rate.
func (in *Rate) DeepCopy() *Rate {
	if in == nil {
		return nil
	}
	out := new(Rate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RateLimit) DeepCopyInto(out *RateLimit) {
	*out = *in
	if in.RateSet != nil {
		in, out := &in.RateSet, &out.RateSet
		*out = make(map[string]*Rate, len(*in))
		for key, val := range *in {
			var outVal *Rate
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(Rate)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RateLimit.
func (in *RateLimit) DeepCopy() *RateLimit {
	if in == nil {
		return nil
	}
	out := new(RateLimit)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedirectRegex) DeepCopyInto(out *RedirectRegex) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedirectRegex.
func (in *RedirectRegex) DeepCopy() *RedirectRegex {
	if in == nil {
		return nil
	}
	out := new(RedirectRegex)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedirectScheme) DeepCopyInto(out *RedirectScheme) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedirectScheme.
func (in *RedirectScheme) DeepCopy() *RedirectScheme {
	if in == nil {
		return nil
	}
	out := new(RedirectScheme)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReplacePath) DeepCopyInto(out *ReplacePath) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReplacePath.
func (in *ReplacePath) DeepCopy() *ReplacePath {
	if in == nil {
		return nil
	}
	out := new(ReplacePath)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ReplacePathRegex) DeepCopyInto(out *ReplacePathRegex) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ReplacePathRegex.
func (in *ReplacePathRegex) DeepCopy() *ReplacePathRegex {
	if in == nil {
		return nil
	}
	out := new(ReplacePathRegex)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ResponseForwarding) DeepCopyInto(out *ResponseForwarding) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ResponseForwarding.
func (in *ResponseForwarding) DeepCopy() *ResponseForwarding {
	if in == nil {
		return nil
	}
	out := new(ResponseForwarding)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Retry) DeepCopyInto(out *Retry) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Retry.
func (in *Retry) DeepCopy() *Retry {
	if in == nil {
		return nil
	}
	out := new(Retry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Router) DeepCopyInto(out *Router) {
	*out = *in
	if in.EntryPoints != nil {
		in, out := &in.EntryPoints, &out.EntryPoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Middlewares != nil {
		in, out := &in.Middlewares, &out.Middlewares
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(RouterTLSConfig)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Router.
func (in *Router) DeepCopy() *Router {
	if in == nil {
		return nil
	}
	out := new(Router)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouterTCPTLSConfig) DeepCopyInto(out *RouterTCPTLSConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouterTCPTLSConfig.
func (in *RouterTCPTLSConfig) DeepCopy() *RouterTCPTLSConfig {
	if in == nil {
		return nil
	}
	out := new(RouterTCPTLSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouterTLSConfig) DeepCopyInto(out *RouterTLSConfig) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouterTLSConfig.
func (in *RouterTLSConfig) DeepCopy() *RouterTLSConfig {
	if in == nil {
		return nil
	}
	out := new(RouterTLSConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Server) DeepCopyInto(out *Server) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Server.
func (in *Server) DeepCopy() *Server {
	if in == nil {
		return nil
	}
	out := new(Server)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	if in.LoadBalancer != nil {
		in, out := &in.LoadBalancer, &out.LoadBalancer
		*out = new(LoadBalancerService)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Stickiness) DeepCopyInto(out *Stickiness) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Stickiness.
func (in *Stickiness) DeepCopy() *Stickiness {
	if in == nil {
		return nil
	}
	out := new(Stickiness)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StripPrefix) DeepCopyInto(out *StripPrefix) {
	*out = *in
	if in.Prefixes != nil {
		in, out := &in.Prefixes, &out.Prefixes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StripPrefix.
func (in *StripPrefix) DeepCopy() *StripPrefix {
	if in == nil {
		return nil
	}
	out := new(StripPrefix)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StripPrefixRegex) DeepCopyInto(out *StripPrefixRegex) {
	*out = *in
	if in.Regex != nil {
		in, out := &in.Regex, &out.Regex
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StripPrefixRegex.
func (in *StripPrefixRegex) DeepCopy() *StripPrefixRegex {
	if in == nil {
		return nil
	}
	out := new(StripPrefixRegex)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCPConfiguration) DeepCopyInto(out *TCPConfiguration) {
	*out = *in
	if in.Routers != nil {
		in, out := &in.Routers, &out.Routers
		*out = make(map[string]*TCPRouter, len(*in))
		for key, val := range *in {
			var outVal *TCPRouter
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(TCPRouter)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	if in.Services != nil {
		in, out := &in.Services, &out.Services
		*out = make(map[string]*TCPService, len(*in))
		for key, val := range *in {
			var outVal *TCPService
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(TCPService)
				(*in).DeepCopyInto(*out)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCPConfiguration.
func (in *TCPConfiguration) DeepCopy() *TCPConfiguration {
	if in == nil {
		return nil
	}
	out := new(TCPConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCPLoadBalancerService) DeepCopyInto(out *TCPLoadBalancerService) {
	*out = *in
	if in.Servers != nil {
		in, out := &in.Servers, &out.Servers
		*out = make([]TCPServer, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCPLoadBalancerService.
func (in *TCPLoadBalancerService) DeepCopy() *TCPLoadBalancerService {
	if in == nil {
		return nil
	}
	out := new(TCPLoadBalancerService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCPRouter) DeepCopyInto(out *TCPRouter) {
	*out = *in
	if in.EntryPoints != nil {
		in, out := &in.EntryPoints, &out.EntryPoints
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.TLS != nil {
		in, out := &in.TLS, &out.TLS
		*out = new(RouterTCPTLSConfig)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCPRouter.
func (in *TCPRouter) DeepCopy() *TCPRouter {
	if in == nil {
		return nil
	}
	out := new(TCPRouter)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCPServer) DeepCopyInto(out *TCPServer) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCPServer.
func (in *TCPServer) DeepCopy() *TCPServer {
	if in == nil {
		return nil
	}
	out := new(TCPServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TCPService) DeepCopyInto(out *TCPService) {
	*out = *in
	if in.LoadBalancer != nil {
		in, out := &in.LoadBalancer, &out.LoadBalancer
		*out = new(TCPLoadBalancerService)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TCPService.
func (in *TCPService) DeepCopy() *TCPService {
	if in == nil {
		return nil
	}
	out := new(TCPService)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSCLientCertificateDNInfo) DeepCopyInto(out *TLSCLientCertificateDNInfo) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSCLientCertificateDNInfo.
func (in *TLSCLientCertificateDNInfo) DeepCopy() *TLSCLientCertificateDNInfo {
	if in == nil {
		return nil
	}
	out := new(TLSCLientCertificateDNInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSClientCertificateInfo) DeepCopyInto(out *TLSClientCertificateInfo) {
	*out = *in
	if in.Subject != nil {
		in, out := &in.Subject, &out.Subject
		*out = new(TLSCLientCertificateDNInfo)
		**out = **in
	}
	if in.Issuer != nil {
		in, out := &in.Issuer, &out.Issuer
		*out = new(TLSCLientCertificateDNInfo)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSClientCertificateInfo.
func (in *TLSClientCertificateInfo) DeepCopy() *TLSClientCertificateInfo {
	if in == nil {
		return nil
	}
	out := new(TLSClientCertificateInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TLSConfiguration) DeepCopyInto(out *TLSConfiguration) {
	*out = *in
	if in.Certificates != nil {
		in, out := &in.Certificates, &out.Certificates
		*out = make([]*tls.CertAndStores, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(tls.CertAndStores)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Options != nil {
		in, out := &in.Options, &out.Options
		*out = make(map[string]tls.Options, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	if in.Stores != nil {
		in, out := &in.Stores, &out.Stores
		*out = make(map[string]tls.Store, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TLSConfiguration.
func (in *TLSConfiguration) DeepCopy() *TLSConfiguration {
	if in == nil {
		return nil
	}
	out := new(TLSConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Users) DeepCopyInto(out *Users) {
	{
		in := &in
		*out = make(Users, len(*in))
		copy(*out, *in)
		return
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Users.
func (in Users) DeepCopy() Users {
	if in == nil {
		return nil
	}
	out := new(Users)
	in.DeepCopyInto(out)
	return *out
}
