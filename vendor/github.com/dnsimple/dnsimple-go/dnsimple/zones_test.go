package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestZonesService_ListZones(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listZones/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	zonesResponse, err := client.Zones.ListZones("1010", nil)
	if err != nil {
		t.Fatalf("Zones.ListZones() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 2}), zonesResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Zones.ListZones() pagination expected to be %v, got %v", want, got)
	}

	zones := zonesResponse.Data
	if want, got := 2, len(zones); want != got {
		t.Errorf("Zones.ListZones() expected to return %v zones, got %v", want, got)
	}

	if want, got := 1, zones[0].ID; want != got {
		t.Fatalf("Zones.ListZones() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "example-alpha.com", zones[0].Name; want != got {
		t.Fatalf("Zones.ListZones() returned Name expected to be `%v`, got `%v`", want, got)
	}
}

func TestZonesService_ListZones_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listZones/success.http")

		testQuery(t, r, url.Values{
			"page":      []string{"2"},
			"per_page":  []string{"20"},
			"sort":      []string{"name,expiration:desc"},
			"name_like": []string{"example"},
		})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Zones.ListZones("1010", &ZoneListOptions{"example", ListOptions{Page: 2, PerPage: 20, Sort: "name,expiration:desc"}})
	if err != nil {
		t.Fatalf("Zones.ListZones() returned error: %v", err)
	}
}

func TestZonesService_GetZone(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getZone/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountID := "1010"
	zoneName := "example.com"

	zoneResponse, err := client.Zones.GetZone(accountID, zoneName)
	if err != nil {
		t.Fatalf("Zones.GetZone() returned error: %v", err)
	}

	zone := zoneResponse.Data
	wantSingle := &Zone{
		ID:        1,
		AccountID: 1010,
		Name:      "example-alpha.com",
		Reverse:   false,
		CreatedAt: "2015-04-23T07:40:03Z",
		UpdatedAt: "2015-04-23T07:40:03Z"}

	if !reflect.DeepEqual(zone, wantSingle) {
		t.Fatalf("Zones.GetZone() returned %+v, want %+v", zone, wantSingle)
	}
}

func TestZonesService_GetZoneFile(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/zones/example.com/file", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getZoneFile/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountID := "1010"
	zoneName := "example.com"

	zoneFileResponse, err := client.Zones.GetZoneFile(accountID, zoneName)
	if err != nil {
		t.Fatalf("Zones.GetZoneFile() returned error: %v", err)
	}

	zoneFile := zoneFileResponse.Data
	wantSingle := &ZoneFile{
		Zone: "$ORIGIN example.com.\n$TTL 1h\nexample.com. 3600 IN SOA ns1.dnsimple.com. admin.dnsimple.com. 1453132552 86400 7200 604800 300\nexample.com. 3600 IN NS ns1.dnsimple.com.\nexample.com. 3600 IN NS ns2.dnsimple.com.\nexample.com. 3600 IN NS ns3.dnsimple.com.\nexample.com. 3600 IN NS ns4.dnsimple.com.\n",
	}

	if !reflect.DeepEqual(zoneFile, wantSingle) {
		t.Fatalf("Zones.GetZoneFile() returned %+v, want %+v", zoneFile, wantSingle)
	}
}
