// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

var errURIScheme = errors.New("AMQP scheme must be either 'amqp://' or 'amqps://'")
var errURIWhitespace = errors.New("URI must not contain whitespace")

var schemePorts = map[string]int{
	"amqp":  5672,
	"amqps": 5671,
}

var defaultURI = URI{
	Scheme:   "amqp",
	Host:     "localhost",
	Port:     5672,
	Username: "guest",
	Password: "guest",
	Vhost:    "/",
}

// URI represents a parsed AMQP URI string.
type URI struct {
	Scheme   string
	Host     string
	Port     int
	Username string
	Password string
	Vhost    string
}

// ParseURI attempts to parse the given AMQP URI according to the spec.
// See http://www.rabbitmq.com/uri-spec.html.
//
// Default values for the fields are:
//
//   Scheme: amqp
//   Host: localhost
//   Port: 5672
//   Username: guest
//   Password: guest
//   Vhost: /
//
func ParseURI(uri string) (URI, error) {
	builder := defaultURI

	if strings.Contains(uri, " ") == true {
		return builder, errURIWhitespace
	}

	u, err := url.Parse(uri)
	if err != nil {
		return builder, err
	}

	defaultPort, okScheme := schemePorts[u.Scheme]

	if okScheme {
		builder.Scheme = u.Scheme
	} else {
		return builder, errURIScheme
	}

	host, port := splitHostPort(u.Host)

	if host != "" {
		builder.Host = host
	}

	if port != "" {
		port32, err := strconv.ParseInt(port, 10, 32)
		if err != nil {
			return builder, err
		}
		builder.Port = int(port32)
	} else {
		builder.Port = defaultPort
	}

	if u.User != nil {
		builder.Username = u.User.Username()
		if password, ok := u.User.Password(); ok {
			builder.Password = password
		}
	}

	if u.Path != "" {
		if strings.HasPrefix(u.Path, "/") {
			if u.Host == "" && strings.HasPrefix(u.Path, "///") {
				// net/url doesn't handle local context authorities and leaves that up
				// to the scheme handler.  In our case, we translate amqp:/// into the
				// default host and whatever the vhost should be
				if len(u.Path) > 3 {
					builder.Vhost = u.Path[3:]
				}
			} else if len(u.Path) > 1 {
				builder.Vhost = u.Path[1:]
			}
		} else {
			builder.Vhost = u.Path
		}
	}

	return builder, nil
}

// Splits host:port, host, [ho:st]:port, or [ho:st].  Unlike net.SplitHostPort
// which splits :port, host:port or [host]:port
//
// Handles hosts that have colons that are in brackets like [::1]:http
func splitHostPort(addr string) (host, port string) {
	i := strings.LastIndex(addr, ":")

	if i >= 0 {
		host, port = addr[:i], addr[i+1:]

		if len(port) > 0 && port[len(port)-1] == ']' && addr[0] == '[' {
			// we've split on an inner colon, the port was missing outside of the
			// brackets so use the full addr.  We could assert that host should not
			// contain any colons here
			host, port = addr, ""
		}
	} else {
		host = addr
	}

	return
}

// PlainAuth returns a PlainAuth structure based on the parsed URI's
// Username and Password fields.
func (uri URI) PlainAuth() *PlainAuth {
	return &PlainAuth{
		Username: uri.Username,
		Password: uri.Password,
	}
}

func (uri URI) String() string {
	var authority string

	if uri.Username != defaultURI.Username || uri.Password != defaultURI.Password {
		authority += uri.Username

		if uri.Password != defaultURI.Password {
			authority += ":" + uri.Password
		}

		authority += "@"
	}

	authority += uri.Host

	if defaultPort, found := schemePorts[uri.Scheme]; !found || defaultPort != uri.Port {
		authority += ":" + strconv.FormatInt(int64(uri.Port), 10)
	}

	var vhost string
	if uri.Vhost != defaultURI.Vhost {
		vhost = uri.Vhost
	}

	return fmt.Sprintf("%s://%s/%s", uri.Scheme, authority, url.QueryEscape(vhost))
}
