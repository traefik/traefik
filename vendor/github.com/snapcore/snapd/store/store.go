// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2018 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

// Package store has support to use the Ubuntu Store for querying and downloading of snaps, and the related services.
package store

import (
	"bytes"
	"crypto"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/juju/ratelimit"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
	"gopkg.in/retry.v1"

	"github.com/snapcore/snapd/arch"
	"github.com/snapcore/snapd/asserts"
	"github.com/snapcore/snapd/dirs"
	"github.com/snapcore/snapd/httputil"
	"github.com/snapcore/snapd/i18n"
	"github.com/snapcore/snapd/jsonutil"
	"github.com/snapcore/snapd/logger"
	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/overlord/auth"
	"github.com/snapcore/snapd/progress"
	"github.com/snapcore/snapd/release"
	"github.com/snapcore/snapd/snap"
)

// TODO: better/shorter names are probably in order once fewer legacy places are using this

const (
	// halJsonContentType is the default accept value for store requests
	halJsonContentType = "application/hal+json"
	// jsonContentType is for store enpoints that don't support HAL
	jsonContentType = "application/json"
	// UbuntuCoreWireProtocol is the protocol level we support when
	// communicating with the store. History:
	//  - "1": client supports squashfs snaps
	UbuntuCoreWireProtocol = "1"
)

type RefreshOptions struct {
	// RefreshManaged indicates to the store that the refresh is
	// managed via snapd-control.
	RefreshManaged bool

	PrivacyKey string
}

// the LimitTime should be slightly more than 3 times of our http.Client
// Timeout value
var defaultRetryStrategy = retry.LimitCount(5, retry.LimitTime(38*time.Second,
	retry.Exponential{
		Initial: 300 * time.Millisecond,
		Factor:  2.5,
	},
))

var connCheckStrategy = retry.LimitCount(3, retry.LimitTime(38*time.Second,
	retry.Exponential{
		Initial: 900 * time.Millisecond,
		Factor:  1.3,
	},
))

// Config represents the configuration to access the snap store
type Config struct {
	// Store API base URLs. The assertions url is only separate because it can
	// be overridden by its own env var.
	StoreBaseURL      *url.URL
	AssertionsBaseURL *url.URL

	// StoreID is the store id used if we can't get one through the AuthContext.
	StoreID string

	Architecture string
	Series       string

	DetailFields []string
	InfoFields   []string
	DeltaFormat  string

	// CacheDownloads is the number of downloads that should be cached
	CacheDownloads int

	// Proxy returns the HTTP proxy to use when talking to the store
	Proxy func(*http.Request) (*url.URL, error)
}

// setBaseURL updates the store API's base URL in the Config. Must not be used
// to change active config.
func (cfg *Config) setBaseURL(u *url.URL) error {
	storeBaseURI, err := storeURL(u)
	if err != nil {
		return err
	}

	assertsBaseURI, err := assertsURL()
	if err != nil {
		return err
	}

	cfg.StoreBaseURL = storeBaseURI
	cfg.AssertionsBaseURL = assertsBaseURI

	return nil
}

// Store represents the ubuntu snap store
type Store struct {
	cfg *Config

	architecture string
	series       string

	noCDN bool

	fallbackStoreID string

	detailFields []string
	infoFields   []string
	deltaFormat  string
	// reused http client
	client *http.Client

	authContext auth.AuthContext

	mu                sync.Mutex
	suggestedCurrency string

	cacher downloadCache
	proxy  func(*http.Request) (*url.URL, error)
}

func respToError(resp *http.Response, msg string) error {
	tpl := "cannot %s: got unexpected HTTP status code %d via %s to %q"
	if oops := resp.Header.Get("X-Oops-Id"); oops != "" {
		tpl += " [%s]"
		return fmt.Errorf(tpl, msg, resp.StatusCode, resp.Request.Method, resp.Request.URL, oops)
	}

	return fmt.Errorf(tpl, msg, resp.StatusCode, resp.Request.Method, resp.Request.URL)
}

// Deltas enabled by default on classic, but allow opting in or out on both classic and core.
func useDeltas() bool {
	// only xdelta3 is supported for now, so check the binary exists here
	// TODO: have a per-format checker instead
	if _, err := getXdelta3Cmd(); err != nil {
		return false
	}

	return osutil.GetenvBool("SNAPD_USE_DELTAS_EXPERIMENTAL", true)
}

func useStaging() bool {
	return osutil.GetenvBool("SNAPPY_USE_STAGING_STORE")
}

// endpointURL clones a base URL and updates it with optional path and query.
func endpointURL(base *url.URL, path string, query url.Values) *url.URL {
	u := *base
	if path != "" {
		u.Path = strings.TrimSuffix(u.Path, "/") + "/" + strings.TrimPrefix(path, "/")
		u.RawQuery = ""
	}
	if len(query) != 0 {
		u.RawQuery = query.Encode()
	}
	return &u
}

// apiURL returns the system default base API URL.
func apiURL() *url.URL {
	s := "https://api.snapcraft.io/"
	if useStaging() {
		s = "https://api.staging.snapcraft.io/"
	}
	u, _ := url.Parse(s)
	return u
}

// storeURL returns the base store URL, derived from either the given API URL
// or an env var override.
func storeURL(api *url.URL) (*url.URL, error) {
	var override string
	var overrideName string
	// XXX: time to drop FORCE_CPI support
	// XXX: Deprecated but present for backward-compatibility: this used
	// to be "Click Package Index".  Remove this once people have got
	// used to SNAPPY_FORCE_API_URL instead.
	if s := os.Getenv("SNAPPY_FORCE_CPI_URL"); s != "" && strings.HasSuffix(s, "api/v1/") {
		overrideName = "SNAPPY_FORCE_CPI_URL"
		override = strings.TrimSuffix(s, "api/v1/")
	} else if s := os.Getenv("SNAPPY_FORCE_API_URL"); s != "" {
		overrideName = "SNAPPY_FORCE_API_URL"
		override = s
	}
	if override != "" {
		u, err := url.Parse(override)
		if err != nil {
			return nil, fmt.Errorf("invalid %s: %s", overrideName, err)
		}
		return u, nil
	}
	return api, nil
}

func assertsURL() (*url.URL, error) {
	if s := os.Getenv("SNAPPY_FORCE_SAS_URL"); s != "" {
		u, err := url.Parse(s)
		if err != nil {
			return nil, fmt.Errorf("invalid SNAPPY_FORCE_SAS_URL: %s", err)
		}
		return u, nil
	}

	// nil means fallback to store base url
	return nil, nil
}

func authLocation() string {
	if useStaging() {
		return "login.staging.ubuntu.com"
	}
	return "login.ubuntu.com"
}

func authURL() string {
	if u := os.Getenv("SNAPPY_FORCE_SSO_URL"); u != "" {
		return u
	}
	return "https://" + authLocation() + "/api/v2"
}

var defaultStoreDeveloperURL = "https://dashboard.snapcraft.io/"

func storeDeveloperURL() string {
	if useStaging() {
		return "https://dashboard.staging.snapcraft.io/"
	}
	return defaultStoreDeveloperURL
}

var defaultConfig = Config{}

// DefaultConfig returns a copy of the default configuration ready to be adapted.
func DefaultConfig() *Config {
	cfg := defaultConfig
	return &cfg
}

func init() {
	storeBaseURI, err := storeURL(apiURL())
	if err != nil {
		panic(err)
	}
	if storeBaseURI.RawQuery != "" {
		panic("store API URL may not contain query string")
	}
	err = defaultConfig.setBaseURL(storeBaseURI)
	if err != nil {
		panic(err)
	}
	defaultConfig.DetailFields = jsonutil.StructFields((*snapDetails)(nil), "snap_yaml_raw")
	defaultConfig.InfoFields = jsonutil.StructFields((*storeSnap)(nil), "snap-yaml")
}

type searchResults struct {
	Payload struct {
		Packages []*snapDetails `json:"clickindex:package"`
	} `json:"_embedded"`
}

type sectionResults struct {
	Payload struct {
		Sections []struct{ Name string } `json:"clickindex:sections"`
	} `json:"_embedded"`
}

// The default delta format if not configured.
var defaultSupportedDeltaFormat = "xdelta3"

// New creates a new Store with the given access configuration and for given the store id.
func New(cfg *Config, authContext auth.AuthContext) *Store {
	if cfg == nil {
		cfg = &defaultConfig
	}

	detailFields := cfg.DetailFields
	if detailFields == nil {
		detailFields = defaultConfig.DetailFields
	}

	infoFields := cfg.InfoFields
	if infoFields == nil {
		infoFields = defaultConfig.InfoFields
	}

	architecture := cfg.Architecture
	if cfg.Architecture == "" {
		architecture = arch.UbuntuArchitecture()
	}

	series := cfg.Series
	if cfg.Series == "" {
		series = release.Series
	}

	deltaFormat := cfg.DeltaFormat
	if deltaFormat == "" {
		deltaFormat = defaultSupportedDeltaFormat
	}

	store := &Store{
		cfg:             cfg,
		series:          series,
		architecture:    architecture,
		noCDN:           osutil.GetenvBool("SNAPPY_STORE_NO_CDN"),
		fallbackStoreID: cfg.StoreID,
		detailFields:    detailFields,
		infoFields:      infoFields,
		authContext:     authContext,
		deltaFormat:     deltaFormat,
		proxy:           cfg.Proxy,

		client: httputil.NewHTTPClient(&httputil.ClientOptions{
			Timeout:    10 * time.Second,
			MayLogBody: true,
			Proxy:      cfg.Proxy,
		}),
	}
	store.SetCacheDownloads(cfg.CacheDownloads)

	return store
}

