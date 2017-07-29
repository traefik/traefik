package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"testing"
)

func TestAccountsService_List(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/accounts", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listAccounts/success-user.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)
		testQuery(t, r, url.Values{})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	accountsResponse, err := client.Accounts.ListAccounts(nil)
	if err != nil {
		t.Fatalf("Accounts.ListAccounts() returned error: %v", err)
	}

	accounts := accountsResponse.Data
	if want, got := 2, len(accounts); want != got {
		t.Errorf("Accounts.ListAccounts() expected to return %v accounts, got %v", want, got)
	}

	if want, got := 123, accounts[0].ID; want != got {
		t.Fatalf("Accounts.ListAccounts() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "john@example.com", accounts[0].Email; want != got {
		t.Fatalf("Accounts.ListAccounts() returned Email expected to be `%v`, got `%v`", want, got)
	}
}
