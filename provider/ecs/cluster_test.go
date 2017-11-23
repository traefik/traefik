package ecs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClustersSet(t *testing.T) {
	tests := []struct {
		desc     string
		value    string
		expected Clusters
	}{
		{
			desc:     "One value should return Clusters of size 1",
			value:    "cluster",
			expected: Clusters{"cluster"},
		},
		{
			desc:     "Two values separated by comma should return Clusters of size 2",
			value:    "cluster1,cluster2",
			expected: Clusters{"cluster1", "cluster2"},
		},
		{
			desc:     "Two values separated by semicolon should return Clusters of size 2",
			value:    "cluster1;cluster2",
			expected: Clusters{"cluster1", "cluster2"},
		},
		{
			desc:     "Three values separated by comma and semicolon should return Clusters of size 3",
			value:    "cluster1,cluster2;cluster3",
			expected: Clusters{"cluster1", "cluster2", "cluster3"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var clusters Clusters
			err := clusters.Set(test.value)
			assert.Nil(t, err)
			assert.Equal(t, test.expected, clusters)
		})
	}
}

func TestClustersGet(t *testing.T) {
	tests := []struct {
		desc     string
		clusters Clusters
		expected Clusters
	}{
		{
			desc:     "Should return 1 cluster",
			clusters: Clusters{"cluster"},
			expected: Clusters{"cluster"},
		},
		{
			desc:     "Should return 2 clusters",
			clusters: Clusters{"cluster1", "cluster2"},
			expected: Clusters{"cluster1", "cluster2"},
		},
		{
			desc:     "Should return 3 clusters",
			clusters: Clusters{"cluster1", "cluster2", "cluster3"},
			expected: Clusters{"cluster1", "cluster2", "cluster3"},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.clusters.Get()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestClustersString(t *testing.T) {
	tests := []struct {
		desc     string
		clusters Clusters
		expected string
	}{
		{
			desc:     "Should return 1 cluster",
			clusters: Clusters{"cluster"},
			expected: "[cluster]",
		},
		{
			desc:     "Should return 2 clusters",
			clusters: Clusters{"cluster1", "cluster2"},
			expected: "[cluster1 cluster2]",
		},
		{
			desc:     "Should return 3 clusters",
			clusters: Clusters{"cluster1", "cluster2", "cluster3"},
			expected: "[cluster1 cluster2 cluster3]",
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			actual := test.clusters.String()
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestClustersSetValue(t *testing.T) {
	tests := []struct {
		desc     string
		clusters Clusters
		expected Clusters
	}{
		{
			desc:     "Should return Clusters of size 1",
			clusters: Clusters{"cluster"},
			expected: Clusters{"cluster"},
		},
		{
			desc:     "Should return Clusters of size 2",
			clusters: Clusters{"cluster1", "cluster2"},
			expected: Clusters{"cluster1", "cluster2"},
		},
		{
			desc:     "Should return Clusters of size 3",
			clusters: Clusters{"cluster1", "cluster2", "cluster3"},
			expected: Clusters{"cluster1", "cluster2", "cluster3"},
		},
	}
	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()
			var slice Clusters
			slice.SetValue(test.clusters)
			assert.Equal(t, test.expected, slice)
		})
	}
}