// API endpoint paths
const (
	// see https://wiki.ubuntu.com/AppStore/Interfaces/ClickPackageIndex
	// XXX: Repeating "api/" here is cumbersome, but the next generation
	// of store APIs will probably drop that prefix (since it now
	// duplicates the hostname), and we may want to switch to v2 APIs
	// one at a time; so it's better to consider that as part of
	// individual endpoint paths.
	searchEndpPath      = "api/v1/snaps/search"
	ordersEndpPath      = "api/v1/snaps/purchases/orders"
	buyEndpPath         = "api/v1/snaps/purchases/buy"
	customersMeEndpPath = "api/v1/snaps/purchases/customers/me"
	sectionsEndpPath    = "api/v1/snaps/sections"
	commandsEndpPath    = "api/v1/snaps/names"
	// v2
	snapActionEndpPath = "v2/snaps/refresh"
	snapInfoEndpPath   = "v2/snaps/info"

	deviceNonceEndpPath   = "api/v1/snaps/auth/nonces"
	deviceSessionEndpPath = "api/v1/snaps/auth/sessions"

	assertionsPath = "api/v1/snaps/assertions"
)

func (s *Store) defaultSnapQuery() url.Values {
	q := url.Values{}
	if len(s.detailFields) != 0 {
		q.Set("fields", strings.Join(s.detailFields, ","))
	}
	return q
}

func (s *Store) baseURL(defaultURL *url.URL) *url.URL {
	u := defaultURL
	if s.authContext != nil {
		var err error
		_, u, err = s.authContext.ProxyStoreParams(defaultURL)
		if err != nil {
			logger.Debugf("cannot get proxy store parameters from state: %v", err)
		}
	}
	if u != nil {
		return u
	}
	return defaultURL
}

func (s *Store) endpointURL(p string, query url.Values) *url.URL {
	return endpointURL(s.baseURL(s.cfg.StoreBaseURL), p, query)
}

func (s *Store) assertionsEndpointURL(p string, query url.Values) *url.URL {
	defBaseURL := s.cfg.StoreBaseURL
	// can be overridden separately!
	if s.cfg.AssertionsBaseURL != nil {
		defBaseURL = s.cfg.AssertionsBaseURL
	}
	return endpointURL(s.baseURL(defBaseURL), path.Join(assertionsPath, p), query)
}

// LoginUser logs user in the store and returns the authentication macaroons.
func LoginUser(username, password, otp string) (string, string, error) {
	macaroon, err := requestStoreMacaroon()
	if err != nil {
		return "", "", err
	}
	deserializedMacaroon, err := auth.MacaroonDeserialize(macaroon)
	if err != nil {
		return "", "", err
	}

	// get SSO 3rd party caveat, and request discharge
	loginCaveat, err := loginCaveatID(deserializedMacaroon)
	if err != nil {
		return "", "", err
	}

	discharge, err := dischargeAuthCaveat(loginCaveat, username, password, otp)
	if err != nil {
		return "", "", err
	}

	return macaroon, discharge, nil
}

// authAvailable returns true if there is a user and/or device session setup
func (s *Store) authAvailable(user *auth.UserState) (bool, error) {
	if user.HasStoreAuth() {
		return true, nil
	} else {
		var device *auth.DeviceState
		var err error
		if s.authContext != nil {
			device, err = s.authContext.Device()
			if err != nil {
				return false, err
			}
		}
		return device != nil && device.SessionMacaroon != "", nil
	}
}

// authenticateUser will add the store expected Macaroon Authorization header for user
func authenticateUser(r *http.Request, user *auth.UserState) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, `Macaroon root="%s"`, user.StoreMacaroon)

	// deserialize root macaroon (we need its signature to do the discharge binding)
	root, err := auth.MacaroonDeserialize(user.StoreMacaroon)
	if err != nil {
		logger.Debugf("cannot deserialize root macaroon: %v", err)
		return
	}

	for _, d := range user.StoreDischarges {
		// prepare discharge for request
		discharge, err := auth.MacaroonDeserialize(d)
		if err != nil {
			logger.Debugf("cannot deserialize discharge macaroon: %v", err)
			return
		}
		discharge.Bind(root.Signature())

		serializedDischarge, err := auth.MacaroonSerialize(discharge)
		if err != nil {
			logger.Debugf("cannot re-serialize discharge macaroon: %v", err)
			return
		}
		fmt.Fprintf(&buf, `, discharge="%s"`, serializedDischarge)
	}
	r.Header.Set("Authorization", buf.String())
}

// refreshDischarges will request refreshed discharge macaroons for the user
func refreshDischarges(user *auth.UserState) ([]string, error) {
	newDischarges := make([]string, len(user.StoreDischarges))
	for i, d := range user.StoreDischarges {
		discharge, err := auth.MacaroonDeserialize(d)
		if err != nil {
			return nil, err
		}
		if discharge.Location() != UbuntuoneLocation {
			newDischarges[i] = d
			continue
		}

		refreshedDischarge, err := refreshDischargeMacaroon(d)
		if err != nil {
			return nil, err
		}
		newDischarges[i] = refreshedDischarge
	}
	return newDischarges, nil
}

// refreshUser will refresh user discharge macaroon and update state
func (s *Store) refreshUser(user *auth.UserState) error {
	if s.authContext == nil {
		return fmt.Errorf("user credentials need to be refreshed but update in place only supported in snapd")
	}
	newDischarges, err := refreshDischarges(user)
	if err != nil {
		return err
	}

	curUser, err := s.authContext.UpdateUserAuth(user, newDischarges)
	if err != nil {
		return err
	}
	// update in place
	*user = *curUser

	return nil
}

// refreshDeviceSession will set or refresh the device session in the state
func (s *Store) refreshDeviceSession(device *auth.DeviceState) error {
	if s.authContext == nil {
		return fmt.Errorf("internal error: no authContext")
	}

	nonce, err := requestStoreDeviceNonce(s.endpointURL(deviceNonceEndpPath, nil).String())
	if err != nil {
		return err
	}

	devSessReqParams, err := s.authContext.DeviceSessionRequestParams(nonce)
	if err != nil {
		return err
	}

	session, err := requestDeviceSession(s.endpointURL(deviceSessionEndpPath, nil).String(), devSessReqParams, device.SessionMacaroon)
	if err != nil {
		return err
	}

	curDevice, err := s.authContext.UpdateDeviceAuth(device, session)
	if err != nil {
		return err
	}
	// update in place
	*device = *curDevice
	return nil
}

// authenticateDevice will add the store expected Macaroon X-Device-Authorization header for device
func authenticateDevice(r *http.Request, device *auth.DeviceState, apiLevel apiLevel) {
	if device.SessionMacaroon != "" {
		r.Header.Set(hdrSnapDeviceAuthorization[apiLevel], fmt.Sprintf(`Macaroon root="%s"`, device.SessionMacaroon))
	}
}

func (s *Store) setStoreID(r *http.Request, apiLevel apiLevel) (customStore bool) {
	storeID := s.fallbackStoreID
	if s.authContext != nil {
		cand, err := s.authContext.StoreID(storeID)
		if err != nil {
			logger.Debugf("cannot get store ID from state: %v", err)
		} else {
			storeID = cand
		}
	}
	if storeID != "" {
		r.Header.Set(hdrSnapDeviceStore[apiLevel], storeID)
		return true
	}
	return false
}

type apiLevel int

const (
	apiV1Endps apiLevel = 0 // api/v1 endpoints
	apiV2Endps apiLevel = 1 // v2 endpoints
)

var (
	hdrSnapDeviceAuthorization = []string{"X-Device-Authorization", "Snap-Device-Authorization"}
	hdrSnapDeviceStore         = []string{"X-Ubuntu-Store", "Snap-Device-Store"}
	hdrSnapDeviceSeries        = []string{"X-Ubuntu-Series", "Snap-Device-Series"}
	hdrSnapDeviceArchitecture  = []string{"X-Ubuntu-Architecture", "Snap-Device-Architecture"}
	hdrSnapClassic             = []string{"X-Ubuntu-Classic", "Snap-Classic"}
)

type deviceAuthNeed int

const (
	deviceAuthPreferred deviceAuthNeed = iota
	deviceAuthCustomStoreOnly
)

// requestOptions specifies parameters for store requests.
type requestOptions struct {
	Method       string
	URL          *url.URL
	Accept       string
	ContentType  string
	APILevel     apiLevel
	ExtraHeaders map[string]string
	Data         []byte

	// DeviceAuthNeed indicates the level of need to supply device
	// authorization for this request, can be:
	//  - deviceAuthPreferred: should be provided if available
	//  - deviceAuthCustomStoreOnly: should be provided only in case
	//    of a custom store
	DeviceAuthNeed deviceAuthNeed
}

