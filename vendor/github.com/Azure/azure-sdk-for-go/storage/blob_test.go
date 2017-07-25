package storage

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"

	chk "gopkg.in/check.v1"
)

type StorageBlobSuite struct{}

var _ = chk.Suite(&StorageBlobSuite{})

const testContainerPrefix = "zzzztest-"

func getBlobClient(c *chk.C) BlobStorageClient {
	return getBasicClient(c).GetBlobService()
}

func (s *StorageBlobSuite) Test_pathForBlob(c *chk.C) {
	c.Assert(pathForBlob("foo", "blob"), chk.Equals, "/foo/blob")
}

func (s *StorageBlobSuite) Test_pathForResource(c *chk.C) {
	c.Assert(pathForResource("foo", ""), chk.Equals, "/foo")
	c.Assert(pathForResource("foo", "blob"), chk.Equals, "/foo/blob")
}

func (s *StorageBlobSuite) Test_blobSASStringToSign(c *chk.C) {
	_, err := blobSASStringToSign("2012-02-12", "CS", "SE", "SP", "", "")
	c.Assert(err, chk.NotNil) // not implemented SAS for versions earlier than 2013-08-15

	out, err := blobSASStringToSign("2013-08-15", "CS", "SE", "SP", "", "")
	c.Assert(err, chk.IsNil)
	c.Assert(out, chk.Equals, "SP\n\nSE\nCS\n\n2013-08-15\n\n\n\n\n")

	// check format for 2015-04-05 version
	out, err = blobSASStringToSign("2015-04-05", "CS", "SE", "SP", "127.0.0.1", "https,http")
	c.Assert(err, chk.IsNil)
	c.Assert(out, chk.Equals, "SP\n\nSE\n/blobCS\n\n127.0.0.1\nhttps,http\n2015-04-05\n\n\n\n\n")
}

func (s *StorageBlobSuite) TestGetBlobSASURI(c *chk.C) {
	api, err := NewClient("foo", "YmFy", DefaultBaseURL, "2013-08-15", true)
	c.Assert(err, chk.IsNil)
	cli := api.GetBlobService()
	expiry := time.Time{}

	expectedParts := url.URL{
		Scheme: "https",
		Host:   "foo.blob.core.windows.net",
		Path:   "container/name",
		RawQuery: url.Values{
			"sv":  {"2013-08-15"},
			"sig": {"/OXG7rWh08jYwtU03GzJM0DHZtidRGpC6g69rSGm3I0="},
			"sr":  {"b"},
			"sp":  {"r"},
			"se":  {"0001-01-01T00:00:00Z"},
		}.Encode()}

	u, err := cli.GetBlobSASURI("container", "name", expiry, "r")
	c.Assert(err, chk.IsNil)
	sasParts, err := url.Parse(u)
	c.Assert(err, chk.IsNil)
	c.Assert(expectedParts.String(), chk.Equals, sasParts.String())
	c.Assert(expectedParts.Query(), chk.DeepEquals, sasParts.Query())
}

//Gets a SASURI for the entire container
func (s *StorageBlobSuite) TestGetBlobSASURIContainer(c *chk.C) {
	api, err := NewClient("foo", "YmFy", DefaultBaseURL, "2013-08-15", true)
	c.Assert(err, chk.IsNil)
	cli := api.GetBlobService()
	expiry := time.Time{}

	expectedParts := url.URL{
		Scheme: "https",
		Host:   "foo.blob.core.windows.net",
		Path:   "container",
		RawQuery: url.Values{
			"sv":  {"2013-08-15"},
			"sig": {"KMjYyQODKp6uK9EKR3yGhO2M84e1LfoztypU32kHj4s="},
			"sr":  {"c"},
			"sp":  {"r"},
			"se":  {"0001-01-01T00:00:00Z"},
		}.Encode()}

	u, err := cli.GetBlobSASURI("container", "", expiry, "r")
	c.Assert(err, chk.IsNil)
	sasParts, err := url.Parse(u)
	c.Assert(err, chk.IsNil)
	c.Assert(expectedParts.String(), chk.Equals, sasParts.String())
	c.Assert(expectedParts.Query(), chk.DeepEquals, sasParts.Query())
}

