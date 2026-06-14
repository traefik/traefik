//go:build go1.18
// +build go1.18

// Copyright Traefik Labs
// SPDX-License-Identifier: MIT

package traefik

import (
	"strings"
	"testing"

	"github.com/traefik/traefik/v3/pkg/muxer"
	"github.com/traefik/traefik/v3/pkg/rules"
)

// FuzzDomainMatchHostExpression tests host header matching
// with arbitrary attacker-controlled domain and host expression.
//
// This is the pre-auth boundary for every HTTP request processed
// by Traefik. The Host header is fully attacker-controlled.
// A crash here = DoS on the entire proxy infrastructure.
//
// 23 GitHub Security Advisories exist for Traefik — host matching
// has been the source of auth bypass vulnerabilities.
func FuzzDomainMatchHostExpression(f *testing.F) {
	f.Add("example.com", "example.com")
	f.Add("sub.example.com", "*.example.com")
	f.Add("example.com", "*.example.com")
	f.Add("", "")
	f.Add(strings.Repeat("a", 1000), "*")
	f.Add("a.b.c.d.e.f.g.h.i.j", "*.b.c.d.e.f.g.h.i.j")

	f.Fuzz(func(t *testing.T, domain, hostExpr string) {
		if len(domain) > 10000 || len(hostExpr) > 10000 {
			return
		}
		// Must never panic
		_ = muxer.DomainMatchHostExpression(domain, hostExpr)
	})
}

// FuzzIsASCII tests ASCII validation with arbitrary strings.
// Used to validate HTTP header values and URIs before processing.
func FuzzIsASCII(f *testing.F) {
	f.Add("hello")
	f.Add("héllo")
	f.Add("")
	f.Add("\x00\x01\x02")
	f.Add(strings.Repeat("a", 10000))

	f.Fuzz(func(t *testing.T, s string) {
		if len(s) > 100000 {
			return
		}
		_ = muxer.IsASCII(s)
	})
}

// FuzzRuleParser tests Traefik rule expression parsing with
// arbitrary attacker-controlled rule expressions.
//
// Rule expressions define routing, middleware, and access control
// in Traefik. They come from Docker labels, K8s annotations, and
// configuration files — all potentially attacker-influenced.
func FuzzRuleParser(f *testing.F) {
	matchers := []string{
		"Host", "Path", "Method", "Headers",
		"Query", "ClientIP", "PathPrefix",
	}

	parser, err := rules.NewParser(matchers)
	if err != nil {
		f.Fatal(err)
	}

	f.Add("Host(`example.com`)")
	f.Add("Host(`example.com`) && Path(`/api`)")
	f.Add("Host(`example.com`) || Host(`other.com`)")
	f.Add("!Method(`GET`)")
	f.Add("")
	f.Add("(")
	f.Add(")))")

	f.Fuzz(func(t *testing.T, expr string) {
		if len(expr) > 10000 {
			return
		}
		// Parse should never panic on any expression
		_, _ = parser.Parse(expr)
	})
}
