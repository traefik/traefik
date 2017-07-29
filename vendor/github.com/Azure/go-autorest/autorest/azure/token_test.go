package azure

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/mocks"
)

const (
	defaultFormData       = "client_id=id&client_secret=secret&grant_type=client_credentials&resource=resource"
	defaultManualFormData = "client_id=id&grant_type=refresh_token&refresh_token=refreshtoken&resource=resource"
)

func TestTokenExpires(t *testing.T) {
	tt := time.Now().Add(5 * time.Second)
	tk := newTokenExpiresAt(tt)

	if tk.Expires().Equal(tt) {
		t.Fatalf("azure: Token#Expires miscalculated expiration time -- received %v, expected %v", tk.Expires(), tt)
	}
}

func TestTokenIsExpired(t *testing.T) {
	tk := newTokenExpiresAt(time.Now().Add(-5 * time.Second))

	if !tk.IsExpired() {
		t.Fatalf("azure: Token#IsExpired failed to mark a stale token as expired -- now %v, token expires at %v",
			time.Now().UTC(), tk.Expires())
	}
}

func TestTokenIsExpiredUninitialized(t *testing.T) {
	tk := &Token{}

	if !tk.IsExpired() {
		t.Fatalf("azure: An uninitialized Token failed to mark itself as expired (expiration time %v)", tk.Expires())
	}
}

func TestTokenIsNoExpired(t *testing.T) {
	tk := newTokenExpiresAt(time.Now().Add(1000 * time.Second))

	if tk.IsExpired() {
		t.Fatalf("azure: Token marked a fresh token as expired -- now %v, token expires at %v", time.Now().UTC(), tk.Expires())
	}
}

func TestTokenWillExpireIn(t *testing.T) {
	d := 5 * time.Second
	tk := newTokenExpiresIn(d)

	if !tk.WillExpireIn(d) {
		t.Fatal("azure: Token#WillExpireIn mismeasured expiration time")
	}
}

func TestTokenWithAuthorization(t *testing.T) {
	tk := newToken()

	req, err := autorest.Prepare(&http.Request{}, tk.WithAuthorization())
	if err != nil {
		t.Fatalf("azure: Token#WithAuthorization returned an error (%v)", err)
	} else if req.Header.Get(http.CanonicalHeaderKey("Authorization")) != fmt.Sprintf("Bearer %s", tk.AccessToken) {
		t.Fatal("azure: Token#WithAuthorization failed to set Authorization header")
	}
}

func TestServicePrincipalTokenSetAutoRefresh(t *testing.T) {
	spt := newServicePrincipalToken()

	if !spt.autoRefresh {
		t.Fatal("azure: ServicePrincipalToken did not default to automatic token refreshing")
	}

	spt.SetAutoRefresh(false)
	if spt.autoRefresh {
		t.Fatal("azure: ServicePrincipalToken#SetAutoRefresh did not disable automatic token refreshing")
	}
}

func TestServicePrincipalTokenSetRefreshWithin(t *testing.T) {
	spt := newServicePrincipalToken()

	if spt.refreshWithin != defaultRefresh {
		t.Fatal("azure: ServicePrincipalToken did not correctly set the default refresh interval")
	}

	spt.SetRefreshWithin(2 * defaultRefresh)
	if spt.refreshWithin != 2*defaultRefresh {
		t.Fatal("azure: ServicePrincipalToken#SetRefreshWithin did not set the refresh interval")
	}
}

func TestServicePrincipalTokenSetSender(t *testing.T) {
	spt := newServicePrincipalToken()

	var s autorest.Sender
	s = mocks.NewSender()
	spt.SetSender(s)
	if !reflect.DeepEqual(s, spt.sender) {
		t.Fatal("azure: ServicePrincipalToken#SetSender did not set the sender")
	}
}

func TestServicePrincipalTokenRefreshUsesPOST(t *testing.T) {
	spt := newServicePrincipalToken()

	body := mocks.NewBody("")
	resp := mocks.NewResponseWithBodyAndStatus(body, 200, "OK")

	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					if r.Method != "POST" {
						t.Fatalf("azure: ServicePrincipalToken#Refresh did not correctly set HTTP method -- expected %v, received %v", "POST", r.Method)
					}
					return resp, nil
				})
			}
		})())
	spt.SetSender(s)
	spt.Refresh()

	if body.IsOpen() {
		t.Fatalf("the response was not closed!")
	}
}