func (s *StorageBlobSuite) TestGetBlobSASURIWithSignedIPAndProtocolValidAPIVersionPassed(c *chk.C) {
	api, err := NewClient("foo", "YmFy", DefaultBaseURL, "2015-04-05", true)
	c.Assert(err, chk.IsNil)
	cli := api.GetBlobService()
	expiry := time.Time{}

	expectedParts := url.URL{
		Scheme: "https",
		Host:   "foo.blob.core.windows.net",
		Path:   "/container/name",
		RawQuery: url.Values{
			"sv":  {"2015-04-05"},
			"sig": {"VBOYJmt89UuBRXrxNzmsCMoC+8PXX2yklV71QcL1BfM="},
			"sr":  {"b"},
			"sip": {"127.0.0.1"},
			"sp":  {"r"},
			"se":  {"0001-01-01T00:00:00Z"},
			"spr": {"https"},
		}.Encode()}

	u, err := cli.GetBlobSASURIWithSignedIPAndProtocol("container", "name", expiry, "r", "127.0.0.1", true)
	c.Assert(err, chk.IsNil)
	sasParts, err := url.Parse(u)
	c.Assert(err, chk.IsNil)
	c.Assert(sasParts.Query(), chk.DeepEquals, expectedParts.Query())
}

// Trying to use SignedIP and Protocol but using an older version of the API.
// Should ignore the signedIP/protocol and just use what the older version requires.
func (s *StorageBlobSuite) TestGetBlobSASURIWithSignedIPAndProtocolUsingOldAPIVersion(c *chk.C) {
	api, err := NewClient("foo", "YmFy", DefaultBaseURL, "2013-08-15", true)
	c.Assert(err, chk.IsNil)
	cli := api.GetBlobService()
	expiry := time.Time{}

	expectedParts := url.URL{
		Scheme: "https",
		Host:   "foo.blob.core.windows.net",
		Path:   "/container/name",
		RawQuery: url.Values{
			"sv":  {"2013-08-15"},
			"sig": {"/OXG7rWh08jYwtU03GzJM0DHZtidRGpC6g69rSGm3I0="},
			"sr":  {"b"},
			"sp":  {"r"},
			"se":  {"0001-01-01T00:00:00Z"},
		}.Encode()}

	u, err := cli.GetBlobSASURIWithSignedIPAndProtocol("container", "name", expiry, "r", "", true)
	c.Assert(err, chk.IsNil)
	sasParts, err := url.Parse(u)
	c.Assert(err, chk.IsNil)
	c.Assert(expectedParts.String(), chk.Equals, sasParts.String())
	c.Assert(expectedParts.Query(), chk.DeepEquals, sasParts.Query())
}

func (s *StorageBlobSuite) TestBlobSASURICorrectness(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randNameWithSpecialChars(5)
	body := []byte(randString(100))
	expiry := now.UTC().Add(time.Hour)
	permissions := "r"

	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, body), chk.IsNil)

	sasURI, err := cli.GetBlobSASURI(cnt.Name, blob, expiry, permissions)
	c.Assert(err, chk.IsNil)

	resp, err := http.Get(sasURI)
	c.Assert(err, chk.IsNil)

	blobResp, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	c.Assert(err, chk.IsNil)

	c.Assert(resp.StatusCode, chk.Equals, http.StatusOK)
	c.Assert(len(blobResp), chk.Equals, len(body))
}

func (s *StorageBlobSuite) TestBlobExists(c *chk.C) {
	blob := randName(5)

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte("Hello!")), chk.IsNil)
	defer cli.DeleteBlob(cnt.Name, blob, nil)

	ok, err := cli.BlobExists(cnt.Name, blob+".foo")
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	ok, err = cli.BlobExists(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)
}

func (s *StorageBlobSuite) TestGetBlobURL(c *chk.C) {
	api, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	cli := api.GetBlobService()

	c.Assert(cli.GetBlobURL("c", "nested/blob"), chk.Equals, "https://foo.blob.core.windows.net/c/nested/blob")
	c.Assert(cli.GetBlobURL("", "blob"), chk.Equals, "https://foo.blob.core.windows.net/$root/blob")
	c.Assert(cli.GetBlobURL("", "nested/blob"), chk.Equals, "https://foo.blob.core.windows.net/$root/nested/blob")
}

func (s *StorageBlobSuite) TestGetBlobContainerURL(c *chk.C) {
	api, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)
	cli := api.GetBlobService()

	c.Assert(cli.GetBlobURL("c", ""), chk.Equals, "https://foo.blob.core.windows.net/c")
	c.Assert(cli.GetBlobURL("", ""), chk.Equals, "https://foo.blob.core.windows.net/$root")
}

