// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"errors"
	"net"
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

	host := u.Hostname()
	port := u.Port()

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

// PlainAuth returns a PlainAuth structure based on the parsed URI's
// Username and Password fields.
func (uri URI) PlainAuth() *PlainAuth {
	return &PlainAuth{
		Username: uri.Username,
		Password: uri.Password,
	}
}

func (uri URI) String() string {
	authority, err := url.Parse("")
	if err != nil {
		return err.Error()
	}

	authority.Scheme = uri.Scheme

	if uri.Username != defaultURI.Username || uri.Password != defaultURI.Password {
		authority.User = url.User(uri.Username)

		if uri.Password != defaultURI.Password {
			authority.User = url.UserPassword(uri.Username, uri.Password)
		}
	}

	authority.Host = net.JoinHostPort(uri.Host, strconv.Itoa(uri.Port))

	if defaultPort, found := schemePorts[uri.Scheme]; !found || defaultPort != uri.Port {
		authority.Host = net.JoinHostPort(uri.Host, strconv.Itoa(uri.Port))
	} else {
		// JoinHostPort() automatically add brackets to the host if it's
		// an IPv6 address.
		//
		// If not port is specified, JoinHostPort() return an IP address in the
		// form of "[::1]:", so we use TrimSuffix() to remove the extra ":".
		authority.Host = strings.TrimSuffix(net.JoinHostPort(uri.Host, ""), ":")
	}

	if uri.Vhost != defaultURI.Vhost {
		// Make sure net/url does not double escape, e.g.
		// "%2F" does not become "%252F".
		authority.Path = uri.Vhost
		authority.RawPath = url.QueryEscape(uri.Vhost)
	} else {
		authority.Path = "/"
	}

	return authority.String()
}
