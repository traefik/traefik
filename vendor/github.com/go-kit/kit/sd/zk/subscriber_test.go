package zk

import (
	"testing"
	"time"
)

func TestSubscriber(t *testing.T) {
	client := newFakeClient()

	s, err := NewSubscriber(client, path, newFactory(""), logger)
	if err != nil {
		t.Fatalf("failed to create new Subscriber: %v", err)
	}
	defer s.Stop()

	if _, err := s.Endpoints(); err != nil {
		t.Fatal(err)
	}
}

func TestBadFactory(t *testing.T) {
	client := newFakeClient()

	s, err := NewSubscriber(client, path, newFactory("kaboom"), logger)
	if err != nil {
		t.Fatalf("failed to create new Subscriber: %v", err)
	}
	defer s.Stop()

	// instance1 came online
	client.AddService(path+"/instance1", "kaboom")

	// instance2 came online
	client.AddService(path+"/instance2", "zookeeper_node_data")

	if err = asyncTest(100*time.Millisecond, 1, s); err != nil {
		t.Error(err)
	}
}

func TestServiceUpdate(t *testing.T) {
	client := newFakeClient()

	s, err := NewSubscriber(client, path, newFactory(""), logger)
	if err != nil {
		t.Fatalf("failed to create new Subscriber: %v", err)
	}
	defer s.Stop()

	endpoints, err := s.Endpoints()
	if err != nil {
		t.Fatal(err)
	}
	if want, have := 0, len(endpoints); want != have {
		t.Errorf("want %d, have %d", want, have)
	}

	// instance1 came online
	client.AddService(path+"/instance1", "zookeeper_node_data1")

	// instance2 came online
	client.AddService(path+"/instance2", "zookeeper_node_data2")

	// we should have 2 instances
	if err = asyncTest(100*time.Millisecond, 2, s); err != nil {
		t.Error(err)
	}

	// TODO(pb): this bit is flaky
	//
	//// watch triggers an error...
	//client.SendErrorOnWatch()
	//
	//// test if error was consumed
	//if err = client.ErrorIsConsumedWithin(100 * time.Millisecond); err != nil {
	//	t.Error(err)
	//}

	// instance3 came online
	client.AddService(path+"/instance3", "zookeeper_node_data3")

	// we should have 3 instances
	if err = asyncTest(100*time.Millisecond, 3, s); err != nil {
		t.Error(err)
	}

	// instance1 goes offline
	client.RemoveService(path + "/instance1")

	// instance2 goes offline
	client.RemoveService(path + "/instance2")

	// we should have 1 instance
	if err = asyncTest(100*time.Millisecond, 1, s); err != nil {
		t.Error(err)
	}
}

func TestBadSubscriberCreate(t *testing.T) {
	client := newFakeClient()
	client.SendErrorOnWatch()
	s, err := NewSubscriber(client, path, newFactory(""), logger)
	if err == nil {
		t.Error("expected error on new Subscriber")
	}
	if s != nil {
		t.Error("expected Subscriber not to be created")
	}
	s, err = NewSubscriber(client, "BadPath", newFactory(""), logger)
	if err == nil {
		t.Error("expected error on new Subscriber")
	}
	if s != nil {
		t.Error("expected Subscriber not to be created")
	}
}
