package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestZoneRecordPath(t *testing.T) {
	if want, got := "/1010/zones/example.com/records", zoneRecordPath("1010", "example.com", 0); want != got {
		t.Errorf("zoneRecordPath(%v) = %v, want %v", 0, got, want)
	}

	if want, got := "/1010/zones/example.com/records/1", zoneRecordPath("1010", "example.com", 1); want != got {
		t.Errorf("zoneRecordPath(%v) = %v, want %v", 1, got, want)
	}
}

func TestZonesService_ListRecords(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listZoneRecords/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordsResponse, err := client.Zones.ListRecords("1010", "example.com", nil)
	if err != nil {
		t.Fatalf("Zones.ListRecords() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 5}), recordsResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Zones.ListRecords() pagination expected to be %v, got %v", want, got)
	}

	records := recordsResponse.Data
	if want, got := 5, len(records); want != got {
		t.Errorf("Zones.ListRecords() expected to return %v contacts, got %v", want, got)
	}

	if want, got := 1, records[0].ID; want != got {
		t.Fatalf("Zones.ListRecords() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "", records[0].Name; want != got {
		t.Fatalf("Zones.ListRecords() returned Name expected to be `%v`, got `%v`", want, got)
	}
	if !reflect.DeepEqual([]string{"global"}, records[0].Regions) {
		t.Fatalf("Zones.ListRecords() returned %+v, want %+v", records[0].Regions, []string{"global"})
	}
}

func TestZonesService_ListRecords_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listZoneRecords/success.http")

		testQuery(t, r, url.Values{
			"page":        []string{"2"},
			"per_page":    []string{"20"},
			"sort":        []string{"name,expiration:desc"},
			"name":        []string{"example"},
			"name_like":   []string{"www"},
			"record_type": []string{"A"},
		})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Zones.ListRecords("1010", "example.com", &ZoneRecordListOptions{"example", "www", "A", ListOptions{Page: 2, PerPage: 20, Sort: "name,expiration:desc"}})
	if err != nil {
		t.Fatalf("Zones.ListRecords() returned error: %v", err)
	}
}

func TestZonesService_CreateRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createZoneRecord/created.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"name": "foo", "content": "mxa.example.com", "type": "MX"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountID := "1010"
	recordValues := ZoneRecord{Name: "foo", Content: "mxa.example.com", Type: "MX"}

	recordResponse, err := client.Zones.CreateRecord(accountID, "example.com", recordValues)
	if err != nil {
		t.Fatalf("Zones.CreateRecord() returned error: %v", err)
	}

	record := recordResponse.Data
	if want, got := 1, record.ID; want != got {
		t.Fatalf("Zones.CreateRecord() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "www", record.Name; want != got {
		t.Fatalf("Zones.CreateRecord() returned Name expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "A", record.Type; want != got {
		t.Fatalf("Zones.CreateRecord() returned Type expected to be `%v`, got `%v`", want, got)
	}
	if !reflect.DeepEqual([]string{"global"}, record.Regions) {
		t.Fatalf("Zones.ListRecords() returned %+v, want %+v", record.Regions, []string{"global"})
	}
}

func TestZonesService_CreateRecord_BlankName(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createZoneRecord/created-apex.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"name": "", "content": "127.0.0.1", "type": "A"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordValues := ZoneRecord{Name: "", Content: "127.0.0.1", Type: "A"}

	recordResponse, err := client.Zones.CreateRecord("1010", "example.com", recordValues)
	if err != nil {
		t.Fatalf("Zones.CreateRecord() returned error: %v", err)
	}

	record := recordResponse.Data
	if want, got := "", record.Name; want != got {
		t.Fatalf("Zones.CreateRecord() returned Name expected to be `%v`, got `%v`", want, got)
	}
	if !reflect.DeepEqual([]string{"global"}, record.Regions) {
		t.Fatalf("Zones.ListRecords() returned %+v, want %+v", record.Regions, []string{"global"})
	}
}

func TestZonesService_CreateRecord_Regions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	var recordValues ZoneRecord

	mux.HandleFunc("/v2/1/zones/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createZoneRecord/created.http")

		want := map[string]interface{}{"name": "foo"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordValues = ZoneRecord{Name: "foo", Regions: []string{}}
	if _, err := client.Zones.CreateRecord("1", "example.com", recordValues); err != nil {
		t.Fatalf("Zones.CreateRecord() returned error: %v", err)
	}

	mux.HandleFunc("/v2/2/zones/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createZoneRecord/created.http")

		want := map[string]interface{}{"name": "foo", "regions": []interface{}{"global"}}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordValues = ZoneRecord{Name: "foo", Regions: []string{"global"}}
	if _, err := client.Zones.CreateRecord("2", "example.com", recordValues); err != nil {
		t.Fatalf("Zones.CreateRecord() returned error: %v", err)
	}

	mux.HandleFunc("/v2/3/zones/example.com/records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createZoneRecord/created.http")

		want := map[string]interface{}{"name": "foo", "regions": []interface{}{"global"}}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordValues = ZoneRecord{Name: "foo", Regions: []string{"global"}}
	if _, err := client.Zones.CreateRecord("2", "example.com", recordValues); err != nil {
		t.Fatalf("Zones.CreateRecord() returned error: %v", err)
	}
}

func TestZonesService_GetRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/records/1539", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getZoneRecord/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountID := "1010"

	recordResponse, err := client.Zones.GetRecord(accountID, "example.com", 1539)
	if err != nil {
		t.Fatalf("Zones.GetRecord() returned error: %v", err)
	}

	record := recordResponse.Data
	wantSingle := &ZoneRecord{
		ID:           5,
		ZoneID:       "example.com",
		ParentID:     0,
		Type:         "MX",
		Name:         "",
		Content:      "mxa.example.com",
		TTL:          600,
		Priority:     10,
		SystemRecord: false,
		Regions:      []string{"SV1", "IAD"},
		CreatedAt:    "2016-10-05T09:51:35Z",
		UpdatedAt:    "2016-10-05T09:51:35Z"}

	if !reflect.DeepEqual(record, wantSingle) {
		t.Fatalf("Zones.GetRecord() returned %+v, want %+v", record, wantSingle)
	}
}