func (r *requestOptions) addHeader(k, v string) {
	if r.ExtraHeaders == nil {
		r.ExtraHeaders = make(map[string]string)
	}
	r.ExtraHeaders[k] = v
}

func cancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

var expectedCatalogPreamble = []interface{}{
	json.Delim('{'),
	"_embedded",
	json.Delim('{'),
	"clickindex:package",
	json.Delim('['),
}

type alias struct {
	Name string `json:"name"`
}

type catalogItem struct {
	Name    string   `json:"package_name"`
	Version string   `json:"version"`
	Summary string   `json:"summary"`
	Aliases []alias  `json:"aliases"`
	Apps    []string `json:"apps"`
}

type SnapAdder interface {
	AddSnap(snapName, version, summary string, commands []string) error
}

func decodeCatalog(resp *http.Response, names io.Writer, db SnapAdder) error {
	const what = "decode new commands catalog"
	if resp.StatusCode != 200 {
		return respToError(resp, what)
	}
	dec := json.NewDecoder(resp.Body)
	for _, expectedToken := range expectedCatalogPreamble {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		if token != expectedToken {
			return fmt.Errorf(what+": bad catalog preamble: expected %#v, got %#v", expectedToken, token)
		}
	}

	for dec.More() {
		var v catalogItem
		if err := dec.Decode(&v); err != nil {
			return fmt.Errorf(what+": %v", err)
		}
		if v.Name == "" {
			continue
		}
		fmt.Fprintln(names, v.Name)
		if len(v.Apps) == 0 {
			continue
		}

		commands := make([]string, 0, len(v.Aliases)+len(v.Apps))

		for _, alias := range v.Aliases {
			commands = append(commands, alias.Name)
		}
		for _, app := range v.Apps {
			commands = append(commands, snap.JoinSnapApp(v.Name, app))
		}

		if err := db.AddSnap(v.Name, v.Version, v.Summary, commands); err != nil {
			return err
		}
	}

	return nil
}

func decodeJSONBody(resp *http.Response, success interface{}, failure interface{}) error {
	ok := (resp.StatusCode == 200 || resp.StatusCode == 201)
	// always decode on success; decode failures only if body is not empty
	if !ok && resp.ContentLength == 0 {
		return nil
	}
	result := success
	if !ok {
		result = failure
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

// retryRequestDecodeJSON calls retryRequest and decodes the response into either success or failure.
func (s *Store) retryRequestDecodeJSON(ctx context.Context, reqOptions *requestOptions, user *auth.UserState, success interface{}, failure interface{}) (resp *http.Response, err error) {
	return httputil.RetryRequest(reqOptions.URL.String(), func() (*http.Response, error) {
		return s.doRequest(ctx, s.client, reqOptions, user)
	}, func(resp *http.Response) error {
		return decodeJSONBody(resp, success, failure)
	}, defaultRetryStrategy)
}

// doRequest does an authenticated request to the store handling a potential macaroon refresh required if needed
func (s *Store) doRequest(ctx context.Context, client *http.Client, reqOptions *requestOptions, user *auth.UserState) (*http.Response, error) {
	authRefreshes := 0
	for {
		req, err := s.newRequest(reqOptions, user)
		if err != nil {
			return nil, err
		}

		var resp *http.Response
		if ctx != nil {
			resp, err = ctxhttp.Do(ctx, client, req)
		} else {
			resp, err = client.Do(req)
		}
		if err != nil {
			return nil, err
		}

		wwwAuth := resp.Header.Get("WWW-Authenticate")
		if resp.StatusCode == 401 && authRefreshes < 4 {
			// 4 tries: 2 tries for each in case both user
			// and device need refreshing
			var refreshNeed authRefreshNeed
			if user != nil && strings.Contains(wwwAuth, "needs_refresh=1") {
				// refresh user
				refreshNeed.user = true
			}
			if strings.Contains(wwwAuth, "refresh_device_session=1") {
				// refresh device session
				refreshNeed.device = true
			}
			if refreshNeed.needed() {
				err := s.refreshAuth(user, refreshNeed)
				if err != nil {
					return nil, err
				}
				// close previous response and retry
				resp.Body.Close()
				authRefreshes++
				continue
			}
		}

		return resp, err
	}
}

type authRefreshNeed struct {
	device bool
	user   bool
}

func (rn *authRefreshNeed) needed() bool {
	return rn.device || rn.user
}

func (s *Store) refreshAuth(user *auth.UserState, need authRefreshNeed) error {
	if need.user {
		// refresh user
		err := s.refreshUser(user)
		if err != nil {
			return err
		}
	}
	if need.device {
		// refresh device session
		if s.authContext == nil {
			return fmt.Errorf("internal error: no authContext")
		}
		device, err := s.authContext.Device()
		if err != nil {
			return err
		}

		err = s.refreshDeviceSession(device)
		if err != nil {
			return err
		}
	}
	return nil
}

// build a new http.Request with headers for the store
func (s *Store) newRequest(reqOptions *requestOptions, user *auth.UserState) (*http.Request, error) {
	var body io.Reader
	if reqOptions.Data != nil {
		body = bytes.NewBuffer(reqOptions.Data)
	}

	req, err := http.NewRequest(reqOptions.Method, reqOptions.URL.String(), body)
	if err != nil {
		return nil, err
	}

	customStore := s.setStoreID(req, reqOptions.APILevel)

	if s.authContext != nil && (customStore || reqOptions.DeviceAuthNeed != deviceAuthCustomStoreOnly) {
		device, err := s.authContext.Device()
		if err != nil {
			return nil, err
		}
		// we don't have a session yet but have a serial, try
		// to get a session
		if device.SessionMacaroon == "" && device.Serial != "" {
			err = s.refreshDeviceSession(device)
			if err == auth.ErrNoSerial {
				// missing serial assertion, log and continue without device authentication
				logger.Debugf("cannot set device session: %v", err)
			}
			if err != nil && err != auth.ErrNoSerial {
				return nil, err
			}
		}
		authenticateDevice(req, device, reqOptions.APILevel)
	}

	// only set user authentication if user logged in to the store
	if user.HasStoreAuth() {
		authenticateUser(req, user)
	}

	req.Header.Set("User-Agent", httputil.UserAgent())
	req.Header.Set("Accept", reqOptions.Accept)
	req.Header.Set(hdrSnapDeviceArchitecture[reqOptions.APILevel], s.architecture)
	req.Header.Set(hdrSnapDeviceSeries[reqOptions.APILevel], s.series)
	req.Header.Set(hdrSnapClassic[reqOptions.APILevel], strconv.FormatBool(release.OnClassic))
	if reqOptions.APILevel == apiV1Endps {
		req.Header.Set("X-Ubuntu-Wire-Protocol", UbuntuCoreWireProtocol)
	}

	if reqOptions.ContentType != "" {
		req.Header.Set("Content-Type", reqOptions.ContentType)
	}

	for header, value := range reqOptions.ExtraHeaders {
		req.Header.Set(header, value)
	}

	return req, nil
}

func (s *Store) cdnHeader() (string, error) {
	if s.noCDN {
		return "none", nil
	}

	if s.authContext == nil {
		return "", nil
	}

	// set Snap-CDN from cloud instance information
	// if available

	// TODO: do we want a more complex retry strategy
	// where we first to send this header and if the
	// operation fails that way to even get the connection
	// then we retry without sending this?

	cloudInfo, err := s.authContext.CloudInfo()
	if err != nil {
		return "", err
	}

	if cloudInfo != nil {
		cdnParams := []string{fmt.Sprintf("cloud-name=%q", cloudInfo.Name)}
		if cloudInfo.Region != "" {
			cdnParams = append(cdnParams, fmt.Sprintf("region=%q", cloudInfo.Region))
		}
		if cloudInfo.AvailabilityZone != "" {
			cdnParams = append(cdnParams, fmt.Sprintf("availability-zone=%q", cloudInfo.AvailabilityZone))
		}

		return strings.Join(cdnParams, " "), nil
	}

	return "", nil
}

func (s *Store) extractSuggestedCurrency(resp *http.Response) {
	suggestedCurrency := resp.Header.Get("X-Suggested-Currency")

	if suggestedCurrency != "" {
		s.mu.Lock()
		s.suggestedCurrency = suggestedCurrency
		s.mu.Unlock()
	}
}

// ordersResult encapsulates the order data sent to us from the software center agent.
//
// {
//   "orders": [
//     {
//       "snap_id": "abcd1234efgh5678ijkl9012",
//       "currency": "USD",
//       "amount": "2.99",
//       "state": "Complete",
//       "refundable_until": null,
//       "purchase_date": "2016-09-20T15:00:00+00:00"
//     },
//     {
//       "snap_id": "abcd1234efgh5678ijkl9012",
//       "currency": null,
//       "amount": null,
//       "state": "Complete",
//       "refundable_until": null,
//       "purchase_date": "2016-09-20T15:00:00+00:00"
//     }
//   ]
// }
type ordersResult struct {
	Orders []*order `json:"orders"`
}

type order struct {
	SnapID          string `json:"snap_id"`
	Currency        string `json:"currency"`
	Amount          string `json:"amount"`
	State           string `json:"state"`
	RefundableUntil string `json:"refundable_until"`
	PurchaseDate    string `json:"purchase_date"`
}

// decorateOrders sets the MustBuy property of each snap in the given list according to the user's known orders.
func (s *Store) decorateOrders(snaps []*snap.Info, user *auth.UserState) error {
	// Mark every non-free snap as must buy until we know better.
	hasPriced := false
	for _, info := range snaps {
		if info.Paid {
			info.MustBuy = true
			hasPriced = true
		}
	}

	if user == nil {
		return nil
	}

	if !hasPriced {
		return nil
	}

	var err error

	reqOptions := &requestOptions{
		Method: "GET",
		URL:    s.endpointURL(ordersEndpPath, nil),
		Accept: jsonContentType,
	}
	var result ordersResult
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, user, &result, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode == 401 {
		// TODO handle token expiry and refresh
		return ErrInvalidCredentials
	}
	if resp.StatusCode != 200 {
		return respToError(resp, "obtain known orders from store")
	}

	// Make a map of the IDs of bought snaps
	bought := make(map[string]bool)
	for _, order := range result.Orders {
		bought[order.SnapID] = true
	}

	for _, info := range snaps {
		info.MustBuy = mustBuy(info.Paid, bought[info.SnapID])
	}

	return nil
}

// mustBuy determines if a snap requires a payment, based on if it is non-free and if the user has already bought it
func mustBuy(paid bool, bought bool) bool {
	if !paid {
		// If the snap is free, then it doesn't need buying
		return false
	}

	return !bought
}

// A SnapSpec describes a single snap wanted from SnapInfo
type SnapSpec struct {
	Name string
}

// SnapInfo returns the snap.Info for the store-hosted snap matching the given spec, or an error.
func (s *Store) SnapInfo(snapSpec SnapSpec, user *auth.UserState) (*snap.Info, error) {
	query := url.Values{}
	query.Set("fields", strings.Join(s.infoFields, ","))
	query.Set("architecture", s.architecture)

	u := s.endpointURL(path.Join(snapInfoEndpPath, snapSpec.Name), query)
	reqOptions := &requestOptions{
		Method:   "GET",
		URL:      u,
		APILevel: apiV2Endps,
	}

	var remote storeInfo
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, user, &remote, nil)
	if err != nil {
		return nil, err
	}

	// check statusCode
	switch resp.StatusCode {
	case 200:
		// OK
	case 404:
		return nil, ErrSnapNotFound
	default:
		msg := fmt.Sprintf("get details for snap %q", snapSpec.Name)
		return nil, respToError(resp, msg)
	}

	info, err := infoFromStoreInfo(&remote)
	if err != nil {
		return nil, err
	}

	err = s.decorateOrders([]*snap.Info{info}, user)
	if err != nil {
		logger.Noticef("cannot get user orders: %v", err)
	}

	s.extractSuggestedCurrency(resp)

	return info, nil
}

// A Search is what you do in order to Find something
type Search struct {
	Query   string
	Section string
	Scope   string
	Private bool
	Prefix  bool
}

// Find finds  (installable) snaps from the store, matching the
// given Search.
func (s *Store) Find(search *Search, user *auth.UserState) ([]*snap.Info, error) {
	searchTerm := search.Query

	if search.Private && user == nil {
		return nil, ErrUnauthenticated
	}

	searchTerm = strings.TrimSpace(searchTerm)

	// these characters might have special meaning on the search
	// server, and don't form part of a reasonable search, so
	// abort if they're included.
	//
	// "-" might also be special on the server, but it's also a
	// valid part of a package name, so we let it pass
	if strings.ContainsAny(searchTerm, `+=&|><!(){}[]^"~*?:\/`) {
		return nil, ErrBadQuery
	}

	q := s.defaultSnapQuery()

	if search.Private {
		if search.Prefix {
			// The store only supports "fuzzy" search for private snaps.
			// See http://search.apps.ubuntu.com/docs/
			return nil, ErrBadQuery
		}

		q.Set("private", "true")
	}

	if search.Prefix {
		q.Set("name", searchTerm)
	} else {
		q.Set("q", searchTerm)
	}
	if search.Section != "" {
		q.Set("section", search.Section)
	}
	if search.Scope != "" {
		q.Set("scope", search.Scope)
	}

	if release.OnClassic {
		q.Set("confinement", "strict,classic")
	} else {
		q.Set("confinement", "strict")
	}

	u := s.endpointURL(searchEndpPath, q)
	reqOptions := &requestOptions{
		Method: "GET",
		URL:    u,
		Accept: halJsonContentType,
	}

	var searchData searchResults
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, user, &searchData, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, respToError(resp, "search")
	}

	if ct := resp.Header.Get("Content-Type"); ct != halJsonContentType {
		return nil, fmt.Errorf("received an unexpected content type (%q) when trying to search via %q", ct, resp.Request.URL)
	}

	snaps := make([]*snap.Info, len(searchData.Payload.Packages))
	for i, pkg := range searchData.Payload.Packages {
		snaps[i] = infoFromRemote(pkg)
	}

	err = s.decorateOrders(snaps, user)
	if err != nil {
		logger.Noticef("cannot get user orders: %v", err)
	}

	s.extractSuggestedCurrency(resp)

	return snaps, nil
}

