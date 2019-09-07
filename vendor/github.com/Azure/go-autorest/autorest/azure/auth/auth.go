package auth

// Copyright 2017 Microsoft Corporation
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"unicode/utf16"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"github.com/dimchansky/utfbom"
	"golang.org/x/crypto/pkcs12"
)

// The possible keys in the Values map.
const (
	SubscriptionID          = "AZURE_SUBSCRIPTION_ID"
	TenantID                = "AZURE_TENANT_ID"
	AuxiliaryTenantIDs      = "AZURE_AUXILIARY_TENANT_IDS"
	ClientID                = "AZURE_CLIENT_ID"
	ClientSecret            = "AZURE_CLIENT_SECRET"
	CertificatePath         = "AZURE_CERTIFICATE_PATH"
	CertificatePassword     = "AZURE_CERTIFICATE_PASSWORD"
	Username                = "AZURE_USERNAME"
	Password                = "AZURE_PASSWORD"
	EnvironmentName         = "AZURE_ENVIRONMENT"
	Resource                = "AZURE_AD_RESOURCE"
	ActiveDirectoryEndpoint = "ActiveDirectoryEndpoint"
	ResourceManagerEndpoint = "ResourceManagerEndpoint"
	GraphResourceID         = "GraphResourceID"
	SQLManagementEndpoint   = "SQLManagementEndpoint"
	GalleryEndpoint         = "GalleryEndpoint"
	ManagementEndpoint      = "ManagementEndpoint"
)

// NewAuthorizerFromEnvironment creates an Authorizer configured from environment variables in the order:
// 1. Client credentials
// 2. Client certificate
// 3. Username password
// 4. MSI
func NewAuthorizerFromEnvironment() (autorest.Authorizer, error) {
	settings, err := GetSettingsFromEnvironment()
	if err != nil {
		return nil, err
	}
	return settings.GetAuthorizer()
}

// NewAuthorizerFromEnvironmentWithResource creates an Authorizer configured from environment variables in the order:
// 1. Client credentials
// 2. Client certificate
// 3. Username password
// 4. MSI
func NewAuthorizerFromEnvironmentWithResource(resource string) (autorest.Authorizer, error) {
	settings, err := GetSettingsFromEnvironment()
	if err != nil {
		return nil, err
	}
	settings.Values[Resource] = resource
	return settings.GetAuthorizer()
}

// EnvironmentSettings contains the available authentication settings.
type EnvironmentSettings struct {
	Values      map[string]string
	Environment azure.Environment
}

// GetSettingsFromEnvironment returns the available authentication settings from the environment.
func GetSettingsFromEnvironment() (s EnvironmentSettings, err error) {
	s = EnvironmentSettings{
		Values: map[string]string{},
	}
	s.setValue(SubscriptionID)
	s.setValue(TenantID)
	s.setValue(AuxiliaryTenantIDs)
	s.setValue(ClientID)
	s.setValue(ClientSecret)
	s.setValue(CertificatePath)
	s.setValue(CertificatePassword)
	s.setValue(Username)
	s.setValue(Password)
	s.setValue(EnvironmentName)
	s.setValue(Resource)
	if v := s.Values[EnvironmentName]; v == "" {
		s.Environment = azure.PublicCloud
	} else {
		s.Environment, err = azure.EnvironmentFromName(v)
	}
	if s.Values[Resource] == "" {
		s.Values[Resource] = s.Environment.ResourceManagerEndpoint
	}
	return
}

// GetSubscriptionID returns the available subscription ID or an empty string.
func (settings EnvironmentSettings) GetSubscriptionID() string {
	return settings.Values[SubscriptionID]
}

// adds the specified environment variable value to the Values map if it exists
func (settings EnvironmentSettings) setValue(key string) {
	if v := os.Getenv(key); v != "" {
		settings.Values[key] = v
	}
}

// helper to return client and tenant IDs
func (settings EnvironmentSettings) getClientAndTenant() (string, string) {
	clientID := settings.Values[ClientID]
	tenantID := settings.Values[TenantID]
	return clientID, tenantID
}

