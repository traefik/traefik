package dnspod

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestRecords_recordPath(t *testing.T) {
	var pathTest = []struct {
		actionInput string
		expected    string
	}{
		{"List", "Record.List"},
		{"", "Record.List"},
	}

	for _, pt := range pathTest {
		actual := recordAction(pt.actionInput)
		if actual != pt.expected {
			t.Errorf("recordPath(%+v): expected %s, actual %s", pt.actionInput, pt.expected, actual)
		}
	}
}

func TestDomainsService_ListRecords_all(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Record.List", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{
			"status": {"code":"1","message":""},
			"records":[
				{"id":"44146112", "name":"yizerowwwww"},
				{"id":"44146112", "name":"yizerowwwww"}
			]}`)
	})

	records, _, err := client.Domains.ListRecords("example.com", "")

	if err != nil {
		t.Errorf("Domains.ListRecords returned error: %v", err)
	}

	want := []Record{{ID: "44146112", Name: "yizerowwwww"}, {ID: "44146112", Name: "yizerowwwww"}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("Domains.ListRecords returned %+v, want %+v", records, want)
	}
}

func TestDomainsService_ListRecords_subdomain(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Record.List", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{
			"status": {"code":"1","message":""},
			"records":[
				{"id":"44146112", "name":"yizerowwwww"},
				{"id":"44146112", "name":"yizerowwwww"}
			]}`)
	})

	records, _, err := client.Domains.ListRecords("example.com", "@")

	if err != nil {
		t.Errorf("Domains.ListRecords returned error: %v", err)
	}

	want := []Record{{ID: "44146112", Name: "yizerowwwww"}, {ID: "44146112", Name: "yizerowwwww"}}
	if !reflect.DeepEqual(records, want) {
		t.Fatalf("Domains.ListRecords returned %+v, want %+v", records, want)
	}
}

func TestDomainsService_CreateRecord(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Record.Create", func(w http.ResponseWriter, r *http.Request) {
		// want := make(map[string]interface{})
		// want["record"] = map[string]interface{}{"name": "foo", "content": "192.168.0.10", "record_type": "A"}

		testMethod(t, r, "POST")
		// testRequestJSON(t, r, want)

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"status": {"code":"1","message":""},"record":{"id":"26954449", "name":"@", "status":"enable"}}`)
	})

	recordValues := Record{Name: "@", Status: "enable"}
	record, _, err := client.Domains.CreateRecord("44146112", recordValues)

	if err != nil {
		t.Errorf("Domains.CreateRecord returned error: %v", err)
	}

	want := Record{ID: "26954449", Name: "@", Status: "enable"}
	if !reflect.DeepEqual(record, want) {
		t.Fatalf("Domains.CreateRecord returned %+v, want %+v", record, want)
	}
}

func TestDomainsService_GetRecord(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Record.Info", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprintf(w, `{"status": {"code":"1","message":""},"record":{"id":"26954449", "name":"@", "status":"enable"}}`)
	})

	record, _, err := client.Domains.GetRecord("44146112", "26954449")

	if err != nil {
		t.Errorf("Domains.GetRecord returned error: %v", err)
	}

	want := Record{ID: "26954449", Name: "@", Status: "enable"}
	if !reflect.DeepEqual(record, want) {
		t.Fatalf("Domains.GetRecord returned %+v, want %+v", record, want)
	}
}

func TestDomainsService_UpdateRecord(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Record.Modify", func(w http.ResponseWriter, r *http.Request) {
		// want := make(map[string]interface{})
		// want["record"] = map[string]interface{}{"content": "192.168.0.10", "name": "bar"}

		testMethod(t, r, "POST")
		// testRequestJSON(t, r, want)

		fmt.Fprint(w, `{"status": {"code":"1","message":""},"record":{"id":"26954449", "name":"@", "status":"enable"}}`)
	})

	recordValues := Record{ID: "26954449", Name: "@", Status: "enable"}
	record, _, err := client.Domains.UpdateRecord("44146112", "26954449", recordValues)

	if err != nil {
		t.Errorf("Domains.UpdateRecord returned error: %v", err)
	}

	want := Record{ID: "26954449", Name: "@", Status: "enable"}
	if !reflect.DeepEqual(record, want) {
		t.Fatalf("Domains.UpdateRecord returned %+v, want %+v", record, want)
	}
}

func TestDomainsService_DeleteRecord(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Record.Remove", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"status": {"code":"1","message":""}}`)
	})

	_, err := client.Domains.DeleteRecord("44146112", "26954449")

	if err != nil {
		t.Errorf("Domains.DeleteRecord returned error: %v", err)
	}
}

func TestDomainsService_DeleteRecord_failed(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/Record.Remove", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")

		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{"message":"InvalID request"}`)
	})

	_, err := client.Domains.DeleteRecord("44146112", "26954449")
	if err == nil {
		t.Errorf("Domains.DeleteRecord expected error to be returned")
	}

	if match := "400 InvalID request"; !strings.Contains(err.Error(), match) {
		t.Errorf("Records.Delete returned %+v, should match %+v", err, match)
	}
}