func (s *StorageBlobSuite) TestBlobCopy(c *chk.C) {
	if testing.Short() {
		c.Skip("skipping blob copy in short mode, no SLA on async operation")
	}

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	src := randName(5)
	dst := randName(5)
	body := []byte(randString(1024))

	c.Assert(cli.putSingleBlockBlob(cnt.Name, src, body), chk.IsNil)
	defer cli.DeleteBlob(cnt.Name, src, nil)

	c.Assert(cli.CopyBlob(cnt.Name, dst, cli.GetBlobURL(cnt.Name, src)), chk.IsNil)
	defer cli.DeleteBlob(cnt.Name, dst, nil)

	blobBody, err := cli.GetBlob(cnt.Name, dst)
	c.Assert(err, chk.IsNil)

	b, err := ioutil.ReadAll(blobBody)
	defer blobBody.Close()
	c.Assert(err, chk.IsNil)
	c.Assert(b, chk.DeepEquals, body)
}

func (s *StorageBlobSuite) TestStartBlobCopy(c *chk.C) {
	if testing.Short() {
		c.Skip("skipping blob copy in short mode, no SLA on async operation")
	}

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	src := randName(5)
	dst := randName(5)
	body := []byte(randString(1024))

	c.Assert(cli.putSingleBlockBlob(cnt.Name, src, body), chk.IsNil)
	defer cli.DeleteBlob(cnt.Name, src, nil)

	// given we dont know when it will start, can we even test destination creation?
	// will just test that an error wasn't thrown for now.
	copyID, err := cli.StartBlobCopy(cnt.Name, dst, cli.GetBlobURL(cnt.Name, src))
	c.Assert(copyID, chk.NotNil)
	c.Assert(err, chk.IsNil)
}

// Tests abort of blobcopy. Given the blobcopy is usually over before we can actually trigger an abort
// it is agreed that we perform a copy then try and perform an abort. It should result in a HTTP status of 409.
// So basically we're testing negative scenario (as good as we can do for now)
func (s *StorageBlobSuite) TestAbortBlobCopy(c *chk.C) {
	if testing.Short() {
		c.Skip("skipping blob copy in short mode, no SLA on async operation")
	}

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	src := randName(5)
	dst := randName(5)
	body := []byte(randString(1024))

	c.Assert(cli.putSingleBlockBlob(cnt.Name, src, body), chk.IsNil)
	defer cli.DeleteBlob(cnt.Name, src, nil)

	// given we dont know when it will start, can we even test destination creation?
	// will just test that an error wasn't thrown for now.
	copyID, err := cli.StartBlobCopy(cnt.Name, dst, cli.GetBlobURL(cnt.Name, src))
	c.Assert(copyID, chk.NotNil)
	c.Assert(err, chk.IsNil)

	err = cli.WaitForBlobCopy(cnt.Name, dst, copyID)
	c.Assert(err, chk.IsNil)

	// abort abort abort, but we *know* its already completed.
	err = cli.AbortBlobCopy(cnt.Name, dst, copyID, "", 0)

	// abort should fail (over already)
	c.Assert(err.(AzureStorageServiceError).StatusCode, chk.Equals, http.StatusConflict)
}

func (s *StorageBlobSuite) TestDeleteBlobIfExists(c *chk.C) {
	blob := randName(5)

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	c.Assert(cli.DeleteBlob(cnt.Name, blob, nil), chk.NotNil)

	ok, err := cli.DeleteBlobIfExists(cnt.Name, blob, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)
}

func (s *StorageBlobSuite) TestDeleteBlobWithConditions(c *chk.C) {
	blob := randName(5)

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	c.Assert(cli.CreateBlockBlob(cnt.Name, blob), chk.IsNil)
	oldProps, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)

	// Update metadata, so Etag changes
	c.Assert(cli.SetBlobMetadata(cnt.Name, blob, map[string]string{}, nil), chk.IsNil)
	newProps, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)

	// "Delete if matches old Etag" should fail without deleting.
	err = cli.DeleteBlob(cnt.Name, blob, map[string]string{
		"If-Match": oldProps.Etag,
	})
	c.Assert(err, chk.FitsTypeOf, AzureStorageServiceError{})
	c.Assert(err.(AzureStorageServiceError).StatusCode, chk.Equals, http.StatusPreconditionFailed)
	_, err = cli.GetBlob(cnt.Name, blob)
	c.Assert(err, chk.IsNil)

	// "Delete if matches new Etag" should succeed.
	err = cli.DeleteBlob(cnt.Name, blob, map[string]string{
		"If-Match": newProps.Etag,
	})
	c.Assert(err, chk.IsNil)
	_, err = cli.GetBlob(cnt.Name, blob)
	c.Assert(err, chk.Not(chk.IsNil))
}