// Sections retrieves the list of available store sections.
func (s *Store) Sections(ctx context.Context, user *auth.UserState) ([]string, error) {
	reqOptions := &requestOptions{
		Method:         "GET",
		URL:            s.endpointURL(sectionsEndpPath, nil),
		Accept:         halJsonContentType,
		DeviceAuthNeed: deviceAuthCustomStoreOnly,
	}

	var sectionData sectionResults
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, user, &sectionData, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, respToError(resp, "sections")
	}

	if ct := resp.Header.Get("Content-Type"); ct != halJsonContentType {
		return nil, fmt.Errorf("received an unexpected content type (%q) when trying to retrieve the sections via %q", ct, resp.Request.URL)
	}

	var sectionNames []string
	for _, s := range sectionData.Payload.Sections {
		sectionNames = append(sectionNames, s.Name)
	}

	return sectionNames, nil
}

// WriteCatalogs queries the "commands" endpoint and writes the
// command names into the given io.Writer.
func (s *Store) WriteCatalogs(ctx context.Context, names io.Writer, adder SnapAdder) error {
	u := *s.endpointURL(commandsEndpPath, nil)

	q := u.Query()
	if release.OnClassic {
		q.Set("confinement", "strict,classic")
	} else {
		q.Set("confinement", "strict")
	}

	u.RawQuery = q.Encode()
	reqOptions := &requestOptions{
		Method:         "GET",
		URL:            &u,
		Accept:         halJsonContentType,
		DeviceAuthNeed: deviceAuthCustomStoreOnly,
	}

	// do not log body for catalog updates (its huge)
	client := httputil.NewHTTPClient(&httputil.ClientOptions{
		MayLogBody: false,
		Timeout:    10 * time.Second,
		Proxy:      s.proxy,
	})
	doRequest := func() (*http.Response, error) {
		return s.doRequest(ctx, client, reqOptions, nil)
	}
	readResponse := func(resp *http.Response) error {
		return decodeCatalog(resp, names, adder)
	}

	resp, err := httputil.RetryRequest(u.String(), doRequest, readResponse, defaultRetryStrategy)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return respToError(resp, "refresh commands catalog")
	}

	return nil
}

// RefreshCandidate contains information for the store about the currently
// installed snap so that the store can decide what update we should see
type RefreshCandidate struct {
	SnapID   string
	Revision snap.Revision
	Epoch    snap.Epoch
	Block    []snap.Revision

	// the desired channel
	Channel string
	// whether validation should be ignored
	IgnoreValidation bool

	// try to refresh a local snap to a store revision
	Amend bool
}

// the exact bits that we need to send to the store
type currentSnapJSON struct {
	SnapID           string     `json:"snap_id"`
	Channel          string     `json:"channel"`
	Revision         int        `json:"revision,omitempty"`
	Epoch            snap.Epoch `json:"epoch"`
	Confinement      string     `json:"confinement"`
	IgnoreValidation bool       `json:"ignore_validation,omitempty"`
}

func currentSnap(cs *RefreshCandidate) *currentSnapJSON {
	// the store gets confused if we send snaps without a snapid
	// (like local ones)
	if cs.SnapID == "" {
		if cs.Revision.Store() {
			logger.Noticef("store.currentSnap got given a RefreshCandidate with an empty SnapID but a store revision!")
		}
		return nil
	}
	if !cs.Revision.Store() && !cs.Amend {
		logger.Noticef("store.currentSnap got given a RefreshCandidate with a non-empty SnapID but a non-store revision!")
		return nil
	}

	channel := cs.Channel
	if channel == "" {
		channel = "stable"
	}

	return &currentSnapJSON{
		SnapID:           cs.SnapID,
		Channel:          channel,
		Epoch:            cs.Epoch,
		Revision:         cs.Revision.N,
		IgnoreValidation: cs.IgnoreValidation,
		// confinement purposely left empty
	}
}

