package ecs

import (
	"reflect"
	"testing"
)

func TestClustersSet(t *testing.T) {
	checkMap := map[string]Clusters{
		"cluster":                    {"cluster"},
		"cluster1,cluster2":          {"cluster1", "cluster2"},
		"cluster1;cluster2":          {"cluster1", "cluster2"},
		"cluster1,cluster2;cluster3": {"cluster1", "cluster2", "cluster3"},
	}
	for str, check := range checkMap {
		var clusters Clusters
		if err := clusters.Set(str); err != nil {
			t.Fatalf("Error :%s", err)
		}
		if !reflect.DeepEqual(clusters, check) {
			t.Fatalf("Expected:%s\ngot:%s", check, clusters)
		}
	}
}

func TestClustersGet(t *testing.T) {
	slices := []Clusters{
		{"cluster"},
		{"cluster1", "cluster2"},
		{"cluster1", "cluster2", "cluster3"},
	}
	check := []Clusters{
		{"cluster"},
		{"cluster1", "cluster2"},
		{"cluster1", "cluster2", "cluster3"},
	}
	for i, slice := range slices {
		if !reflect.DeepEqual(slice.Get(), check[i]) {
			t.Fatalf("Expected:%s\ngot:%s", check[i], slice)
		}
	}
}

func TestClustersString(t *testing.T) {
	slices := []Clusters{
		{"cluster"},
		{"cluster1", "cluster2"},
		{"cluster1", "cluster2", "cluster3"},
	}
	check := []string{
		"[cluster]",
		"[cluster1 cluster2]",
		"[cluster1 cluster2 cluster3]",
	}
	for i, slice := range slices {
		if !reflect.DeepEqual(slice.String(), check[i]) {
			t.Fatalf("Expected:%s\ngot:%s", check[i], slice)
		}
	}
}

func TestClustersSetValue(t *testing.T) {
	check := []Clusters{
		{"cluster"},
		{"cluster1", "cluster2"},
		{"cluster1", "cluster2", "cluster3"},
	}
	slices := []Clusters{
		{"cluster"},
		{"cluster1", "cluster2"},
		{"cluster1", "cluster2", "cluster3"},
	}
	for i, s := range slices {
		var slice Clusters
		slice.SetValue(s)
		if !reflect.DeepEqual(slice, check[i]) {
			t.Fatalf("Expected:%s\ngot:%s", check[i], slice)
		}
	}
}