func (s *StorageBlobSuite) TestGetBlobProperties(c *chk.C) {
	blob := randName(5)
	contents := randString(64)

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	// Nonexisting blob
	_, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.NotNil)

	// Put the blob
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte(contents)), chk.IsNil)

	// Get blob properties
	props, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)

	c.Assert(props.ContentLength, chk.Equals, int64(len(contents)))
	c.Assert(props.ContentType, chk.Equals, "application/octet-stream")
	c.Assert(props.BlobType, chk.Equals, BlobTypeBlock)
}

// Ensure it's possible to generate a ListBlobs response with
// metadata, e.g., for a stub server.
func (s *StorageBlobSuite) TestMarshalBlobMetadata(c *chk.C) {
	buf, err := xml.Marshal(Blob{
		Name:       randName(5),
		Properties: BlobProperties{},
		Metadata:   BlobMetadata{"foo": "baz < waz"},
	})
	c.Assert(err, chk.IsNil)
	c.Assert(string(buf), chk.Matches, `.*<Metadata><Foo>baz &lt; waz</Foo></Metadata>.*`)
}

func (s *StorageBlobSuite) TestGetAndSetMetadata(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	m, err := cli.GetBlobMetadata(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(m, chk.Not(chk.Equals), nil)
	c.Assert(len(m), chk.Equals, 0)

	mPut := map[string]string{
		"foo":     "bar",
		"bar_baz": "waz qux",
	}

	err = cli.SetBlobMetadata(cnt.Name, blob, mPut, nil)
	c.Assert(err, chk.IsNil)

	m, err = cli.GetBlobMetadata(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Check(m, chk.DeepEquals, mPut)

	// Case munging

	mPutUpper := map[string]string{
		"Foo":     "different bar",
		"bar_BAZ": "different waz qux",
	}
	mExpectLower := map[string]string{
		"foo":     "different bar",
		"bar_baz": "different waz qux",
	}

	err = cli.SetBlobMetadata(cnt.Name, blob, mPutUpper, nil)
	c.Assert(err, chk.IsNil)

	m, err = cli.GetBlobMetadata(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Check(m, chk.DeepEquals, mExpectLower)
}

func (s *StorageBlobSuite) TestSetMetadataWithExtraHeaders(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	mPut := map[string]string{
		"foo":     "bar",
		"bar_baz": "waz qux",
	}

	extraHeaders := map[string]string{
		"If-Match": "incorrect-etag",
	}

	// Set with incorrect If-Match in extra headers should result in error
	err := cli.SetBlobMetadata(cnt.Name, blob, mPut, extraHeaders)
	c.Assert(err, chk.NotNil)

	props, err := cli.GetBlobProperties(cnt.Name, blob)
	extraHeaders = map[string]string{
		"If-Match": props.Etag,
	}

	// Set with matching If-Match in extra headers should succeed
	err = cli.SetBlobMetadata(cnt.Name, blob, mPut, extraHeaders)
	c.Assert(err, chk.IsNil)
}

func (s *StorageBlobSuite) TestSetBlobProperties(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	mPut := BlobHeaders{
		CacheControl:    "private, max-age=0, no-cache",
		ContentMD5:      "oBATU+oaDduHWbVZLuzIJw==",
		ContentType:     "application/json",
		ContentEncoding: "gzip",
		ContentLanguage: "de-DE",
	}

	err := cli.SetBlobProperties(cnt.Name, blob, mPut)
	c.Assert(err, chk.IsNil)

	props, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)

	c.Check(mPut.CacheControl, chk.Equals, props.CacheControl)
	c.Check(mPut.ContentType, chk.Equals, props.ContentType)
	c.Check(mPut.ContentMD5, chk.Equals, props.ContentMD5)
	c.Check(mPut.ContentEncoding, chk.Equals, props.ContentEncoding)
	c.Check(mPut.ContentLanguage, chk.Equals, props.ContentLanguage)
}

func (s *StorageBlobSuite) TestSnapshotBlob(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	snapshotTime, err := cli.SnapshotBlob(cnt.Name, blob, 0, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(snapshotTime, chk.NotNil)
}

func (s *StorageBlobSuite) TestSnapshotBlobWithTimeout(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	snapshotTime, err := cli.SnapshotBlob(cnt.Name, blob, 30, nil)
	c.Assert(err, chk.IsNil)
	c.Assert(snapshotTime, chk.NotNil)
}

func (s *StorageBlobSuite) TestSnapshotBlobWithValidLease(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	// generate lease.
	currentLeaseID, err := cli.AcquireLease(cnt.Name, blob, 30, "")
	c.Assert(err, chk.IsNil)

	extraHeaders := map[string]string{
		headerLeaseID: currentLeaseID,
	}

	snapshotTime, err := cli.SnapshotBlob(cnt.Name, blob, 0, extraHeaders)
	c.Assert(err, chk.IsNil)
	c.Assert(snapshotTime, chk.NotNil)
}

func (s *StorageBlobSuite) TestSnapshotBlobWithInvalidLease(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	// generate lease.
	leaseID, err := cli.AcquireLease(cnt.Name, blob, 30, "")
	c.Assert(err, chk.IsNil)
	c.Assert(leaseID, chk.Not(chk.Equals), "")

	extraHeaders := map[string]string{
		headerLeaseID: "GolangRocksOnAzure",
	}

	snapshotTime, err := cli.SnapshotBlob(cnt.Name, blob, 0, extraHeaders)
	c.Assert(err, chk.NotNil)
	c.Assert(snapshotTime, chk.IsNil)
}

func (s *StorageBlobSuite) TestAcquireLeaseWithNoProposedLeaseID(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	_, err := cli.AcquireLease(cnt.Name, blob, 30, "")
	c.Assert(err, chk.IsNil)
}

func (s *StorageBlobSuite) TestAcquireLeaseWithProposedLeaseID(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	proposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fea"
	leaseID, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.IsNil)
	c.Assert(leaseID, chk.Equals, proposedLeaseID)
}

func (s *StorageBlobSuite) TestAcquireLeaseWithBadProposedLeaseID(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	proposedLeaseID := "badbadbad"
	_, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.NotNil)
}

func (s *StorageBlobSuite) TestRenewLeaseSuccessful(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	proposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fea"
	leaseID, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.IsNil)

	err = cli.RenewLease(cnt.Name, blob, leaseID)
	c.Assert(err, chk.IsNil)
}