func findRev(needle snap.Revision, haystack []snap.Revision) bool {
	for _, r := range haystack {
		if needle == r {
			return true
		}
	}
	return false
}

type HashError struct {
	name           string
	sha3_384       string
	targetSha3_384 string
}

func (e HashError) Error() string {
	return fmt.Sprintf("sha3-384 mismatch for %q: got %s but expected %s", e.name, e.sha3_384, e.targetSha3_384)
}

type DownloadOptions struct {
	RateLimit int64
}

// Download downloads the snap addressed by download info and returns its
// filename.
// The file is saved in temporary storage, and should be removed
// after use to prevent the disk from running out of space.
func (s *Store) Download(ctx context.Context, name string, targetPath string, downloadInfo *snap.DownloadInfo, pbar progress.Meter, user *auth.UserState, dlOpts *DownloadOptions) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}

	if err := s.cacher.Get(downloadInfo.Sha3_384, targetPath); err == nil {
		logger.Debugf("Cache hit for SHA3_384 …%.5s.", downloadInfo.Sha3_384)
		return nil
	}

	if useDeltas() {
		logger.Debugf("Available deltas returned by store: %v", downloadInfo.Deltas)

		if len(downloadInfo.Deltas) == 1 {
			err := s.downloadAndApplyDelta(name, targetPath, downloadInfo, pbar, user)
			if err == nil {
				return nil
			}
			// We revert to normal downloads if there is any error.
			logger.Noticef("Cannot download or apply deltas for %s: %v", name, err)
		}
	}

	partialPath := targetPath + ".partial"
	w, err := os.OpenFile(partialPath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	resume, err := w.Seek(0, os.SEEK_END)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = cerr
		}
		if err != nil {
			os.Remove(w.Name())
		}
	}()
	if resume > 0 {
		logger.Debugf("Resuming download of %q at %d.", partialPath, resume)
	} else {
		logger.Debugf("Starting download of %q.", partialPath)
	}

	authAvail, err := s.authAvailable(user)
	if err != nil {
		return err
	}

	url := downloadInfo.AnonDownloadURL
	if url == "" || authAvail {
		url = downloadInfo.DownloadURL
	}

	if downloadInfo.Size == 0 || resume < downloadInfo.Size {
		err = download(ctx, name, downloadInfo.Sha3_384, url, user, s, w, resume, pbar, dlOpts)
		if err != nil {
			logger.Debugf("download of %q failed: %#v", url, err)
		}
	} else {
		// we're done! check the hash though
		h := crypto.SHA3_384.New()
		if _, err := w.Seek(0, os.SEEK_SET); err != nil {
			return err
		}
		if _, err := io.Copy(h, w); err != nil {
			return err
		}
		actualSha3 := fmt.Sprintf("%x", h.Sum(nil))
		if downloadInfo.Sha3_384 != actualSha3 {
			err = HashError{name, actualSha3, downloadInfo.Sha3_384}
		}
	}
	// If hashsum is incorrect retry once
	if _, ok := err.(HashError); ok {
		logger.Debugf("Hashsum error on download: %v", err.Error())
		logger.Debugf("Truncating and trying again from scratch.")
		err = w.Truncate(0)
		if err != nil {
			return err
		}
		_, err = w.Seek(0, os.SEEK_SET)
		if err != nil {
			return err
		}
		err = download(ctx, name, downloadInfo.Sha3_384, url, user, s, w, 0, pbar, nil)
		if err != nil {
			logger.Debugf("download of %q failed: %#v", url, err)
		}
	}

	if err != nil {
		return err
	}

	if err := os.Rename(w.Name(), targetPath); err != nil {
		return err
	}

	if err := w.Sync(); err != nil {
		return err
	}

	return s.cacher.Put(downloadInfo.Sha3_384, targetPath)
}

func reqOptions(storeURL *url.URL, cdnHeader string) *requestOptions {
	reqOptions := requestOptions{
		Method:       "GET",
		URL:          storeURL,
		ExtraHeaders: map[string]string{},
	}
	if cdnHeader != "" {
		reqOptions.ExtraHeaders["Snap-CDN"] = cdnHeader
	}

	return &reqOptions
}

var ratelimitReader = ratelimit.Reader

var download = downloadImpl

// download writes an http.Request showing a progress.Meter
func downloadImpl(ctx context.Context, name, sha3_384, downloadURL string, user *auth.UserState, s *Store, w io.ReadWriteSeeker, resume int64, pbar progress.Meter, dlOpts *DownloadOptions) error {
	if dlOpts == nil {
		dlOpts = &DownloadOptions{}
	}

	storeURL, err := url.Parse(downloadURL)
	if err != nil {
		return err
	}

	cdnHeader, err := s.cdnHeader()
	if err != nil {
		return err
	}

	var finalErr error
	var dlSize float64
	startTime := time.Now()
	for attempt := retry.Start(defaultRetryStrategy, nil); attempt.Next(); {
		reqOptions := reqOptions(storeURL, cdnHeader)

		httputil.MaybeLogRetryAttempt(reqOptions.URL.String(), attempt, startTime)

		h := crypto.SHA3_384.New()

		if resume > 0 {
			reqOptions.ExtraHeaders["Range"] = fmt.Sprintf("bytes=%d-", resume)
			// seed the sha3 with the already local file
			if _, err := w.Seek(0, os.SEEK_SET); err != nil {
				return err
			}
			n, err := io.Copy(h, w)
			if err != nil {
				return err
			}
			if n != resume {
				return fmt.Errorf("resume offset wrong: %d != %d", resume, n)
			}
		}

		if cancelled(ctx) {
			return fmt.Errorf("The download has been cancelled: %s", ctx.Err())
		}
		var resp *http.Response
		resp, finalErr = s.doRequest(ctx, httputil.NewHTTPClient(&httputil.ClientOptions{Proxy: s.proxy}), reqOptions, user)

		if cancelled(ctx) {
			return fmt.Errorf("The download has been cancelled: %s", ctx.Err())
		}
		if finalErr != nil {
			if httputil.ShouldRetryError(attempt, finalErr) {
				continue
			}
			break
		}

		if httputil.ShouldRetryHttpResponse(attempt, resp) {
			resp.Body.Close()
			continue
		}

		defer resp.Body.Close()

		switch resp.StatusCode {
		case 200, 206: // OK, Partial Content
		case 402: // Payment Required

			return fmt.Errorf("please buy %s before installing it.", name)
		default:
			return &DownloadError{Code: resp.StatusCode, URL: resp.Request.URL}
		}

		if pbar == nil {
			pbar = progress.Null
		}
		dlSize = float64(resp.ContentLength)
		pbar.Start(name, dlSize)
		mw := io.MultiWriter(w, h, pbar)
		var limiter io.Reader
		limiter = resp.Body
		if limit := dlOpts.RateLimit; limit > 0 {
			bucket := ratelimit.NewBucketWithRate(float64(limit), 2*limit)
			limiter = ratelimitReader(resp.Body, bucket)
		}
		_, finalErr = io.Copy(mw, limiter)
		pbar.Finished()
		if finalErr != nil {
			if httputil.ShouldRetryError(attempt, finalErr) {
				// error while downloading should resume
				var seekerr error
				resume, seekerr = w.Seek(0, os.SEEK_END)
				if seekerr == nil {
					continue
				}
				// if seek failed, then don't retry end return the original error
			}
			break
		}

		if cancelled(ctx) {
			return fmt.Errorf("The download has been cancelled: %s", ctx.Err())
		}

		actualSha3 := fmt.Sprintf("%x", h.Sum(nil))
		if sha3_384 != "" && sha3_384 != actualSha3 {
			finalErr = HashError{name, actualSha3, sha3_384}
		}
		break
	}
	if finalErr == nil {
		// not using quantity.FormatFoo as this is just for debug
		dt := time.Since(startTime)
		r := dlSize / dt.Seconds()
		var p rune
		for _, p = range " kMGTPEZY" {
			if r < 1000 {
				break
			}
			r /= 1000
		}

		logger.Debugf("Download succeeded in %.03fs (%.0f%cB/s).", dt.Seconds(), r, p)
	}
	return finalErr
}

// downloadDelta downloads the delta for the preferred format, returning the path.
func (s *Store) downloadDelta(deltaName string, downloadInfo *snap.DownloadInfo, w io.ReadWriteSeeker, pbar progress.Meter, user *auth.UserState) error {

	if len(downloadInfo.Deltas) != 1 {
		return errors.New("store returned more than one download delta")
	}

	deltaInfo := downloadInfo.Deltas[0]

	if deltaInfo.Format != s.deltaFormat {
		return fmt.Errorf("store returned unsupported delta format %q (only xdelta3 currently)", deltaInfo.Format)
	}

	authAvail, err := s.authAvailable(user)
	if err != nil {
		return err
	}

	url := deltaInfo.AnonDownloadURL
	if url == "" || authAvail {
		url = deltaInfo.DownloadURL
	}

	return download(context.TODO(), deltaName, deltaInfo.Sha3_384, url, user, s, w, 0, pbar, nil)
}