func TestServicePrincipalTokenRefreshSetsMimeType(t *testing.T) {
	spt := newServicePrincipalToken()

	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					if r.Header.Get(http.CanonicalHeaderKey("Content-Type")) != "application/x-www-form-urlencoded" {
						t.Fatalf("azure: ServicePrincipalToken#Refresh did not correctly set Content-Type -- expected %v, received %v",
							"application/x-form-urlencoded",
							r.Header.Get(http.CanonicalHeaderKey("Content-Type")))
					}
					return mocks.NewResponse(), nil
				})
			}
		})())
	spt.SetSender(s)
	spt.Refresh()
}

func TestServicePrincipalTokenRefreshSetsURL(t *testing.T) {
	spt := newServicePrincipalToken()

	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					if r.URL.String() != TestOAuthConfig.TokenEndpoint.String() {
						t.Fatalf("azure: ServicePrincipalToken#Refresh did not correctly set the URL -- expected %v, received %v",
							TestOAuthConfig.TokenEndpoint, r.URL)
					}
					return mocks.NewResponse(), nil
				})
			}
		})())
	spt.SetSender(s)
	spt.Refresh()
}

func testServicePrincipalTokenRefreshSetsBody(t *testing.T, spt *ServicePrincipalToken, f func(*testing.T, []byte)) {
	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					b, err := ioutil.ReadAll(r.Body)
					if err != nil {
						t.Fatalf("azure: Failed to read body of Service Principal token request (%v)", err)
					}
					f(t, b)
					return mocks.NewResponse(), nil
				})
			}
		})())
	spt.SetSender(s)
	spt.Refresh()
}

func TestServicePrincipalTokenManualRefreshSetsBody(t *testing.T) {
	sptManual := newServicePrincipalTokenManual()
	testServicePrincipalTokenRefreshSetsBody(t, sptManual, func(t *testing.T, b []byte) {
		if string(b) != defaultManualFormData {
			t.Fatalf("azure: ServicePrincipalToken#Refresh did not correctly set the HTTP Request Body -- expected %v, received %v",
				defaultManualFormData, string(b))
		}
	})
}

func TestServicePrincipalTokenCertficateRefreshSetsBody(t *testing.T) {
	sptCert := newServicePrincipalTokenCertificate(t)
	testServicePrincipalTokenRefreshSetsBody(t, sptCert, func(t *testing.T, b []byte) {
		body := string(b)

		values, _ := url.ParseQuery(body)
		if values["client_assertion_type"][0] != "urn:ietf:params:oauth:client-assertion-type:jwt-bearer" ||
			values["client_id"][0] != "id" ||
			values["grant_type"][0] != "client_credentials" ||
			values["resource"][0] != "resource" {
			t.Fatalf("azure: ServicePrincipalTokenCertificate#Refresh did not correctly set the HTTP Request Body.")
		}
	})
}

func TestServicePrincipalTokenSecretRefreshSetsBody(t *testing.T) {
	spt := newServicePrincipalToken()
	testServicePrincipalTokenRefreshSetsBody(t, spt, func(t *testing.T, b []byte) {
		if string(b) != defaultFormData {
			t.Fatalf("azure: ServicePrincipalToken#Refresh did not correctly set the HTTP Request Body -- expected %v, received %v",
				defaultFormData, string(b))
		}

	})
}

func TestServicePrincipalTokenRefreshClosesRequestBody(t *testing.T) {
	spt := newServicePrincipalToken()

	resp := mocks.NewResponse()
	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					return resp, nil
				})
			}
		})())
	spt.SetSender(s)
	spt.Refresh()

	if resp.Body.(*mocks.Body).IsOpen() {
		t.Fatal("azure: ServicePrincipalToken#Refresh failed to close the HTTP Response Body")
	}
}

func TestServicePrincipalTokenRefreshPropagatesErrors(t *testing.T) {
	spt := newServicePrincipalToken()

	c := mocks.NewSender()
	c.SetError(fmt.Errorf("Faux Error"))
	spt.SetSender(c)

	err := spt.Refresh()
	if err == nil {
		t.Fatal("azure: Failed to propagate the request error")
	}
}