func (s *StorageBlobSuite) TestRenewLeaseAgainstNoCurrentLease(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	badLeaseID := "1f812371-a41d-49e6-b123-f4b542e85144"
	err := cli.RenewLease(cnt.Name, blob, badLeaseID)
	c.Assert(err, chk.NotNil)
}

func (s *StorageBlobSuite) TestChangeLeaseSuccessful(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)
	proposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fea"
	leaseID, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.IsNil)

	newProposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fbb"
	newLeaseID, err := cli.ChangeLease(cnt.Name, blob, leaseID, newProposedLeaseID)
	c.Assert(err, chk.IsNil)
	c.Assert(newLeaseID, chk.Equals, newProposedLeaseID)
}

func (s *StorageBlobSuite) TestChangeLeaseNotSuccessfulbadProposedLeaseID(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)
	proposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fea"
	leaseID, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.IsNil)

	newProposedLeaseID := "1f812371-a41d-49e6-b123-f4b542e"
	_, err = cli.ChangeLease(cnt.Name, blob, leaseID, newProposedLeaseID)
	c.Assert(err, chk.NotNil)
}

func (s *StorageBlobSuite) TestReleaseLeaseSuccessful(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)
	proposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fea"
	leaseID, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.IsNil)

	err = cli.ReleaseLease(cnt.Name, blob, leaseID)
	c.Assert(err, chk.IsNil)
}

func (s *StorageBlobSuite) TestReleaseLeaseNotSuccessfulBadLeaseID(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)
	proposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fea"
	_, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.IsNil)

	err = cli.ReleaseLease(cnt.Name, blob, "badleaseid")
	c.Assert(err, chk.NotNil)
}

func (s *StorageBlobSuite) TestBreakLeaseSuccessful(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	proposedLeaseID := "dfe6dde8-68d5-4910-9248-c97c61768fea"
	_, err := cli.AcquireLease(cnt.Name, blob, 30, proposedLeaseID)
	c.Assert(err, chk.IsNil)

	_, err = cli.BreakLease(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
}

func (s *StorageBlobSuite) TestPutEmptyBlockBlob(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte{}), chk.IsNil)

	props, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(props.ContentLength, chk.Not(chk.Equals), 0)
}