func getXdelta3Cmd(args ...string) (*exec.Cmd, error) {
	switch {
	case osutil.ExecutableExists("xdelta3"):
		return exec.Command("xdelta3", args...), nil
	case osutil.FileExists(filepath.Join(dirs.SnapMountDir, "/core/current/usr/bin/xdelta3")):
		return osutil.CommandFromCore("/usr/bin/xdelta3", args...)
	}
	return nil, fmt.Errorf("cannot find xdelta3 binary in PATH or core snap")
}

// applyDelta generates a target snap from a previously downloaded snap and a downloaded delta.
var applyDelta = func(name string, deltaPath string, deltaInfo *snap.DeltaInfo, targetPath string, targetSha3_384 string) error {
	snapBase := fmt.Sprintf("%s_%d.snap", name, deltaInfo.FromRevision)
	snapPath := filepath.Join(dirs.SnapBlobDir, snapBase)

	if !osutil.FileExists(snapPath) {
		return fmt.Errorf("snap %q revision %d not found at %s", name, deltaInfo.FromRevision, snapPath)
	}

	if deltaInfo.Format != "xdelta3" {
		return fmt.Errorf("cannot apply unsupported delta format %q (only xdelta3 currently)", deltaInfo.Format)
	}

	partialTargetPath := targetPath + ".partial"

	xdelta3Args := []string{"-d", "-s", snapPath, deltaPath, partialTargetPath}
	cmd, err := getXdelta3Cmd(xdelta3Args...)
	if err != nil {
		return err
	}

	if err := cmd.Run(); err != nil {
		if err := os.Remove(partialTargetPath); err != nil {
			logger.Noticef("failed to remove partial delta target %q: %s", partialTargetPath, err)
		}
		return err
	}

	bsha3_384, _, err := osutil.FileDigest(partialTargetPath, crypto.SHA3_384)
	if err != nil {
		return err
	}
	sha3_384 := fmt.Sprintf("%x", bsha3_384)
	if targetSha3_384 != "" && sha3_384 != targetSha3_384 {
		if err := os.Remove(partialTargetPath); err != nil {
			logger.Noticef("failed to remove partial delta target %q: %s", partialTargetPath, err)
		}
		return HashError{name, sha3_384, targetSha3_384}
	}

	if err := os.Rename(partialTargetPath, targetPath); err != nil {
		return osutil.CopyFile(partialTargetPath, targetPath, 0)
	}

	return nil
}

// downloadAndApplyDelta downloads and then applies the delta to the current snap.
func (s *Store) downloadAndApplyDelta(name, targetPath string, downloadInfo *snap.DownloadInfo, pbar progress.Meter, user *auth.UserState) error {
	deltaInfo := &downloadInfo.Deltas[0]

	deltaPath := fmt.Sprintf("%s.%s-%d-to-%d.partial", targetPath, deltaInfo.Format, deltaInfo.FromRevision, deltaInfo.ToRevision)
	deltaName := fmt.Sprintf(i18n.G("%s (delta)"), name)

	w, err := os.OpenFile(deltaPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := w.Close(); cerr != nil && err == nil {
			err = cerr
		}
		os.Remove(deltaPath)
	}()

	err = s.downloadDelta(deltaName, downloadInfo, w, pbar, user)
	if err != nil {
		return err
	}

	logger.Debugf("Successfully downloaded delta for %q at %s", name, deltaPath)
	if err := applyDelta(name, deltaPath, deltaInfo, targetPath, downloadInfo.Sha3_384); err != nil {
		return err
	}

	logger.Debugf("Successfully applied delta for %q at %s, saving %d bytes.", name, deltaPath, downloadInfo.Size-deltaInfo.Size)
	return nil
}

type assertionSvcError struct {
	Status int    `json:"status"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// Assertion retrivies the assertion for the given type and primary key.
func (s *Store) Assertion(assertType *asserts.AssertionType, primaryKey []string, user *auth.UserState) (asserts.Assertion, error) {
	v := url.Values{}
	v.Set("max-format", strconv.Itoa(assertType.MaxSupportedFormat()))
	u := s.assertionsEndpointURL(path.Join(assertType.Name, path.Join(primaryKey...)), v)

	reqOptions := &requestOptions{
		Method: "GET",
		URL:    u,
		Accept: asserts.MediaType,
	}

	var asrt asserts.Assertion

	resp, err := httputil.RetryRequest(reqOptions.URL.String(), func() (*http.Response, error) {
		return s.doRequest(context.TODO(), s.client, reqOptions, user)
	}, func(resp *http.Response) error {
		var e error
		if resp.StatusCode == 200 {
			// decode assertion
			dec := asserts.NewDecoder(resp.Body)
			asrt, e = dec.Decode()
		} else {
			contentType := resp.Header.Get("Content-Type")
			if contentType == jsonContentType || contentType == "application/problem+json" {
				var svcErr assertionSvcError
				dec := json.NewDecoder(resp.Body)
				if e = dec.Decode(&svcErr); e != nil {
					return fmt.Errorf("cannot decode assertion service error with HTTP status code %d: %v", resp.StatusCode, e)
				}
				if svcErr.Status == 404 {
					// best-effort
					headers, _ := asserts.HeadersFromPrimaryKey(assertType, primaryKey)
					return &asserts.NotFoundError{
						Type:    assertType,
						Headers: headers,
					}
				}
				return fmt.Errorf("assertion service error: [%s] %q", svcErr.Title, svcErr.Detail)
			}
		}
		return e
	}, defaultRetryStrategy)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, respToError(resp, "fetch assertion")
	}

	return asrt, err
}

// SuggestedCurrency retrieves the cached value for the store's suggested currency
func (s *Store) SuggestedCurrency() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.suggestedCurrency == "" {
		return "USD"
	}
	return s.suggestedCurrency
}

// BuyOptions specifies parameters to buy from the store.
type BuyOptions struct {
	SnapID   string  `json:"snap-id"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"` // ISO 4217 code as string
}

// BuyResult holds the state of a buy attempt.
type BuyResult struct {
	State string `json:"state,omitempty"`
}

// orderInstruction holds data sent to the store for orders.
type orderInstruction struct {
	SnapID   string `json:"snap_id"`
	Amount   string `json:"amount,omitempty"`
	Currency string `json:"currency,omitempty"`
}

type storeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (s *storeError) Error() string {
	return s.Message
}

type storeErrors struct {
	Errors []*storeError `json:"error_list"`
}

func (s *storeErrors) Code() string {
	if len(s.Errors) == 0 {
		return ""
	}
	return s.Errors[0].Code
}

func (s *storeErrors) Error() string {
	if len(s.Errors) == 0 {
		return "internal error: empty store error used as an actual error"
	}
	return s.Errors[0].Error()
}

func buyOptionError(message string) (*BuyResult, error) {
	return nil, fmt.Errorf("cannot buy snap: %s", message)
}

// Buy sends a buy request for the specified snap.
// Returns the state of the order: Complete, Cancelled.
func (s *Store) Buy(options *BuyOptions, user *auth.UserState) (*BuyResult, error) {
	if options.SnapID == "" {
		return buyOptionError("snap ID missing")
	}
	if options.Price <= 0 {
		return buyOptionError("invalid expected price")
	}
	if options.Currency == "" {
		return buyOptionError("currency missing")
	}
	if user == nil {
		return nil, ErrUnauthenticated
	}

	instruction := orderInstruction{
		SnapID:   options.SnapID,
		Amount:   fmt.Sprintf("%.2f", options.Price),
		Currency: options.Currency,
	}

	jsonData, err := json.Marshal(instruction)
	if err != nil {
		return nil, err
	}

	reqOptions := &requestOptions{
		Method:      "POST",
		URL:         s.endpointURL(buyEndpPath, nil),
		Accept:      jsonContentType,
		ContentType: jsonContentType,
		Data:        jsonData,
	}

	var orderDetails order
	var errorInfo storeErrors
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, user, &orderDetails, &errorInfo)
	if err != nil {
		return nil, err
	}

	switch resp.StatusCode {
	case 200, 201:
		// user already ordered or order successful
		if orderDetails.State == "Cancelled" {
			return buyOptionError("payment cancelled")
		}

		return &BuyResult{
			State: orderDetails.State,
		}, nil
	case 400:
		// Invalid price was specified, etc.
		return buyOptionError(fmt.Sprintf("bad request: %v", errorInfo.Error()))
	case 403:
		// Customer account not set up for purchases.
		switch errorInfo.Code() {
		case "no-payment-methods":
			return nil, ErrNoPaymentMethods
		case "tos-not-accepted":
			return nil, ErrTOSNotAccepted
		}
		return buyOptionError(fmt.Sprintf("permission denied: %v", errorInfo.Error()))
	case 404:
		// Likely because customer account or snap ID doesn't exist.
		return buyOptionError(fmt.Sprintf("server says not found: %v", errorInfo.Error()))
	case 402: // Payment Required
		// Payment failed for some reason.
		return nil, ErrPaymentDeclined
	case 401:
		// TODO handle token expiry and refresh
		return nil, ErrInvalidCredentials
	default:
		return nil, respToError(resp, fmt.Sprintf("buy snap: %v", errorInfo))
	}
}

