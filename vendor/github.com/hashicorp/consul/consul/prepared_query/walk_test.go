package prepared_query

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/consul/consul/structs"
	"sort"
)

func TestWalk_ServiceQuery(t *testing.T) {
	var actual []string
	fn := func(path string, v reflect.Value) error {
		actual = append(actual, fmt.Sprintf("%s:%s", path, v.String()))
		return nil
	}

	service := &structs.ServiceQuery{
		Service: "the-service",
		Failover: structs.QueryDatacenterOptions{
			Datacenters: []string{"dc1", "dc2"},
		},
		Near:     "_agent",
		Tags:     []string{"tag1", "tag2", "tag3"},
		NodeMeta: map[string]string{"foo": "bar", "role": "server"},
	}
	if err := walk(service, fn); err != nil {
		t.Fatalf("err: %v", err)
	}

	expected := []string{
		".Failover.Datacenters[0]:dc1",
		".Failover.Datacenters[1]:dc2",
		".Near:_agent",
		".NodeMeta[foo]:bar",
		".NodeMeta[role]:server",
		".Service:the-service",
		".Tags[0]:tag1",
		".Tags[1]:tag2",
		".Tags[2]:tag3",
	}
	sort.Strings(actual)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("bad: %#v", actual)
	}
}

func TestWalk_Visitor_Errors(t *testing.T) {
	fn := func(path string, v reflect.Value) error {
		return fmt.Errorf("bad")
	}

	service := &structs.ServiceQuery{}
	err := walk(service, fn)
	if err == nil || err.Error() != "bad" {
		t.Fatalf("bad: %#v", err)
	}
}
