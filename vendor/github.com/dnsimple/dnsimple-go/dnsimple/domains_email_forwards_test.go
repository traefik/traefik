package dnsimple

import (
	"io"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"testing"
)

var regexpEmail = regexp.MustCompile(`.+@.+`)

func TestEmailForwardPath(t *testing.T) {
	if want, got := "/1010/domains/example.com/email_forwards", emailForwardPath("1010", "example.com", 0); want != got {
		t.Errorf("emailForwardPath(%v) = %v, want %v", "", got, want)
	}

	if want, got := "/1010/domains/example.com/email_forwards/2", emailForwardPath("1010", "example.com", 2); want != got {
		t.Errorf("emailForwardPath(%v) = %v, want %v", "2", got, want)
	}
}

func TestDomainsService_EmailForwardsList(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/email_forwards", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listEmailForwards/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	forwardsResponse, err := client.Domains.ListEmailForwards("1010", "example.com", nil)
	if err != nil {
		t.Fatalf("Domains.ListEmailForwards() returned error: %v", err)
	}

	if want, got := (&Pagination{CurrentPage: 1, PerPage: 30, TotalPages: 1, TotalEntries: 2}), forwardsResponse.Pagination; !reflect.DeepEqual(want, got) {
		t.Errorf("Domains.ListEmailForwards() pagination expected to be %v, got %v", want, got)
	}

	forwards := forwardsResponse.Data
	if want, got := 2, len(forwards); want != got {
		t.Errorf("Domains.ListEmailForwards() expected to return %v contacts, got %v", want, got)
	}

	if want, got := 17702, forwards[0].ID; want != got {
		t.Fatalf("Domains.ListEmailForwards() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if !regexpEmail.MatchString(forwards[0].From) {
		t.Errorf("Domains.ListEmailForwards() From expected to be an email, got %v", forwards[0].From)
	}
}

func TestDomainsService_EmailForwardsList_WithOptions(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/email_forwards", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listEmailForwards/success.http")

		testQuery(t, r, url.Values{"page": []string{"2"}, "per_page": []string{"20"}})

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Domains.ListEmailForwards("1010", "example.com", &ListOptions{Page: 2, PerPage: 20})
	if err != nil {
		t.Fatalf("Domains.ListEmailForwards() returned error: %v", err)
	}
}

func TestDomainsService_CreateEmailForward(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/email_forwards", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createEmailForward/created.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"from": "me"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	forwardAttributes := EmailForward{From: "me"}

	forwardResponse, err := client.Domains.CreateEmailForward("1010", "example.com", forwardAttributes)
	if err != nil {
		t.Fatalf("Domains.CreateEmailForward() returned error: %v", err)
	}

	forward := forwardResponse.Data
	if want, got := 17706, forward.ID; want != got {
		t.Fatalf("Domains.CreateEmailForward() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if !regexpEmail.MatchString(forward.From) {
		t.Errorf("Domains.CreateEmailForward() From expected to be an email, got %v", forward.From)
	}
}

func TestDomainsService_GetEmailForward(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/email_forwards/2", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getEmailForward/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	forwardResponse, err := client.Domains.GetEmailForward("1010", "example.com", 2)
	if err != nil {
		t.Errorf("Domains.GetEmailForward() returned error: %v", err)
	}

	forward := forwardResponse.Data
	wantSingle := &EmailForward{
		ID:        17706,
		DomainID:  228963,
		From:      "jim@a-domain.com",
		To:        "jim@another.com",
		CreatedAt: "2016-02-04T14:26:50Z",
		UpdatedAt: "2016-02-04T14:26:50Z"}

	if !reflect.DeepEqual(forward, wantSingle) {
		t.Fatalf("Domains.GetEmailForward() returned %+v, want %+v", forward, wantSingle)
	}
}

func TestDomainsService_DeleteEmailForward(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/domains/example.com/email_forwards/2", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/deleteEmailForward/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Domains.DeleteEmailForward("1010", "example.com", 2)
	if err != nil {
		t.Fatalf("Domains.DeleteEmailForward() returned error: %v", err)
	}
}
