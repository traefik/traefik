package storage

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	chk "gopkg.in/check.v1"
)

// Hook up gocheck to testing
func Test(t *testing.T) { chk.TestingT(t) }

type StorageClientSuite struct{}

var _ = chk.Suite(&StorageClientSuite{})

var now = time.Now()

// getBasicClient returns a test client from storage credentials in the env
func getBasicClient(c *chk.C) Client {
	name := os.Getenv("ACCOUNT_NAME")
	if name == "" {
		c.Fatal("ACCOUNT_NAME not set, need an empty storage account to test")
	}
	key := os.Getenv("ACCOUNT_KEY")
	if key == "" {
		c.Fatal("ACCOUNT_KEY not set")
	}
	cli, err := NewBasicClient(name, key)
	c.Assert(err, chk.IsNil)
	return cli
}

//getEmulatorClient returns a test client for Azure Storeage Emulator
func getEmulatorClient(c *chk.C) Client {
	cli, err := NewBasicClient(StorageEmulatorAccountName, "")
	c.Assert(err, chk.IsNil)
	return cli
}

func (s *StorageClientSuite) TestNewEmulatorClient(c *chk.C) {
	cli, err := NewBasicClient(StorageEmulatorAccountName, "")
	c.Assert(err, chk.IsNil)
	c.Assert(cli.accountName, chk.Equals, StorageEmulatorAccountName)
	expectedKey, err := base64.StdEncoding.DecodeString(StorageEmulatorAccountKey)
	c.Assert(err, chk.IsNil)
	c.Assert(cli.accountKey, chk.DeepEquals, expectedKey)
}

func (s *StorageClientSuite) TestMalformedKeyError(c *chk.C) {
	_, err := NewBasicClient("foo", "malformed")
	c.Assert(err, chk.ErrorMatches, "azure: malformed storage account key: .*")
}

func (s *StorageClientSuite) TestGetBaseURL_Basic_Https(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	c.Assert(cli.apiVersion, chk.Equals, DefaultAPIVersion)
	c.Assert(err, chk.IsNil)
	c.Assert(cli.getBaseURL("table"), chk.Equals, "https://foo.table.core.windows.net")
}

func (s *StorageClientSuite) TestGetBaseURL_Custom_NoHttps(c *chk.C) {
	apiVersion := "2015-01-01" // a non existing one
	cli, err := NewClient("foo", "YmFy", "core.chinacloudapi.cn", apiVersion, false)
	c.Assert(err, chk.IsNil)
	c.Assert(cli.apiVersion, chk.Equals, apiVersion)
	c.Assert(cli.getBaseURL("table"), chk.Equals, "http://foo.table.core.chinacloudapi.cn")
}

func (s *StorageClientSuite) TestGetBaseURL_StorageEmulator(c *chk.C) {
	cli, err := NewBasicClient(StorageEmulatorAccountName, StorageEmulatorAccountKey)
	c.Assert(err, chk.IsNil)

	type test struct{ service, expected string }
	tests := []test{
		{blobServiceName, "http://127.0.0.1:10000"},
		{tableServiceName, "http://127.0.0.1:10002"},
		{queueServiceName, "http://127.0.0.1:10001"},
	}
	for _, i := range tests {
		baseURL := cli.getBaseURL(i.service)
		c.Assert(baseURL, chk.Equals, i.expected)
	}
}

func (s *StorageClientSuite) TestGetEndpoint_None(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	output := cli.getEndpoint(blobServiceName, "", url.Values{})
	c.Assert(output, chk.Equals, "https://foo.blob.core.windows.net/")
}

func (s *StorageClientSuite) TestGetEndpoint_PathOnly(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	output := cli.getEndpoint(blobServiceName, "path", url.Values{})
	c.Assert(output, chk.Equals, "https://foo.blob.core.windows.net/path")
}

func (s *StorageClientSuite) TestGetEndpoint_ParamsOnly(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	params := url.Values{}
	params.Set("a", "b")
	params.Set("c", "d")
	output := cli.getEndpoint(blobServiceName, "", params)
	c.Assert(output, chk.Equals, "https://foo.blob.core.windows.net/?a=b&c=d")
}

func (s *StorageClientSuite) TestGetEndpoint_Mixed(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	params := url.Values{}
	params.Set("a", "b")
	params.Set("c", "d")
	output := cli.getEndpoint(blobServiceName, "path", params)
	c.Assert(output, chk.Equals, "https://foo.blob.core.windows.net/path?a=b&c=d")
}