// GetClientCredentials creates a config object from the available client credentials.
// An error is returned if no client credentials are available.
func (settings EnvironmentSettings) GetClientCredentials() (ClientCredentialsConfig, error) {
	secret := settings.Values[ClientSecret]
	if secret == "" {
		return ClientCredentialsConfig{}, errors.New("missing client secret")
	}
	clientID, tenantID := settings.getClientAndTenant()
	config := NewClientCredentialsConfig(clientID, secret, tenantID)
	config.AADEndpoint = settings.Environment.ActiveDirectoryEndpoint
	config.Resource = settings.Values[Resource]
	if auxTenants, ok := settings.Values[AuxiliaryTenantIDs]; ok {
		config.AuxTenants = strings.Split(auxTenants, ";")
		for i := range config.AuxTenants {
			config.AuxTenants[i] = strings.TrimSpace(config.AuxTenants[i])
		}
	}
	return config, nil
}

// GetClientCertificate creates a config object from the available certificate credentials.
// An error is returned if no certificate credentials are available.
func (settings EnvironmentSettings) GetClientCertificate() (ClientCertificateConfig, error) {
	certPath := settings.Values[CertificatePath]
	if certPath == "" {
		return ClientCertificateConfig{}, errors.New("missing certificate path")
	}
	certPwd := settings.Values[CertificatePassword]
	clientID, tenantID := settings.getClientAndTenant()
	config := NewClientCertificateConfig(certPath, certPwd, clientID, tenantID)
	config.AADEndpoint = settings.Environment.ActiveDirectoryEndpoint
	config.Resource = settings.Values[Resource]
	return config, nil
}

// GetUsernamePassword creates a config object from the available username/password credentials.
// An error is returned if no username/password credentials are available.
func (settings EnvironmentSettings) GetUsernamePassword() (UsernamePasswordConfig, error) {
	username := settings.Values[Username]
	password := settings.Values[Password]
	if username == "" || password == "" {
		return UsernamePasswordConfig{}, errors.New("missing username/password")
	}
	clientID, tenantID := settings.getClientAndTenant()
	config := NewUsernamePasswordConfig(username, password, clientID, tenantID)
	config.AADEndpoint = settings.Environment.ActiveDirectoryEndpoint
	config.Resource = settings.Values[Resource]
	return config, nil
}

// GetMSI creates a MSI config object from the available client ID.
func (settings EnvironmentSettings) GetMSI() MSIConfig {
	config := NewMSIConfig()
	config.Resource = settings.Values[Resource]
	config.ClientID = settings.Values[ClientID]
	return config
}

// GetDeviceFlow creates a device-flow config object from the available client and tenant IDs.
func (settings EnvironmentSettings) GetDeviceFlow() DeviceFlowConfig {
	clientID, tenantID := settings.getClientAndTenant()
	config := NewDeviceFlowConfig(clientID, tenantID)
	config.AADEndpoint = settings.Environment.ActiveDirectoryEndpoint
	config.Resource = settings.Values[Resource]
	return config
}

// GetAuthorizer creates an Authorizer configured from environment variables in the order:
// 1. Client credentials
// 2. Client certificate
// 3. Username password
// 4. MSI
func (settings EnvironmentSettings) GetAuthorizer() (autorest.Authorizer, error) {
	//1.Client Credentials
	if c, e := settings.GetClientCredentials(); e == nil {
		return c.Authorizer()
	}

	//2. Client Certificate
	if c, e := settings.GetClientCertificate(); e == nil {
		return c.Authorizer()
	}

	//3. Username Password
	if c, e := settings.GetUsernamePassword(); e == nil {
		return c.Authorizer()
	}

	// 4. MSI
	return settings.GetMSI().Authorizer()
}

// NewAuthorizerFromFile creates an Authorizer configured from a configuration file in the following order.
// 1. Client credentials
// 2. Client certificate
func NewAuthorizerFromFile(baseURI string) (autorest.Authorizer, error) {
	settings, err := GetSettingsFromFile()
	if err != nil {
		return nil, err
	}
	if a, err := settings.ClientCredentialsAuthorizer(baseURI); err == nil {
		return a, err
	}
	if a, err := settings.ClientCertificateAuthorizer(baseURI); err == nil {
		return a, err
	}
	return nil, errors.New("auth file missing client and certificate credentials")
}