func (s *StorageBlobSuite) TestGetBlobRange(c *chk.C) {
	blob := randName(5)
	body := "0123456789"

	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	c.Assert(cli.putSingleBlockBlob(cnt.Name, blob, []byte(body)), chk.IsNil)
	defer cli.DeleteBlob(cnt.Name, blob, nil)

	// Read 1-3
	for _, r := range []struct {
		rangeStr string
		expected string
	}{
		{"0-", body},
		{"1-3", body[1 : 3+1]},
		{"3-", body[3:]},
	} {
		resp, err := cli.GetBlobRange(cnt.Name, blob, r.rangeStr, nil)
		c.Assert(err, chk.IsNil)
		blobBody, err := ioutil.ReadAll(resp)
		c.Assert(err, chk.IsNil)

		str := string(blobBody)
		c.Assert(str, chk.Equals, r.expected)
	}
}

func (s *StorageBlobSuite) TestCreateBlockBlobFromReader(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	name := randName(5)
	data := randBytes(8888)
	c.Assert(cli.CreateBlockBlobFromReader(cnt.Name, name, uint64(len(data)), bytes.NewReader(data), nil), chk.IsNil)

	body, err := cli.GetBlob(cnt.Name, name)
	c.Assert(err, chk.IsNil)
	gotData, err := ioutil.ReadAll(body)
	body.Close()

	c.Assert(err, chk.IsNil)
	c.Assert(gotData, chk.DeepEquals, data)
}

func (s *StorageBlobSuite) TestCreateBlockBlobFromReaderWithShortData(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	name := randName(5)
	data := randBytes(8888)
	err := cli.CreateBlockBlobFromReader(cnt.Name, name, 9999, bytes.NewReader(data), nil)
	c.Assert(err, chk.Not(chk.IsNil))

	_, err = cli.GetBlob(cnt.Name, name)
	// Upload was incomplete: blob should not have been created.
	c.Assert(err, chk.Not(chk.IsNil))
}

func (s *StorageBlobSuite) TestPutBlock(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	chunk := []byte(randString(1024))
	blockID := base64.StdEncoding.EncodeToString([]byte("foo"))
	c.Assert(cli.PutBlock(cnt.Name, blob, blockID, chunk), chk.IsNil)
}

func (s *StorageBlobSuite) TestGetBlockList_PutBlockList(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	chunk := []byte(randString(1024))
	blockID := base64.StdEncoding.EncodeToString([]byte("foo"))

	// Put one block
	c.Assert(cli.PutBlock(cnt.Name, blob, blockID, chunk), chk.IsNil)
	defer cli.deleteBlob(cnt.Name, blob, nil)

	// Get committed blocks
	committed, err := cli.GetBlockList(cnt.Name, blob, BlockListTypeCommitted)
	c.Assert(err, chk.IsNil)

	if len(committed.CommittedBlocks) > 0 {
		c.Fatal("There are committed blocks")
	}

	// Get uncommitted blocks
	uncommitted, err := cli.GetBlockList(cnt.Name, blob, BlockListTypeUncommitted)
	c.Assert(err, chk.IsNil)

	c.Assert(len(uncommitted.UncommittedBlocks), chk.Equals, 1)
	// Commit block list
	c.Assert(cli.PutBlockList(cnt.Name, blob, []Block{{blockID, BlockStatusUncommitted}}), chk.IsNil)

	// Get all blocks
	all, err := cli.GetBlockList(cnt.Name, blob, BlockListTypeAll)
	c.Assert(err, chk.IsNil)
	c.Assert(len(all.CommittedBlocks), chk.Equals, 1)
	c.Assert(len(all.UncommittedBlocks), chk.Equals, 0)

	// Verify the block
	thatBlock := all.CommittedBlocks[0]
	c.Assert(thatBlock.Name, chk.Equals, blockID)
	c.Assert(thatBlock.Size, chk.Equals, int64(len(chunk)))
}

func (s *StorageBlobSuite) TestCreateBlockBlob(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.CreateBlockBlob(cnt.Name, blob), chk.IsNil)

	// Verify
	blocks, err := cli.GetBlockList(cnt.Name, blob, BlockListTypeAll)
	c.Assert(err, chk.IsNil)
	c.Assert(len(blocks.CommittedBlocks), chk.Equals, 0)
	c.Assert(len(blocks.UncommittedBlocks), chk.Equals, 0)
}

