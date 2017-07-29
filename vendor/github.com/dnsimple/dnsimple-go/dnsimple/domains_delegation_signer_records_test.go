package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestDelegationSignerRecordPath(t *testing.T) {
	if want, got := "/1010/domains/example.com/ds_records", delegationSignerRecordPath("1010", "example.com", 0); want != got {
		t.Errorf("delegationSignerRecordPath(%v) = %v, want %v", "", got, want)
	}

	if want, got := "/1010/domains/example.com/ds_records/2", delegationSignerRecordPath("1010", "example.com", 2); want != got {
		t.Errorf("delegationSignerRecordPath(%v) = %v, want %v", "2", got, want)
	}
}

func TestDomainsService_ListDelegationSignerRecords(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/ds_records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listDelegationSignerRecords/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	dsRecordsResponse, err := client.Domains.ListDelegationSignerRecords("1010", "example.com", nil)
	if err != nil {
		t.Fatalf("Domains.ListDelegationSignerRecords() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 1}), dsRecordsResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Domains.ListDelegationSignerRecords() pagination expected to be %v, got %v", want, got)
	}

	dsRecords := dsRecordsResponse.Data
	if want, got := 1, len(dsRecords); want != got {
		t.Errorf("Domains.ListDelegationSignerRecords() expected to return %v delegation signer records, got %v", want, got)
	}

	if want, got := 24, dsRecords[0].ID; want != got {
		t.Fatalf("Domains.ListDelegationSignerRecords() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "8", dsRecords[0].Algorithm; want != got {
		t.Fatalf("Domains.ListDelegationSignerRecords() returned Algorithm expected to be `%v`, got `%v`", want, got)
	}
}

func TestDomainsService_ListDelegationSignerRecords_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/ds_records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listDelegationSignerRecords/success.http")

		testQuery(t, r, url.Values{"page": []string{"2"}, "per_page": []string{"20"}})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Domains.ListDelegationSignerRecords("1010", "example.com", &ListOptions{Page: 2, PerPage: 20})
	if err != nil {
		t.Fatalf("Domains.ListDelegationSignerRecords() returned error: %v", err)
	}
}

func TestDomainsService_CreateDelegationSignerRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/ds_records", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createDelegationSignerRecord/created.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"algorithm": "13", "digest": "ABC123", "digest_type": "2", "keytag": "1234"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	dsRecordAttributes := DelegationSignerRecord{Algorithm: "13", Digest: "ABC123", DigestType: "2", Keytag: "1234"}

	dsRecordResponse, err := client.Domains.CreateDelegationSignerRecord("1010", "example.com", dsRecordAttributes)
	if err != nil {
		t.Fatalf("Domains.CreateDelegationSignerRecord() returned error: %v", err)
	}

	dsRecord := dsRecordResponse.Data
	if want, got := 2, dsRecord.ID; want != got {
		t.Fatalf("Domains.CreateDelegationSignerRecord() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "13", dsRecord.Algorithm; want != got {
		t.Errorf("Domains.CreateDelegationSignerRecord() returned Algorithm expected to be `%v`, got %v", want, got)
	}
}

func TestDomainsService_GetDelegationSignerRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/ds_records/2", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getDelegationSignerRecord/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	dsRecordResponse, err := client.Domains.GetDelegationSignerRecord("1010", "example.com", 2)
	if err != nil {
		t.Errorf("Domains.GetDelegationSignerRecord() returned error: %v", err)
	}

	dsRecord := dsRecordResponse.Data
	wantSingle := &DelegationSignerRecord{
		ID:         24,
		DomainID:   1010,
		Algorithm:  "8",
		DigestType: "2",
		Digest:     "C1F6E04A5A61FBF65BF9DC8294C363CF11C89E802D926BDAB79C55D27BEFA94F",
		Keytag:     "44620",
		CreatedAt:  "2017-03-03T13:49:58Z",
		UpdatedAt:  "2017-03-03T13:49:58Z"}

	if !reflect.DeepEqual(dsRecord, wantSingle) {
		t.Fatalf("Domains.GetDelegationSignerRecord() returned %+v, want %+v", dsRecord, wantSingle)
	}
}

func TestDomainsService_DeleteDelegationSignerRecord(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/ds_records/2", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/deleteDelegationSignerRecord/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Domains.DeleteDelegationSignerRecord("1010", "example.com", 2)
	if err != nil {
		t.Fatalf("Domains.DeleteDelegationSignerRecord() returned error: %v", err)
	}
}