// NewAuthorizerFromFileWithResource creates an Authorizer configured from a configuration file in the following order.
// 1. Client credentials
// 2. Client certificate
func NewAuthorizerFromFileWithResource(resource string) (autorest.Authorizer, error) {
	s, err := GetSettingsFromFile()
	if err != nil {
		return nil, err
	}
	if a, err := s.ClientCredentialsAuthorizerWithResource(resource); err == nil {
		return a, err
	}
	if a, err := s.ClientCertificateAuthorizerWithResource(resource); err == nil {
		return a, err
	}
	return nil, errors.New("auth file missing client and certificate credentials")
}

// NewAuthorizerFromCLI creates an Authorizer configured from Azure CLI 2.0 for local development scenarios.
func NewAuthorizerFromCLI() (autorest.Authorizer, error) {
	settings, err := GetSettingsFromEnvironment()
	if err != nil {
		return nil, err
	}

	if settings.Values[Resource] == "" {
		settings.Values[Resource] = settings.Environment.ResourceManagerEndpoint
	}

	return NewAuthorizerFromCLIWithResource(settings.Values[Resource])
}

// NewAuthorizerFromCLIWithResource creates an Authorizer configured from Azure CLI 2.0 for local development scenarios.
func NewAuthorizerFromCLIWithResource(resource string) (autorest.Authorizer, error) {
	token, err := cli.GetTokenFromCLI(resource)
	if err != nil {
		return nil, err
	}

	adalToken, err := token.ToADALToken()
	if err != nil {
		return nil, err
	}

	return autorest.NewBearerAuthorizer(&adalToken), nil
}

// GetSettingsFromFile returns the available authentication settings from an Azure CLI authentication file.
func GetSettingsFromFile() (FileSettings, error) {
	s := FileSettings{}
	fileLocation := os.Getenv("AZURE_AUTH_LOCATION")
	if fileLocation == "" {
		return s, errors.New("environment variable AZURE_AUTH_LOCATION is not set")
	}

	contents, err := ioutil.ReadFile(fileLocation)
	if err != nil {
		return s, err
	}

	// Auth file might be encoded
	decoded, err := decode(contents)
	if err != nil {
		return s, err
	}

	authFile := map[string]interface{}{}
	err = json.Unmarshal(decoded, &authFile)
	if err != nil {
		return s, err
	}

	s.Values = map[string]string{}
	s.setKeyValue(ClientID, authFile["clientId"])
	s.setKeyValue(ClientSecret, authFile["clientSecret"])
	s.setKeyValue(CertificatePath, authFile["clientCertificate"])
	s.setKeyValue(CertificatePassword, authFile["clientCertificatePassword"])
	s.setKeyValue(SubscriptionID, authFile["subscriptionId"])
	s.setKeyValue(TenantID, authFile["tenantId"])
	s.setKeyValue(ActiveDirectoryEndpoint, authFile["activeDirectoryEndpointUrl"])
	s.setKeyValue(ResourceManagerEndpoint, authFile["resourceManagerEndpointUrl"])
	s.setKeyValue(GraphResourceID, authFile["activeDirectoryGraphResourceId"])
	s.setKeyValue(SQLManagementEndpoint, authFile["sqlManagementEndpointUrl"])
	s.setKeyValue(GalleryEndpoint, authFile["galleryEndpointUrl"])
	s.setKeyValue(ManagementEndpoint, authFile["managementEndpointUrl"])
	return s, nil
}

// FileSettings contains the available authentication settings.
type FileSettings struct {
	Values map[string]string
}

// GetSubscriptionID returns the available subscription ID or an empty string.
func (settings FileSettings) GetSubscriptionID() string {
	return settings.Values[SubscriptionID]
}