func (s *StorageBlobSuite) TestPutPageBlob(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	size := int64(10 * 1024 * 1024)
	c.Assert(cli.PutPageBlob(cnt.Name, blob, size, nil), chk.IsNil)

	// Verify
	props, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(props.ContentLength, chk.Equals, size)
	c.Assert(props.BlobType, chk.Equals, BlobTypePage)
}

func (s *StorageBlobSuite) TestPutPagesUpdate(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	size := int64(10 * 1024 * 1024) // larger than we'll use
	c.Assert(cli.PutPageBlob(cnt.Name, blob, size, nil), chk.IsNil)

	chunk1 := []byte(randString(1024))
	chunk2 := []byte(randString(512))

	// Append chunks
	c.Assert(cli.PutPage(cnt.Name, blob, 0, int64(len(chunk1)-1), PageWriteTypeUpdate, chunk1, nil), chk.IsNil)
	c.Assert(cli.PutPage(cnt.Name, blob, int64(len(chunk1)), int64(len(chunk1)+len(chunk2)-1), PageWriteTypeUpdate, chunk2, nil), chk.IsNil)

	// Verify contents
	out, err := cli.GetBlobRange(cnt.Name, blob, fmt.Sprintf("%v-%v", 0, len(chunk1)+len(chunk2)-1), nil)
	c.Assert(err, chk.IsNil)
	defer out.Close()
	blobContents, err := ioutil.ReadAll(out)
	c.Assert(err, chk.IsNil)
	c.Assert(blobContents, chk.DeepEquals, append(chunk1, chunk2...))
	out.Close()

	// Overwrite first half of chunk1
	chunk0 := []byte(randString(512))
	c.Assert(cli.PutPage(cnt.Name, blob, 0, int64(len(chunk0)-1), PageWriteTypeUpdate, chunk0, nil), chk.IsNil)

	// Verify contents
	out, err = cli.GetBlobRange(cnt.Name, blob, fmt.Sprintf("%v-%v", 0, len(chunk1)+len(chunk2)-1), nil)
	c.Assert(err, chk.IsNil)
	defer out.Close()
	blobContents, err = ioutil.ReadAll(out)
	c.Assert(err, chk.IsNil)
	c.Assert(blobContents, chk.DeepEquals, append(append(chunk0, chunk1[512:]...), chunk2...))
}

func (s *StorageBlobSuite) TestPutPagesClear(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	size := int64(10 * 1024 * 1024) // larger than we'll use
	c.Assert(cli.PutPageBlob(cnt.Name, blob, size, nil), chk.IsNil)

	// Put 0-2047
	chunk := []byte(randString(2048))
	c.Assert(cli.PutPage(cnt.Name, blob, 0, 2047, PageWriteTypeUpdate, chunk, nil), chk.IsNil)

	// Clear 512-1023
	c.Assert(cli.PutPage(cnt.Name, blob, 512, 1023, PageWriteTypeClear, nil, nil), chk.IsNil)

	// Verify contents
	out, err := cli.GetBlobRange(cnt.Name, blob, "0-2047", nil)
	c.Assert(err, chk.IsNil)
	contents, err := ioutil.ReadAll(out)
	c.Assert(err, chk.IsNil)
	defer out.Close()
	c.Assert(contents, chk.DeepEquals, append(append(chunk[:512], make([]byte, 512)...), chunk[1024:]...))
}