func TestZonesService_UpdateRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/records/5", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/updateZoneRecord/success.http")

		testMethod(t, r, "PATCH")
		testHeaders(t, r)

		want := map[string]interface{}{"name": "foo", "content": "127.0.0.1"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountID := "1010"
	recordValues := ZoneRecord{Name: "foo", Content: "127.0.0.1"}

	recordResponse, err := client.Zones.UpdateRecord(accountID, "example.com", 5, recordValues)
	if err != nil {
		t.Fatalf("Zones.UpdateRecord() returned error: %v", err)
	}

	record := recordResponse.Data
	if want, got := 5, record.ID; want != got {
		t.Fatalf("Zones.UpdateRecord() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "mxb.example.com", record.Content; want != got {
		t.Fatalf("Zones.UpdateRecord() returned Label expected to be `%v`, got `%v`", want, got)
	}
}

func TestZonesService_UpdateRecord_Regions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	var recordValues ZoneRecord

	mux.HandleFunc("/v2/1/zones/example.com/records/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/updateZoneRecord/success.http")

		want := map[string]interface{}{"name": "foo"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordValues = ZoneRecord{Name: "foo", Regions: []string{}}
	if _, err := client.Zones.UpdateRecord("1", "example.com", 1, recordValues); err != nil {
		t.Fatalf("Zones.UpdateRecord() returned error: %v", err)
	}

	mux.HandleFunc("/v2/2/zones/example.com/records/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/updateZoneRecord/success.http")

		want := map[string]interface{}{"name": "foo", "regions": []interface{}{"global"}}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordValues = ZoneRecord{Name: "foo", Regions: []string{"global"}}
	if _, err := client.Zones.UpdateRecord("2", "example.com", 1, recordValues); err != nil {
		t.Fatalf("Zones.UpdateRecord() returned error: %v", err)
	}

	mux.HandleFunc("/v2/3/zones/example.com/records/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/updateZoneRecord/success.http")

		want := map[string]interface{}{"name": "foo", "regions": []interface{}{"global"}}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	recordValues = ZoneRecord{Name: "foo", Regions: []string{"global"}}
	if _, err := client.Zones.UpdateRecord("2", "example.com", 1, recordValues); err != nil {
		t.Fatalf("Zones.UpdateRecord() returned error: %v", err)
	}
}

func TestZonesService_DeleteRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/records/2", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/deleteZoneRecord/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountID := "1010"

	_, err := client.Zones.DeleteRecord(accountID, "example.com", 2)
	if err != nil {
		t.Fatalf("Zones.DeleteRecord() returned error: %v", err)
	}
}