// adds the specified value to the Values map if it isn't nil
func (settings FileSettings) setKeyValue(key string, val interface{}) {
	if val != nil {
		settings.Values[key] = val.(string)
	}
}

// returns the specified AAD endpoint or the public cloud endpoint if unspecified
func (settings FileSettings) getAADEndpoint() string {
	if v, ok := settings.Values[ActiveDirectoryEndpoint]; ok {
		return v
	}
	return azure.PublicCloud.ActiveDirectoryEndpoint
}

// ServicePrincipalTokenFromClientCredentials creates a ServicePrincipalToken from the available client credentials.
func (settings FileSettings) ServicePrincipalTokenFromClientCredentials(baseURI string) (*adal.ServicePrincipalToken, error) {
	resource, err := settings.getResourceForToken(baseURI)
	if err != nil {
		return nil, err
	}
	return settings.ServicePrincipalTokenFromClientCredentialsWithResource(resource)
}

// ClientCredentialsAuthorizer creates an authorizer from the available client credentials.
func (settings FileSettings) ClientCredentialsAuthorizer(baseURI string) (autorest.Authorizer, error) {
	resource, err := settings.getResourceForToken(baseURI)
	if err != nil {
		return nil, err
	}
	return settings.ClientCredentialsAuthorizerWithResource(resource)
}

// ServicePrincipalTokenFromClientCredentialsWithResource creates a ServicePrincipalToken
// from the available client credentials and the specified resource.
func (settings FileSettings) ServicePrincipalTokenFromClientCredentialsWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	if _, ok := settings.Values[ClientSecret]; !ok {
		return nil, errors.New("missing client secret")
	}
	config, err := adal.NewOAuthConfig(settings.getAADEndpoint(), settings.Values[TenantID])
	if err != nil {
		return nil, err
	}
	return adal.NewServicePrincipalToken(*config, settings.Values[ClientID], settings.Values[ClientSecret], resource)
}

func (settings FileSettings) clientCertificateConfigWithResource(resource string) (ClientCertificateConfig, error) {
	if _, ok := settings.Values[CertificatePath]; !ok {
		return ClientCertificateConfig{}, errors.New("missing certificate path")
	}
	cfg := NewClientCertificateConfig(settings.Values[CertificatePath], settings.Values[CertificatePassword], settings.Values[ClientID], settings.Values[TenantID])
	cfg.AADEndpoint = settings.getAADEndpoint()
	cfg.Resource = resource
	return cfg, nil
}

// ClientCredentialsAuthorizerWithResource creates an authorizer from the available client credentials and the specified resource.
func (settings FileSettings) ClientCredentialsAuthorizerWithResource(resource string) (autorest.Authorizer, error) {
	spToken, err := settings.ServicePrincipalTokenFromClientCredentialsWithResource(resource)
	if err != nil {
		return nil, err
	}
	return autorest.NewBearerAuthorizer(spToken), nil
}

// ServicePrincipalTokenFromClientCertificate creates a ServicePrincipalToken from the available certificate credentials.
func (settings FileSettings) ServicePrincipalTokenFromClientCertificate(baseURI string) (*adal.ServicePrincipalToken, error) {
	resource, err := settings.getResourceForToken(baseURI)
	if err != nil {
		return nil, err
	}
	return settings.ServicePrincipalTokenFromClientCertificateWithResource(resource)
}

// ClientCertificateAuthorizer creates an authorizer from the available certificate credentials.
func (settings FileSettings) ClientCertificateAuthorizer(baseURI string) (autorest.Authorizer, error) {
	resource, err := settings.getResourceForToken(baseURI)
	if err != nil {
		return nil, err
	}
	return settings.ClientCertificateAuthorizerWithResource(resource)
}

// ServicePrincipalTokenFromClientCertificateWithResource creates a ServicePrincipalToken from the available certificate credentials.
func (settings FileSettings) ServicePrincipalTokenFromClientCertificateWithResource(resource string) (*adal.ServicePrincipalToken, error) {
	cfg, err := settings.clientCertificateConfigWithResource(resource)
	if err != nil {
		return nil, err
	}
	return cfg.ServicePrincipalToken()
}