func TestServicePrincipalTokenRefreshReturnsErrorIfNotOk(t *testing.T) {
	spt := newServicePrincipalToken()

	c := mocks.NewSender()
	c.AppendResponse(mocks.NewResponseWithStatus("401 NotAuthorized", 401))
	spt.SetSender(c)

	err := spt.Refresh()
	if err == nil {
		t.Fatal("azure: Failed to return an when receiving a status code other than HTTP 200")
	}
}

func TestServicePrincipalTokenRefreshUnmarshals(t *testing.T) {
	spt := newServicePrincipalToken()

	expiresOn := strconv.Itoa(int(time.Now().Add(3600 * time.Second).Sub(expirationBase).Seconds()))
	j := newTokenJSON(expiresOn, "resource")
	resp := mocks.NewResponseWithContent(j)
	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					return resp, nil
				})
			}
		})())
	spt.SetSender(s)

	err := spt.Refresh()
	if err != nil {
		t.Fatalf("azure: ServicePrincipalToken#Refresh returned an unexpected error (%v)", err)
	} else if spt.AccessToken != "accessToken" ||
		spt.ExpiresIn != "3600" ||
		spt.ExpiresOn != expiresOn ||
		spt.NotBefore != expiresOn ||
		spt.Resource != "resource" ||
		spt.Type != "Bearer" {
		t.Fatalf("azure: ServicePrincipalToken#Refresh failed correctly unmarshal the JSON -- expected %v, received %v",
			j, *spt)
	}
}

func TestServicePrincipalTokenEnsureFreshRefreshes(t *testing.T) {
	spt := newServicePrincipalToken()
	expireToken(&spt.Token)

	f := false
	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					f = true
					return mocks.NewResponse(), nil
				})
			}
		})())
	spt.SetSender(s)
	spt.EnsureFresh()
	if !f {
		t.Fatal("azure: ServicePrincipalToken#EnsureFresh failed to call Refresh for stale token")
	}
}

func TestServicePrincipalTokenEnsureFreshSkipsIfFresh(t *testing.T) {
	spt := newServicePrincipalToken()
	setTokenToExpireIn(&spt.Token, 1000*time.Second)

	f := false
	c := mocks.NewSender()
	s := autorest.DecorateSender(c,
		(func() autorest.SendDecorator {
			return func(s autorest.Sender) autorest.Sender {
				return autorest.SenderFunc(func(r *http.Request) (*http.Response, error) {
					f = true
					return mocks.NewResponse(), nil
				})
			}
		})())
	spt.SetSender(s)
	spt.EnsureFresh()
	if f {
		t.Fatal("azure: ServicePrincipalToken#EnsureFresh invoked Refresh for fresh token")
	}
}

func TestServicePrincipalTokenWithAuthorization(t *testing.T) {
	spt := newServicePrincipalToken()
	setTokenToExpireIn(&spt.Token, 1000*time.Second)
	r := mocks.NewRequest()
	s := mocks.NewSender()
	spt.SetSender(s)

	req, err := autorest.Prepare(r, spt.WithAuthorization())
	if err != nil {
		t.Fatalf("azure: ServicePrincipalToken#WithAuthorization returned an error (%v)", err)
	} else if req.Header.Get(http.CanonicalHeaderKey("Authorization")) != fmt.Sprintf("Bearer %s", spt.AccessToken) {
		t.Fatal("azure: ServicePrincipalToken#WithAuthorization failed to set Authorization header")
	}
}

func TestServicePrincipalTokenWithAuthorizationReturnsErrorIfCannotRefresh(t *testing.T) {
	spt := newServicePrincipalToken()
	s := mocks.NewSender()
	s.AppendResponse(mocks.NewResponseWithStatus("400 Bad Request", 400))
	spt.SetSender(s)

	_, err := autorest.Prepare(mocks.NewRequest(), spt.WithAuthorization())
	if err == nil {
		t.Fatal("azure: ServicePrincipalToken#WithAuthorization failed to return an error when refresh fails")
	}
}

