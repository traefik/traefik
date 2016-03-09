package mocks

import (
	"github.com/stretchr/testify/mock"
	"github.com/hashicorp/consul/api"
	"strings"
)


// ConsulCatalog is a mock of api.Service
type MockConsulCatalog struct {
	mock.Mock
}


func (m *MockConsulCatalog) Service(service, tag string, q *api.QueryOptions) ([]*api.CatalogService, *api.QueryMeta, error) {
	ret := m.Called(service, tag, q)

	var svcs []*api.CatalogService = nil
	var qm *api.QueryMeta = nil

	if ret.Get(0) != nil {
		svcs = ret.Get(0).([]*api.CatalogService)
	}

	if ret.Get(1) != nil {
		qm = ret.Get(1).(*api.QueryMeta)
	}

	r2 := ret.Error(2)

	return svcs, qm, r2
}


// ConsulKV is a mock of api.KV
type MockConsulKV struct {
	KVPairs []*api.KVPair
}

func (m *MockConsulKV) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	key = strings.TrimPrefix(key, "/")
	for _, pair := range m.KVPairs {
		if strings.TrimPrefix(pair.Key, "/") == key {
			return pair, nil, nil
		}
	}
	return nil, nil, nil
}

func (m *MockConsulKV) List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error) {
	prefix = strings.TrimPrefix(prefix, "/")
	pairs := api.KVPairs{}
	for _, pair := range m.KVPairs {
		if strings.Index(strings.TrimPrefix(pair.Key, "/"), prefix) == 0 {
			pairs = append(pairs, pair)
		}
	}
	return pairs, &api.QueryMeta{ LastIndex:q.WaitIndex }, nil
}