// ClientCertificateAuthorizerWithResource creates an authorizer from the available certificate credentials and the specified resource.
func (settings FileSettings) ClientCertificateAuthorizerWithResource(resource string) (autorest.Authorizer, error) {
	cfg, err := settings.clientCertificateConfigWithResource(resource)
	if err != nil {
		return nil, err
	}
	return cfg.Authorizer()
}

func decode(b []byte) ([]byte, error) {
	reader, enc := utfbom.Skip(bytes.NewReader(b))

	switch enc {
	case utfbom.UTF16LittleEndian:
		u16 := make([]uint16, (len(b)/2)-1)
		err := binary.Read(reader, binary.LittleEndian, &u16)
		if err != nil {
			return nil, err
		}
		return []byte(string(utf16.Decode(u16))), nil
	case utfbom.UTF16BigEndian:
		u16 := make([]uint16, (len(b)/2)-1)
		err := binary.Read(reader, binary.BigEndian, &u16)
		if err != nil {
			return nil, err
		}
		return []byte(string(utf16.Decode(u16))), nil
	}
	return ioutil.ReadAll(reader)
}

func (settings FileSettings) getResourceForToken(baseURI string) (string, error) {
	// Compare dafault base URI from the SDK to the endpoints from the public cloud
	// Base URI and token resource are the same string. This func finds the authentication
	// file field that matches the SDK base URI. The SDK defines the public cloud
	// endpoint as its default base URI
	if !strings.HasSuffix(baseURI, "/") {
		baseURI += "/"
	}
	switch baseURI {
	case azure.PublicCloud.ServiceManagementEndpoint:
		return settings.Values[ManagementEndpoint], nil
	case azure.PublicCloud.ResourceManagerEndpoint:
		return settings.Values[ResourceManagerEndpoint], nil
	case azure.PublicCloud.ActiveDirectoryEndpoint:
		return settings.Values[ActiveDirectoryEndpoint], nil
	case azure.PublicCloud.GalleryEndpoint:
		return settings.Values[GalleryEndpoint], nil
	case azure.PublicCloud.GraphEndpoint:
		return settings.Values[GraphResourceID], nil
	}
	return "", fmt.Errorf("auth: base URI not found in endpoints")
}

// NewClientCredentialsConfig creates an AuthorizerConfig object configured to obtain an Authorizer through Client Credentials.
// Defaults to Public Cloud and Resource Manager Endpoint.
func NewClientCredentialsConfig(clientID string, clientSecret string, tenantID string) ClientCredentialsConfig {
	return ClientCredentialsConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TenantID:     tenantID,
		Resource:     azure.PublicCloud.ResourceManagerEndpoint,
		AADEndpoint:  azure.PublicCloud.ActiveDirectoryEndpoint,
	}
}

// NewClientCertificateConfig creates a ClientCertificateConfig object configured to obtain an Authorizer through client certificate.
// Defaults to Public Cloud and Resource Manager Endpoint.
func NewClientCertificateConfig(certificatePath string, certificatePassword string, clientID string, tenantID string) ClientCertificateConfig {
	return ClientCertificateConfig{
		CertificatePath:     certificatePath,
		CertificatePassword: certificatePassword,
		ClientID:            clientID,
		TenantID:            tenantID,
		Resource:            azure.PublicCloud.ResourceManagerEndpoint,
		AADEndpoint:         azure.PublicCloud.ActiveDirectoryEndpoint,
	}
}

// NewUsernamePasswordConfig creates an UsernamePasswordConfig object configured to obtain an Authorizer through username and password.
// Defaults to Public Cloud and Resource Manager Endpoint.
func NewUsernamePasswordConfig(username string, password string, clientID string, tenantID string) UsernamePasswordConfig {
	return UsernamePasswordConfig{
		Username:    username,
		Password:    password,
		ClientID:    clientID,
		TenantID:    tenantID,
		Resource:    azure.PublicCloud.ResourceManagerEndpoint,
		AADEndpoint: azure.PublicCloud.ActiveDirectoryEndpoint,
	}
}