type storeCustomer struct {
	LatestTOSDate     string `json:"latest_tos_date"`
	AcceptedTOSDate   string `json:"accepted_tos_date"`
	LatestTOSAccepted bool   `json:"latest_tos_accepted"`
	HasPaymentMethod  bool   `json:"has_payment_method"`
}

// ReadyToBuy returns nil if the user's account has accepted T&Cs and has a payment method registered, and an error otherwise
func (s *Store) ReadyToBuy(user *auth.UserState) error {
	if user == nil {
		return ErrUnauthenticated
	}

	reqOptions := &requestOptions{
		Method: "GET",
		URL:    s.endpointURL(customersMeEndpPath, nil),
		Accept: jsonContentType,
	}

	var customer storeCustomer
	var errors storeErrors
	resp, err := s.retryRequestDecodeJSON(context.TODO(), reqOptions, user, &customer, &errors)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		if !customer.HasPaymentMethod {
			return ErrNoPaymentMethods
		}
		if !customer.LatestTOSAccepted {
			return ErrTOSNotAccepted
		}
		return nil
	case 404:
		// Likely because user has no account registered on the pay server
		return fmt.Errorf("cannot get customer details: server says no account exists")
	case 401:
		return ErrInvalidCredentials
	default:
		if len(errors.Errors) == 0 {
			return fmt.Errorf("cannot get customer details: unexpected HTTP code %d", resp.StatusCode)
		}
		return &errors
	}
}

func (s *Store) CacheDownloads() int {
	return s.cfg.CacheDownloads
}

func (s *Store) SetCacheDownloads(fileCount int) {
	s.cfg.CacheDownloads = fileCount
	if fileCount > 0 {
		s.cacher = NewCacheManager(dirs.SnapDownloadCacheDir, fileCount)
	} else {
		s.cacher = &nullCache{}
	}
}

// snap action: install/refresh

type CurrentSnap struct {
	InstanceName     string
	SnapID           string
	Revision         snap.Revision
	TrackingChannel  string
	RefreshedDate    time.Time
	IgnoreValidation bool
	Block            []snap.Revision
}

type currentSnapV2JSON struct {
	SnapID           string     `json:"snap-id"`
	InstanceKey      string     `json:"instance-key"`
	Revision         int        `json:"revision"`
	TrackingChannel  string     `json:"tracking-channel"`
	RefreshedDate    *time.Time `json:"refreshed-date,omitempty"`
	IgnoreValidation bool       `json:"ignore-validation,omitempty"`
}

type SnapActionFlags int

const (
	SnapActionIgnoreValidation SnapActionFlags = 1 << iota
	SnapActionEnforceValidation
)

type SnapAction struct {
	Action       string
	InstanceName string
	SnapID       string
	Channel      string
	Revision     snap.Revision
	Flags        SnapActionFlags
	Epoch        *snap.Epoch
}

func isValidAction(action string) bool {
	switch action {
	case "download", "install", "refresh":
		return true
	default:
		return false
	}
}

type snapActionJSON struct {
	Action           string      `json:"action"`
	InstanceKey      string      `json:"instance-key"`
	Name             string      `json:"name,omitempty"`
	SnapID           string      `json:"snap-id,omitempty"`
	Channel          string      `json:"channel,omitempty"`
	Revision         int         `json:"revision,omitempty"`
	Epoch            *snap.Epoch `json:"epoch,omitempty"`
	IgnoreValidation *bool       `json:"ignore-validation,omitempty"`
}

type snapRelease struct {
	Architecture string `json:"architecture"`
	Channel      string `json:"channel"`
}

type snapActionResult struct {
	Result           string    `json:"result"`
	InstanceKey      string    `json:"instance-key"`
	SnapID           string    `json:"snap-id,omitempy"`
	Name             string    `json:"name,omitempty"`
	Snap             storeSnap `json:"snap"`
	EffectiveChannel string    `json:"effective-channel,omitempty"`
	Error            struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		Extra   struct {
			Releases []snapRelease `json:"releases"`
		} `json:"extra"`
	} `json:"error"`
}

type snapActionRequest struct {
	Context []*currentSnapV2JSON `json:"context"`
	Actions []*snapActionJSON    `json:"actions"`
	Fields  []string             `json:"fields"`
}

type snapActionResultList struct {
	Results   []*snapActionResult `json:"results"`
	ErrorList []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error-list"`
}

var snapActionFields = jsonutil.StructFields((*storeSnap)(nil))

// SnapAction queries the store for snap information for the given
// install/refresh actions, given the context information about
// current installed snaps in currentSnaps. If the request was overall
// successul (200) but there were reported errors it will return both
// the snap infos and an SnapActionError.
func (s *Store) SnapAction(ctx context.Context, currentSnaps []*CurrentSnap, actions []*SnapAction, user *auth.UserState, opts *RefreshOptions) ([]*snap.Info, error) {
	if opts == nil {
		opts = &RefreshOptions{}
	}

	if len(currentSnaps) == 0 && len(actions) == 0 {
		// nothing to do
		return nil, &SnapActionError{NoResults: true}
	}

	authRefreshes := 0
	for {
		snaps, err := s.snapAction(ctx, currentSnaps, actions, user, opts)

		if saErr, ok := err.(*SnapActionError); ok && authRefreshes < 2 && len(saErr.Other) > 0 {
			// do we need to try to refresh auths?, 2 tries
			var refreshNeed authRefreshNeed
			for _, otherErr := range saErr.Other {
				switch otherErr {
				case errUserAuthorizationNeedsRefresh:
					refreshNeed.user = true
				case errDeviceAuthorizationNeedsRefresh:
					refreshNeed.device = true
				}
			}
			if refreshNeed.needed() {
				err := s.refreshAuth(user, refreshNeed)
				if err != nil {
					// best effort
					logger.Noticef("cannot refresh soft-expired authorisation: %v", err)
				}
				authRefreshes++
				// TODO: we could avoid retrying here
				// if refreshAuth gave no error we got
				// as many non-error results from the
				// store as actions anyway
				continue
			}
		}

		return snaps, err
	}
}

func genInstanceKey(curSnap *CurrentSnap, salt string) (string, error) {
	_, snapInstanceKey := snap.SplitInstanceName(curSnap.InstanceName)

	if snapInstanceKey == "" {
		return curSnap.SnapID, nil
	}

	if salt == "" {
		return "", fmt.Errorf("internal error: request salt not provided")
	}

	// due to privacy concerns, avoid sending the local names to the
	// backend, instead hash the snap ID and instance key together
	h := crypto.SHA256.New()
	h.Write([]byte(curSnap.SnapID))
	h.Write([]byte(snapInstanceKey))
	h.Write([]byte(salt))
	enc := base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s:%s", curSnap.SnapID, enc), nil
}