func (s *StorageClientSuite) TestGetEndpoint_StorageEmulator(c *chk.C) {
	cli, err := NewBasicClient(StorageEmulatorAccountName, StorageEmulatorAccountKey)
	c.Assert(err, chk.IsNil)

	type test struct{ service, expected string }
	tests := []test{
		{blobServiceName, "http://127.0.0.1:10000/devstoreaccount1/"},
		{tableServiceName, "http://127.0.0.1:10002/devstoreaccount1/"},
		{queueServiceName, "http://127.0.0.1:10001/devstoreaccount1/"},
	}
	for _, i := range tests {
		endpoint := cli.getEndpoint(i.service, "", url.Values{})
		c.Assert(endpoint, chk.Equals, i.expected)
	}
}

func (s *StorageClientSuite) Test_getStandardHeaders(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)

	headers := cli.getStandardHeaders()
	c.Assert(len(headers), chk.Equals, 3)
	c.Assert(headers["x-ms-version"], chk.Equals, cli.apiVersion)
	if _, ok := headers["x-ms-date"]; !ok {
		c.Fatal("Missing date header")
	}
	c.Assert(headers[userAgentHeader], chk.Equals, cli.getDefaultUserAgent())
}

func (s *StorageClientSuite) TestReturnsStorageServiceError(c *chk.C) {
	// attempt to delete a nonexisting container
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	_, err := cnt.delete()
	c.Assert(err, chk.NotNil)

	v, ok := err.(AzureStorageServiceError)
	c.Check(ok, chk.Equals, true)
	c.Assert(v.StatusCode, chk.Equals, 404)
	c.Assert(v.Code, chk.Equals, "ContainerNotFound")
	c.Assert(v.Code, chk.Not(chk.Equals), "")
	c.Assert(v.RequestID, chk.Not(chk.Equals), "")
}

func (s *StorageClientSuite) TestReturnsStorageServiceError_withoutResponseBody(c *chk.C) {
	// HEAD on non-existing blob
	_, err := getBlobClient(c).GetBlobProperties("non-existing-blob", "non-existing-container")

	c.Assert(err, chk.NotNil)
	c.Assert(err, chk.FitsTypeOf, AzureStorageServiceError{})

	v, ok := err.(AzureStorageServiceError)
	c.Check(ok, chk.Equals, true)
	c.Assert(v.StatusCode, chk.Equals, http.StatusNotFound)
	c.Assert(v.Code, chk.Equals, "404 The specified container does not exist.")
	c.Assert(v.RequestID, chk.Not(chk.Equals), "")
	c.Assert(v.Message, chk.Equals, "no response body was available for error status code")
}

func (s *StorageClientSuite) Test_createServiceClients(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)

	ua := cli.getDefaultUserAgent()

	headers := cli.getStandardHeaders()
	c.Assert(headers[userAgentHeader], chk.Equals, ua)
	c.Assert(cli.userAgent, chk.Equals, ua)

	b := cli.GetBlobService()
	c.Assert(b.client.userAgent, chk.Equals, ua+" "+blobServiceName)
	c.Assert(cli.userAgent, chk.Equals, ua)

	t := cli.GetTableService()
	c.Assert(t.client.userAgent, chk.Equals, ua+" "+tableServiceName)
	c.Assert(cli.userAgent, chk.Equals, ua)

	q := cli.GetQueueService()
	c.Assert(q.client.userAgent, chk.Equals, ua+" "+queueServiceName)
	c.Assert(cli.userAgent, chk.Equals, ua)

	f := cli.GetFileService()
	c.Assert(f.client.userAgent, chk.Equals, ua+" "+fileServiceName)
	c.Assert(cli.userAgent, chk.Equals, ua)
}

func (s *StorageClientSuite) TestAddToUserAgent(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)

	ua := cli.getDefaultUserAgent()

	err = cli.AddToUserAgent("bar")
	c.Assert(err, chk.IsNil)
	c.Assert(cli.userAgent, chk.Equals, ua+" bar")

	err = cli.AddToUserAgent("")
	c.Assert(err, chk.NotNil)
}

func (s *StorageClientSuite) Test_protectUserAgent(c *chk.C) {
	extraheaders := map[string]string{
		"1":             "one",
		"2":             "two",
		"3":             "three",
		userAgentHeader: "four",
	}

	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)

	ua := cli.getDefaultUserAgent()

	got := cli.protectUserAgent(extraheaders)
	c.Assert(cli.userAgent, chk.Equals, ua+" four")
	c.Assert(got, chk.HasLen, 3)
	c.Assert(got, chk.DeepEquals, map[string]string{
		"1": "one",
		"2": "two",
		"3": "three",
	})
}