func (s *StorageBlobSuite) TestGetPageRanges(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	size := int64(10 * 1024 * 1024) // larger than we'll use
	c.Assert(cli.PutPageBlob(cnt.Name, blob, size, nil), chk.IsNil)

	// Get page ranges on empty blob
	out, err := cli.GetPageRanges(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(len(out.PageList), chk.Equals, 0)

	// Add 0-512 page
	c.Assert(cli.PutPage(cnt.Name, blob, 0, 511, PageWriteTypeUpdate, []byte(randString(512)), nil), chk.IsNil)

	out, err = cli.GetPageRanges(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(len(out.PageList), chk.Equals, 1)

	// Add 1024-2048
	c.Assert(cli.PutPage(cnt.Name, blob, 1024, 2047, PageWriteTypeUpdate, []byte(randString(1024)), nil), chk.IsNil)

	out, err = cli.GetPageRanges(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(len(out.PageList), chk.Equals, 2)
}

func (s *StorageBlobSuite) TestPutAppendBlob(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.PutAppendBlob(cnt.Name, blob, nil), chk.IsNil)

	// Verify
	props, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(props.ContentLength, chk.Equals, int64(0))
	c.Assert(props.BlobType, chk.Equals, BlobTypeAppend)
}

func (s *StorageBlobSuite) TestPutAppendBlobAppendBlocks(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randName(5)
	c.Assert(cli.PutAppendBlob(cnt.Name, blob, nil), chk.IsNil)

	chunk1 := []byte(randString(1024))
	chunk2 := []byte(randString(512))

	// Append first block
	c.Assert(cli.AppendBlock(cnt.Name, blob, chunk1, nil), chk.IsNil)

	// Verify contents
	out, err := cli.GetBlobRange(cnt.Name, blob, fmt.Sprintf("%v-%v", 0, len(chunk1)-1), nil)
	c.Assert(err, chk.IsNil)
	defer out.Close()
	blobContents, err := ioutil.ReadAll(out)
	c.Assert(err, chk.IsNil)
	c.Assert(blobContents, chk.DeepEquals, chunk1)
	out.Close()

	// Append second block
	c.Assert(cli.AppendBlock(cnt.Name, blob, chunk2, nil), chk.IsNil)

	// Verify contents
	out, err = cli.GetBlobRange(cnt.Name, blob, fmt.Sprintf("%v-%v", 0, len(chunk1)+len(chunk2)-1), nil)
	c.Assert(err, chk.IsNil)
	defer out.Close()
	blobContents, err = ioutil.ReadAll(out)
	c.Assert(err, chk.IsNil)
	c.Assert(blobContents, chk.DeepEquals, append(chunk1, chunk2...))
	out.Close()
}

func (b BlobStorageClient) putSingleBlockBlob(container, name string, chunk []byte) error {
	if len(chunk) > MaxBlobBlockSize {
		return fmt.Errorf("storage: provided chunk (%d bytes) cannot fit into single-block blob (max %d bytes)", len(chunk), MaxBlobBlockSize)
	}

	uri := b.client.getEndpoint(blobServiceName, pathForBlob(container, name), url.Values{})
	headers := b.client.getStandardHeaders()
	headers["x-ms-blob-type"] = string(BlobTypeBlock)
	headers["Content-Length"] = fmt.Sprintf("%v", len(chunk))

	resp, err := b.client.exec(http.MethodPut, uri, headers, bytes.NewReader(chunk), b.auth)
	if err != nil {
		return err
	}
	return checkRespCode(resp.statusCode, []int{http.StatusCreated})
}

func (s *StorageBlobSuite) TestPutAppendBlobSpecialChars(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blob := randNameWithSpecialChars(5)
	c.Assert(cli.PutAppendBlob(cnt.Name, blob, nil), chk.IsNil)

	// Verify metadata
	props, err := cli.GetBlobProperties(cnt.Name, blob)
	c.Assert(err, chk.IsNil)
	c.Assert(props.ContentLength, chk.Equals, int64(0))
	c.Assert(props.BlobType, chk.Equals, BlobTypeAppend)

	chunk1 := []byte(randString(1024))
	chunk2 := []byte(randString(512))

	// Append first block
	c.Assert(cli.AppendBlock(cnt.Name, blob, chunk1, nil), chk.IsNil)

	// Verify contents
	out, err := cli.GetBlobRange(cnt.Name, blob, fmt.Sprintf("%v-%v", 0, len(chunk1)-1), nil)
	c.Assert(err, chk.IsNil)
	defer out.Close()
	blobContents, err := ioutil.ReadAll(out)
	c.Assert(err, chk.IsNil)
	c.Assert(blobContents, chk.DeepEquals, chunk1)
	out.Close()

	// Append second block
	c.Assert(cli.AppendBlock(cnt.Name, blob, chunk2, nil), chk.IsNil)

	// Verify contents
	out, err = cli.GetBlobRange(cnt.Name, blob, fmt.Sprintf("%v-%v", 0, len(chunk1)+len(chunk2)-1), nil)
	c.Assert(err, chk.IsNil)
	defer out.Close()
	blobContents, err = ioutil.ReadAll(out)
	c.Assert(err, chk.IsNil)
	c.Assert(blobContents, chk.DeepEquals, append(chunk1, chunk2...))
	out.Close()
}

func randBytes(n int) []byte {
	data := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, data); err != nil {
		panic(err)
	}
	return data
}

func randName(n int) string {
	name := randString(n) + "/" + randString(n)
	return name
}

func randNameWithSpecialChars(n int) string {
	name := randString(n) + "/" + randString(n) + "-._~:?#[]@!$&'()*,;+= " + randString(n)
	return name
}