func (s *Store) snapAction(ctx context.Context, currentSnaps []*CurrentSnap, actions []*SnapAction, user *auth.UserState, opts *RefreshOptions) ([]*snap.Info, error) {

	// TODO: the store already requires instance-key but doesn't
	// yet support repeating in context or sending actions for the
	// same snap-id, for now we keep instance-key handling internal

	requestSalt := ""
	if opts != nil {
		requestSalt = opts.PrivacyKey
	}
	curSnaps := make(map[string]*CurrentSnap, len(currentSnaps))
	curSnapJSONs := make([]*currentSnapV2JSON, len(currentSnaps))
	instanceNameToKey := make(map[string]string, len(currentSnaps))
	for i, curSnap := range currentSnaps {
		if curSnap.SnapID == "" || curSnap.InstanceName == "" || curSnap.Revision.Unset() {
			return nil, fmt.Errorf("internal error: invalid current snap information")
		}
		instanceKey, err := genInstanceKey(curSnap, requestSalt)
		if err != nil {
			return nil, err
		}
		curSnaps[instanceKey] = curSnap
		instanceNameToKey[curSnap.InstanceName] = instanceKey

		channel := curSnap.TrackingChannel
		if channel == "" {
			channel = "stable"
		}
		var refreshedDate *time.Time
		if !curSnap.RefreshedDate.IsZero() {
			refreshedDate = &curSnap.RefreshedDate
		}
		curSnapJSONs[i] = &currentSnapV2JSON{
			SnapID:           curSnap.SnapID,
			InstanceKey:      instanceKey,
			Revision:         curSnap.Revision.N,
			TrackingChannel:  channel,
			IgnoreValidation: curSnap.IgnoreValidation,
			RefreshedDate:    refreshedDate,
		}
	}

	downloadNum := 0
	installNum := 0
	installs := make(map[string]*SnapAction, len(actions))
	downloads := make(map[string]*SnapAction, len(actions))
	refreshes := make(map[string]*SnapAction, len(actions))
	actionJSONs := make([]*snapActionJSON, len(actions))
	for i, a := range actions {
		if !isValidAction(a.Action) {
			return nil, fmt.Errorf("internal error: unsupported action %q", a.Action)
		}
		if a.InstanceName == "" {
			return nil, fmt.Errorf("internal error: action without instance name")
		}
		var ignoreValidation *bool
		if a.Flags&SnapActionIgnoreValidation != 0 {
			var t = true
			ignoreValidation = &t
		} else if a.Flags&SnapActionEnforceValidation != 0 {
			var f = false
			ignoreValidation = &f
		}

		var instanceKey string
		aJSON := &snapActionJSON{
			Action:           a.Action,
			SnapID:           a.SnapID,
			Channel:          a.Channel,
			Revision:         a.Revision.N,
			Epoch:            a.Epoch,
			IgnoreValidation: ignoreValidation,
		}
		if !a.Revision.Unset() {
			a.Channel = ""
		}
		if a.Action == "install" {
			installNum++
			instanceKey = fmt.Sprintf("install-%d", installNum)
			installs[instanceKey] = a
		} else if a.Action == "download" {
			downloadNum++
			instanceKey = fmt.Sprintf("download-%d", downloadNum)
			downloads[instanceKey] = a
			if _, key := snap.SplitInstanceName(a.InstanceName); key != "" {
				return nil, fmt.Errorf("internal error: unsupported download with instance name %q", a.InstanceName)
			}
		} else {
			instanceKey = instanceNameToKey[a.InstanceName]
			refreshes[instanceKey] = a
		}

		if a.Action != "refresh" {
			aJSON.Name = snap.InstanceSnap(a.InstanceName)
		}

		aJSON.InstanceKey = instanceKey

		actionJSONs[i] = aJSON
	}

	// build input for the install/refresh endpoint
	jsonData, err := json.Marshal(snapActionRequest{
		Context: curSnapJSONs,
		Actions: actionJSONs,
		Fields:  snapActionFields,
	})
	if err != nil {
		return nil, err
	}

	reqOptions := &requestOptions{
		Method:      "POST",
		URL:         s.endpointURL(snapActionEndpPath, nil),
		Accept:      jsonContentType,
		ContentType: jsonContentType,
		Data:        jsonData,
		APILevel:    apiV2Endps,
	}

	if useDeltas() {
		logger.Debugf("Deltas enabled. Adding header Snap-Accept-Delta-Format: %v", s.deltaFormat)
		reqOptions.addHeader("Snap-Accept-Delta-Format", s.deltaFormat)
	}
	if opts.RefreshManaged {
		reqOptions.addHeader("Snap-Refresh-Managed", "true")
	}

	var results snapActionResultList
	resp, err := s.retryRequestDecodeJSON(ctx, reqOptions, user, &results, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, respToError(resp, "query the store for updates")
	}

	s.extractSuggestedCurrency(resp)

	refreshErrors := make(map[string]error)
	installErrors := make(map[string]error)
	downloadErrors := make(map[string]error)
	var otherErrors []error

	var snaps []*snap.Info
	for _, res := range results.Results {
		if res.Result == "error" {
			if a := installs[res.InstanceKey]; a != nil {
				if res.Name != "" {
					installErrors[a.InstanceName] = translateSnapActionError("install", a.Channel, res.Error.Code, res.Error.Message, res.Error.Extra.Releases)
					continue
				}
			} else if a := downloads[res.InstanceKey]; a != nil {
				if res.Name != "" {
					downloadErrors[res.Name] = translateSnapActionError("download", a.Channel, res.Error.Code, res.Error.Message, res.Error.Extra.Releases)
					continue
				}
			} else {
				if cur := curSnaps[res.InstanceKey]; cur != nil {
					a := refreshes[res.InstanceKey]
					if a == nil {
						// got an error for a snap that was not part of an 'action'
						otherErrors = append(otherErrors, translateSnapActionError("", "", res.Error.Code, fmt.Sprintf("snap %q: %s", cur.InstanceName, res.Error.Message), nil))
						logger.Debugf("Unexpected error for snap %q, instance key %v: [%v] %v", cur.InstanceName, res.InstanceKey, res.Error.Code, res.Error.Message)
						continue
					}
					channel := a.Channel
					if channel == "" && a.Revision.Unset() {
						channel = cur.TrackingChannel
					}
					refreshErrors[cur.InstanceName] = translateSnapActionError("refresh", channel, res.Error.Code, res.Error.Message, res.Error.Extra.Releases)
					continue
				}
			}
			otherErrors = append(otherErrors, translateSnapActionError("", "", res.Error.Code, res.Error.Message, nil))
			continue
		}
		snapInfo, err := infoFromStoreSnap(&res.Snap)
		if err != nil {
			return nil, fmt.Errorf("unexpected invalid install/refresh API result: %v", err)
		}

		snapInfo.Channel = res.EffectiveChannel

		var instanceName string
		if res.Result == "refresh" {
			cur := curSnaps[res.InstanceKey]
			if cur == nil {
				return nil, fmt.Errorf("unexpected invalid install/refresh API result: unexpected refresh")
			}
			rrev := snap.R(res.Snap.Revision)
			if rrev == cur.Revision || findRev(rrev, cur.Block) {
				refreshErrors[cur.InstanceName] = ErrNoUpdateAvailable
				continue
			}
			instanceName = cur.InstanceName
		} else if res.Result == "install" {
			if action := installs[res.InstanceKey]; action != nil {
				instanceName = action.InstanceName
			}
		}

		if res.Result != "download" && instanceName == "" {
			return nil, fmt.Errorf("unexpected invalid install/refresh API result: unexpected instance-key %q", res.InstanceKey)
		}

		_, instanceKey := snap.SplitInstanceName(instanceName)
		snapInfo.InstanceKey = instanceKey

		snaps = append(snaps, snapInfo)
	}

	for _, errObj := range results.ErrorList {
		otherErrors = append(otherErrors, translateSnapActionError("", "", errObj.Code, errObj.Message, nil))
	}

	if len(refreshErrors)+len(installErrors)+len(downloadErrors) != 0 || len(results.Results) == 0 || len(otherErrors) != 0 {
		// normalize empty maps
		if len(refreshErrors) == 0 {
			refreshErrors = nil
		}
		if len(installErrors) == 0 {
			installErrors = nil
		}
		if len(downloadErrors) == 0 {
			downloadErrors = nil
		}
		return snaps, &SnapActionError{
			NoResults: len(results.Results) == 0,
			Refresh:   refreshErrors,
			Install:   installErrors,
			Download:  downloadErrors,
			Other:     otherErrors,
		}
	}

	return snaps, nil
}

// abbreviated info structs just for the download info
type storeInfoChannelAbbrev struct {
	Download storeSnapDownload `json:"download"`
}

type storeInfoAbbrev struct {
	// discard anything beyond the first entry
	ChannelMap [1]storeInfoChannelAbbrev `json:"channel-map"`
}

var errUnexpectedConnCheckResponse = errors.New("unexpected response during connection check")

func (s *Store) snapConnCheck() ([]string, error) {
	var hosts []string
	// NOTE: "core" is possibly the only snap that's sure to be in all stores
	//       when we drop "core" in the move to snapd/core18/etc, change this
	infoURL := s.endpointURL(path.Join(snapInfoEndpPath, "core"), url.Values{
		// we only want the download URL
		"fields": {"download"},
		// we only need *one* (but can't filter by channel ... yet)
		"architecture": {s.architecture},
	})
	hosts = append(hosts, infoURL.Host)

	var result storeInfoAbbrev
	resp, err := httputil.RetryRequest(infoURL.String(), func() (*http.Response, error) {
		return s.doRequest(context.TODO(), s.client, &requestOptions{
			Method:   "GET",
			URL:      infoURL,
			APILevel: apiV2Endps,
		}, nil)
	}, func(resp *http.Response) error {
		return decodeJSONBody(resp, &result, nil)
	}, connCheckStrategy)

	if err != nil {
		return hosts, err
	}
	resp.Body.Close()

	dlURLraw := result.ChannelMap[0].Download.URL
	dlURL, err := url.ParseRequestURI(dlURLraw)
	if err != nil {
		return hosts, err
	}
	hosts = append(hosts, dlURL.Host)

	cdnHeader, err := s.cdnHeader()
	if err != nil {
		return hosts, err
	}

	reqOptions := reqOptions(dlURL, cdnHeader)
	reqOptions.Method = "HEAD" // not actually a download

	// TODO: We need the HEAD here so that we get redirected to the
	//       right CDN machine. Consider just doing a "net.Dial"
	//       after the redirect here. Suggested in
	// https://github.com/snapcore/snapd/pull/5176#discussion_r193437230
	resp, err = httputil.RetryRequest(dlURLraw, func() (*http.Response, error) {
		return s.doRequest(context.TODO(), s.client, reqOptions, nil)
	}, func(resp *http.Response) error {
		// account for redirect
		hosts[len(hosts)-1] = resp.Request.URL.Host
		return nil
	}, connCheckStrategy)
	if err != nil {
		return hosts, err
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return hosts, errUnexpectedConnCheckResponse
	}

	return hosts, nil
}

func (s *Store) ConnectivityCheck() (status map[string]bool, err error) {
	status = make(map[string]bool)

	checkers := []func() ([]string, error){
		s.snapConnCheck,
	}

	for _, checker := range checkers {
		hosts, err := checker()
		for _, host := range hosts {
			status[host] = (err == nil)
		}
	}

	return status, nil
}