// NewMSIConfig creates an MSIConfig object configured to obtain an Authorizer through MSI.
func NewMSIConfig() MSIConfig {
	return MSIConfig{
		Resource: azure.PublicCloud.ResourceManagerEndpoint,
	}
}

// NewDeviceFlowConfig creates a DeviceFlowConfig object configured to obtain an Authorizer through device flow.
// Defaults to Public Cloud and Resource Manager Endpoint.
func NewDeviceFlowConfig(clientID string, tenantID string) DeviceFlowConfig {
	return DeviceFlowConfig{
		ClientID:    clientID,
		TenantID:    tenantID,
		Resource:    azure.PublicCloud.ResourceManagerEndpoint,
		AADEndpoint: azure.PublicCloud.ActiveDirectoryEndpoint,
	}
}

//AuthorizerConfig provides an authorizer from the configuration provided.
type AuthorizerConfig interface {
	Authorizer() (autorest.Authorizer, error)
}

// ClientCredentialsConfig provides the options to get a bearer authorizer from client credentials.
type ClientCredentialsConfig struct {
	ClientID     string
	ClientSecret string
	TenantID     string
	AuxTenants   []string
	AADEndpoint  string
	Resource     string
}

// ServicePrincipalToken creates a ServicePrincipalToken from client credentials.
func (ccc ClientCredentialsConfig) ServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(ccc.AADEndpoint, ccc.TenantID)
	if err != nil {
		return nil, err
	}
	return adal.NewServicePrincipalToken(*oauthConfig, ccc.ClientID, ccc.ClientSecret, ccc.Resource)
}

// MultiTenantServicePrincipalToken creates a MultiTenantServicePrincipalToken from client credentials.
func (ccc ClientCredentialsConfig) MultiTenantServicePrincipalToken() (*adal.MultiTenantServicePrincipalToken, error) {
	oauthConfig, err := adal.NewMultiTenantOAuthConfig(ccc.AADEndpoint, ccc.TenantID, ccc.AuxTenants, adal.OAuthOptions{})
	if err != nil {
		return nil, err
	}
	return adal.NewMultiTenantServicePrincipalToken(oauthConfig, ccc.ClientID, ccc.ClientSecret, ccc.Resource)
}

// Authorizer gets the authorizer from client credentials.
func (ccc ClientCredentialsConfig) Authorizer() (autorest.Authorizer, error) {
	if len(ccc.AuxTenants) == 0 {
		spToken, err := ccc.ServicePrincipalToken()
		if err != nil {
			return nil, fmt.Errorf("failed to get SPT from client credentials: %v", err)
		}
		return autorest.NewBearerAuthorizer(spToken), nil
	}
	mtSPT, err := ccc.MultiTenantServicePrincipalToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get multitenant SPT from client credentials: %v", err)
	}
	return autorest.NewMultiTenantServicePrincipalTokenAuthorizer(mtSPT), nil
}

// ClientCertificateConfig provides the options to get a bearer authorizer from a client certificate.
type ClientCertificateConfig struct {
	ClientID            string
	CertificatePath     string
	CertificatePassword string
	TenantID            string
	AADEndpoint         string
	Resource            string
}

// ServicePrincipalToken creates a ServicePrincipalToken from client certificate.
func (ccc ClientCertificateConfig) ServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(ccc.AADEndpoint, ccc.TenantID)
	if err != nil {
		return nil, err
	}
	certData, err := ioutil.ReadFile(ccc.CertificatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read the certificate file (%s): %v", ccc.CertificatePath, err)
	}
	certificate, rsaPrivateKey, err := decodePkcs12(certData, ccc.CertificatePassword)
	if err != nil {
		return nil, fmt.Errorf("failed to decode pkcs12 certificate while creating spt: %v", err)
	}
	return adal.NewServicePrincipalTokenFromCertificate(*oauthConfig, ccc.ClientID, certificate, rsaPrivateKey, ccc.Resource)
}