func TestRefreshCallback(t *testing.T) {
	callbackTriggered := false
	spt := newServicePrincipalToken(func(Token) error {
		callbackTriggered = true
		return nil
	})

	expiresOn := strconv.Itoa(int(time.Now().Add(3600 * time.Second).Sub(expirationBase).Seconds()))

	sender := mocks.NewSender()
	j := newTokenJSON(expiresOn, "resource")
	sender.AppendResponse(mocks.NewResponseWithContent(j))
	spt.SetSender(sender)
	spt.Refresh()

	if !callbackTriggered {
		t.Fatalf("azure: RefreshCallback failed to trigger call callback")
	}
}

func TestRefreshCallbackErrorPropagates(t *testing.T) {
	errorText := "this is an error text"
	spt := newServicePrincipalToken(func(Token) error {
		return fmt.Errorf(errorText)
	})

	expiresOn := strconv.Itoa(int(time.Now().Add(3600 * time.Second).Sub(expirationBase).Seconds()))

	sender := mocks.NewSender()
	j := newTokenJSON(expiresOn, "resource")
	sender.AppendResponse(mocks.NewResponseWithContent(j))
	spt.SetSender(sender)
	err := spt.Refresh()

	if err == nil || !strings.Contains(err.Error(), errorText) {
		t.Fatalf("azure: RefreshCallback failed to propagate error")
	}
}

// This demonstrates the danger of manual token without a refresh token
func TestServicePrincipalTokenManualRefreshFailsWithoutRefresh(t *testing.T) {
	spt := newServicePrincipalTokenManual()
	spt.RefreshToken = ""
	err := spt.Refresh()
	if err == nil {
		t.Fatalf("azure: ServicePrincipalToken#Refresh should have failed with a ManualTokenSecret without a refresh token")
	}
}

func newToken() *Token {
	return &Token{
		AccessToken: "ASECRETVALUE",
		Resource:    "https://azure.microsoft.com/",
		Type:        "Bearer",
	}
}

func newTokenJSON(expiresOn string, resource string) string {
	return fmt.Sprintf(`{
		"access_token" : "accessToken",
		"expires_in"   : "3600",
		"expires_on"   : "%s",
		"not_before"   : "%s",
		"resource"     : "%s",
		"token_type"   : "Bearer"
		}`,
		expiresOn, expiresOn, resource)
}

func newTokenExpiresIn(expireIn time.Duration) *Token {
	return setTokenToExpireIn(newToken(), expireIn)
}

func newTokenExpiresAt(expireAt time.Time) *Token {
	return setTokenToExpireAt(newToken(), expireAt)
}

func expireToken(t *Token) *Token {
	return setTokenToExpireIn(t, 0)
}

func setTokenToExpireAt(t *Token, expireAt time.Time) *Token {
	t.ExpiresIn = "3600"
	t.ExpiresOn = strconv.Itoa(int(expireAt.Sub(expirationBase).Seconds()))
	t.NotBefore = t.ExpiresOn
	return t
}

func setTokenToExpireIn(t *Token, expireIn time.Duration) *Token {
	return setTokenToExpireAt(t, time.Now().Add(expireIn))
}

func newServicePrincipalToken(callbacks ...TokenRefreshCallback) *ServicePrincipalToken {
	spt, _ := NewServicePrincipalToken(TestOAuthConfig, "id", "secret", "resource", callbacks...)
	return spt
}

func newServicePrincipalTokenManual() *ServicePrincipalToken {
	token := newToken()
	token.RefreshToken = "refreshtoken"
	spt, _ := NewServicePrincipalTokenFromManualToken(TestOAuthConfig, "id", "resource", *token)
	return spt
}

func newServicePrincipalTokenCertificate(t *testing.T) *ServicePrincipalToken {
	template := x509.Certificate{
		SerialNumber:          big.NewInt(0),
		Subject:               pkix.Name{CommonName: "test"},
		BasicConstraintsValid: true,
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	certificateBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	certificate, err := x509.ParseCertificate(certificateBytes)
	if err != nil {
		t.Fatal(err)
	}

	spt, _ := NewServicePrincipalTokenFromCertificate(TestOAuthConfig, "id", certificate, privateKey, "resource")
	return spt
}
