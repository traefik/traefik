package dnsimple

import (
	"io"
	"net/http"
	"reflect"
	"testing"
)

func TestWebhooks_webhookPath(t *testing.T) {
	if want, got := "/1010/webhooks", webhookPath("1010", 0); want != got {
		t.Errorf("webhookPath(%v,  ) = %v, want %v", "1010", got, want)
	}

	if want, got := "/1010/webhooks/1", webhookPath("1010", 1); want != got {
		t.Errorf("webhookPath(%v, 1) = %v, want %v", "1010", got, want)
	}
}

func TestWebhooksService_ListWebhooks(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/webhooks", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/listWebhooks/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	webhooksResponse, err := client.Webhooks.ListWebhooks("1010", nil)
	if err != nil {
		t.Fatalf("Webhooks.List() returned error: %v", err)
	}

	webhooks := webhooksResponse.Data
	if want, got := 2, len(webhooks); want != got {
		t.Errorf("Webhooks.List() expected to return %v webhooks, got %v", want, got)
	}

	if want, got := 1, webhooks[0].ID; want != got {
		t.Fatalf("Webhooks.List() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "https://webhook.test", webhooks[0].URL; want != got {
		t.Fatalf("Webhooks.List() returned URL expected to be `%v`, got `%v`", want, got)
	}
}

func TestWebhooksService_CreateWebhook(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/webhooks", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/createWebhook/created.http")

		testMethod(t, r, "POST")
		testHeaders(t, r)

		want := map[string]interface{}{"url": "https://webhook.test"}
		testRequestJSON(t, r, want)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	webhookAttributes := Webhook{URL: "https://webhook.test"}

	webhookResponse, err := client.Webhooks.CreateWebhook("1010", webhookAttributes)
	if err != nil {
		t.Fatalf("Webhooks.Create() returned error: %v", err)
	}

	webhook := webhookResponse.Data
	if want, got := 1, webhook.ID; want != got {
		t.Fatalf("Webhooks.Create() returned ID expected to be `%v`, got `%v`", want, got)
	}
	if want, got := "https://webhook.test", webhook.URL; want != got {
		t.Fatalf("Webhooks.Create() returned URL expected to be `%v`, got `%v`", want, got)
	}
}

func TestWebhooksService_GetWebhook(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/webhooks/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/getWebhook/success.http")

		testMethod(t, r, "GET")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	webhookResponse, err := client.Webhooks.GetWebhook("1010", 1)
	if err != nil {
		t.Fatalf("Webhooks.Get() returned error: %v", err)
	}

	webhook := webhookResponse.Data
	wantSingle := &Webhook{
		ID:  1,
		URL: "https://webhook.test"}

	if !reflect.DeepEqual(webhook, wantSingle) {
		t.Fatalf("Webhooks.Get() returned %+v, want %+v", webhook, wantSingle)
	}
}

func TestWebhooksService_DeleteWebhook(t *testing.T) {
	setupMockServer()
	defer teardownMockServer()

	mux.HandleFunc("/v2/1010/webhooks/1", func(w http.ResponseWriter, r *http.Request) {
		httpResponse := httpResponseFixture(t, "/deleteWebhook/success.http")

		testMethod(t, r, "DELETE")
		testHeaders(t, r)

		w.WriteHeader(httpResponse.StatusCode)
		io.Copy(w, httpResponse.Body)
	})

	_, err := client.Webhooks.DeleteWebhook("1010", 1)
	if err != nil {
		t.Fatalf("Webhooks.Delete() returned error: %v", err)
	}
}