// Authorizer gets an authorizer object from client certificate.
func (ccc ClientCertificateConfig) Authorizer() (autorest.Authorizer, error) {
	spToken, err := ccc.ServicePrincipalToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth token from certificate auth: %v", err)
	}
	return autorest.NewBearerAuthorizer(spToken), nil
}

// DeviceFlowConfig provides the options to get a bearer authorizer using device flow authentication.
type DeviceFlowConfig struct {
	ClientID    string
	TenantID    string
	AADEndpoint string
	Resource    string
}

// Authorizer gets the authorizer from device flow.
func (dfc DeviceFlowConfig) Authorizer() (autorest.Authorizer, error) {
	spToken, err := dfc.ServicePrincipalToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth token from device flow: %v", err)
	}
	return autorest.NewBearerAuthorizer(spToken), nil
}

// ServicePrincipalToken gets the service principal token from device flow.
func (dfc DeviceFlowConfig) ServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(dfc.AADEndpoint, dfc.TenantID)
	if err != nil {
		return nil, err
	}
	oauthClient := &autorest.Client{}
	deviceCode, err := adal.InitiateDeviceAuth(oauthClient, *oauthConfig, dfc.ClientID, dfc.Resource)
	if err != nil {
		return nil, fmt.Errorf("failed to start device auth flow: %s", err)
	}
	log.Println(*deviceCode.Message)
	token, err := adal.WaitForUserCompletion(oauthClient, deviceCode)
	if err != nil {
		return nil, fmt.Errorf("failed to finish device auth flow: %s", err)
	}
	return adal.NewServicePrincipalTokenFromManualToken(*oauthConfig, dfc.ClientID, dfc.Resource, *token)
}

func decodePkcs12(pkcs []byte, password string) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, certificate, err := pkcs12.Decode(pkcs, password)
	if err != nil {
		return nil, nil, err
	}

	rsaPrivateKey, isRsaKey := privateKey.(*rsa.PrivateKey)
	if !isRsaKey {
		return nil, nil, fmt.Errorf("PKCS#12 certificate must contain an RSA private key")
	}

	return certificate, rsaPrivateKey, nil
}

// UsernamePasswordConfig provides the options to get a bearer authorizer from a username and a password.
type UsernamePasswordConfig struct {
	ClientID    string
	Username    string
	Password    string
	TenantID    string
	AADEndpoint string
	Resource    string
}

// ServicePrincipalToken creates a ServicePrincipalToken from username and password.
func (ups UsernamePasswordConfig) ServicePrincipalToken() (*adal.ServicePrincipalToken, error) {
	oauthConfig, err := adal.NewOAuthConfig(ups.AADEndpoint, ups.TenantID)
	if err != nil {
		return nil, err
	}
	return adal.NewServicePrincipalTokenFromUsernamePassword(*oauthConfig, ups.ClientID, ups.Username, ups.Password, ups.Resource)
}

// Authorizer gets the authorizer from a username and a password.
func (ups UsernamePasswordConfig) Authorizer() (autorest.Authorizer, error) {
	spToken, err := ups.ServicePrincipalToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth token from username and password auth: %v", err)
	}
	return autorest.NewBearerAuthorizer(spToken), nil
}

// MSIConfig provides the options to get a bearer authorizer through MSI.
type MSIConfig struct {
	Resource string
	ClientID string
}

// Authorizer gets the authorizer from MSI.
func (mc MSIConfig) Authorizer() (autorest.Authorizer, error) {
	msiEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		return nil, err
	}

	var spToken *adal.ServicePrincipalToken
	if mc.ClientID == "" {
		spToken, err = adal.NewServicePrincipalTokenFromMSI(msiEndpoint, mc.Resource)
		if err != nil {
			return nil, fmt.Errorf("failed to get oauth token from MSI: %v", err)
		}
	} else {
		spToken, err = adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint, mc.Resource, mc.ClientID)
		if err != nil {
			return nil, fmt.Errorf("failed to get oauth token from MSI for user assigned identity: %v", err)
		}
	}

	return autorest.NewBearerAuthorizer(spToken), nil
}
