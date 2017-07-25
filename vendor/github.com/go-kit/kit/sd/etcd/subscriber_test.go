package etcd

import (
	"errors"
	"io"
	"testing"

	stdetcd "github.com/coreos/etcd/client"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

var (
	node = &stdetcd.Node{
		Key: "/foo",
		Nodes: []*stdetcd.Node{
			{Key: "/foo/1", Value: "1:1"},
			{Key: "/foo/2", Value: "1:2"},
		},
	}
	fakeResponse = &stdetcd.Response{
		Node: node,
	}
)

func TestSubscriber(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return endpoint.Nop, nil, nil
	}

	client := &fakeClient{
		responses: map[string]*stdetcd.Response{"/foo": fakeResponse},
	}

	s, err := NewSubscriber(client, "/foo", factory, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	if _, err := s.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadFactory(t *testing.T) {
	factory := func(string) (endpoint.Endpoint, io.Closer, error) {
		return nil, nil, errors.New("kaboom")
	}

	client := &fakeClient{
		responses: map[string]*stdetcd.Response{"/foo": fakeResponse},
	}

	s, err := NewSubscriber(client, "/foo", factory, log.NewNopLogger())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()

	endpoints, err := s.Endpoints()
	if err != nil {
		t.Fatal(err)
	}

	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}
}

type fakeClient struct {
	responses map[string]*stdetcd.Response
}

func (c *fakeClient) GetEntries(prefix string) ([]string, error) {
	response, ok := c.responses[prefix]
	if !ok {
		return nil, errors.New("key not exist")
	}

	entries := make([]string, len(response.Node.Nodes))
	for i, node := range response.Node.Nodes {
		entries[i] = node.Value
	}
	return entries, nil
}

func (c *fakeClient) WatchPrefix(prefix string, ch chan struct{}) {}

func (c *fakeClient) Register(Service) error {
	return nil
}
func (c *fakeClient) Deregister(Service) error {
	return nil
}
