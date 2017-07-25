package storage

import (
	"encoding/base64"
	"net/http"

	chk "gopkg.in/check.v1"
)

type AuthorizationSuite struct{}

var _ = chk.Suite(&AuthorizationSuite{})

func (a *AuthorizationSuite) Test_addAuthorizationHeader(c *chk.C) {
	cli, err := NewBasicClient("mindgotest", "zHDHGs7C+Di9pZSDMuarxJJz3xRBzAHBYaobxpLEc7kwTptR/hPEa9j93hIfb2Tbe9IA50MViGmjQ6nUF/OVvA==")
	c.Assert(err, chk.IsNil)
	cli.UseSharedKeyLite = true
	tableCli := cli.GetTableService()

	headers := map[string]string{
		"Accept-Charset":    "UTF-8",
		headerContentType:   "application/json",
		headerXmsDate:       "Wed, 23 Sep 2015 16:40:05 GMT",
		headerContentLength: "0",
		headerXmsVersion:    "2015-02-21",
		"Accept":            "application/json;odata=nometadata",
	}
	url := "https://mindgotest.table.core.windows.net/tquery()"
	headers, err = tableCli.client.addAuthorizationHeader("", url, headers, tableCli.auth)
	c.Assert(err, chk.IsNil)

	c.Assert(headers[headerAuthorization], chk.Equals, "SharedKeyLite mindgotest:+32DTgsPUgXPo/O7RYaTs0DllA6FTXMj3uK4Qst8y/E=")
}

func (a *AuthorizationSuite) Test_getSharedKey(c *chk.C) {
	// Shared Key Lite for Tables
	cli, err := NewBasicClient("mindgotest", "zHDHGs7C+Di9pZSDMuarxJJz3xRBzAHBYaobxpLEc7kwTptR/hPEa9j93hIfb2Tbe9IA50MViGmjQ6nUF/OVvA==")
	c.Assert(err, chk.IsNil)

	headers := map[string]string{
		"Accept-Charset":    "UTF-8",
		headerContentType:   "application/json",
		headerXmsDate:       "Wed, 23 Sep 2015 16:40:05 GMT",
		headerContentLength: "0",
		headerXmsVersion:    "2015-02-21",
		"Accept":            "application/json;odata=nometadata",
	}
	url := "https://mindgotest.table.core.windows.net/tquery()"

	key, err := cli.getSharedKey("", url, headers, sharedKeyLiteForTable)
	c.Assert(err, chk.IsNil)
	c.Assert(key, chk.Equals, "SharedKeyLite mindgotest:+32DTgsPUgXPo/O7RYaTs0DllA6FTXMj3uK4Qst8y/E=")
}

func (a *AuthorizationSuite) Test_buildCanonicalizedResource(c *chk.C) {
	cli, err := NewBasicClient("foo", "YmFy")
	c.Assert(err, chk.IsNil)

	type test struct {
		url      string
		auth     authentication
		expected string
	}
	tests := []test{
		// Shared Key
		{"https://foo.blob.core.windows.net/path?a=b&c=d", sharedKey, "/foo/path\na:b\nc:d"},
		{"https://foo.blob.core.windows.net/?comp=list", sharedKey, "/foo/\ncomp:list"},
		{"https://foo.blob.core.windows.net/cnt/blob", sharedKey, "/foo/cnt/blob"},
		{"https://foo.blob.core.windows.net/cnt/bl ob", sharedKey, "/foo/cnt/bl%20ob"},
		{"https://foo.blob.core.windows.net/c nt/blob", sharedKey, "/foo/c%20nt/blob"},
		{"https://foo.blob.core.windows.net/cnt/blob%3F%23%5B%5D%21$&%27%28%29%2A blob", sharedKey, "/foo/cnt/blob%3F%23%5B%5D%21$&%27%28%29%2A%20blob"},
		{"https://foo.blob.core.windows.net/cnt/blob-._~:,@;+=blob", sharedKey, "/foo/cnt/blob-._~:,@;+=blob"},
		{"https://foo.blob.core.windows.net/c nt/blob-._~:%3F%23%5B%5D@%21$&%27%28%29%2A,;+=/blob", sharedKey, "/foo/c%20nt/blob-._~:%3F%23%5B%5D@%21$&%27%28%29%2A,;+=/blob"},
		// Shared Key Lite for Table
		{"https://foo.table.core.windows.net/mytable", sharedKeyLiteForTable, "/foo/mytable"},
		{"https://foo.table.core.windows.net/mytable?comp=acl", sharedKeyLiteForTable, "/foo/mytable?comp=acl"},
		{"https://foo.table.core.windows.net/mytable?comp=acl&timeout=10", sharedKeyForTable, "/foo/mytable?comp=acl"},
		{"https://foo.table.core.windows.net/mytable(PartitionKey='pkey',RowKey='rowkey%3D')", sharedKeyForTable, "/foo/mytable(PartitionKey='pkey',RowKey='rowkey%3D')"},
	}

	for _, t := range tests {
		out, err := cli.buildCanonicalizedResource(t.url, t.auth)
		c.Assert(err, chk.IsNil)
		c.Assert(out, chk.Equals, t.expected)
	}
}

func (a *AuthorizationSuite) Test_buildCanonicalizedString(c *chk.C) {
	var tests = []struct {
		verb                  string
		headers               map[string]string
		canonicalizedResource string
		auth                  authentication
		out                   string
	}{
		{
			// Shared Key
			verb: http.MethodGet,
			headers: map[string]string{
				headerXmsDate:    "Sun, 11 Oct 2009 21:49:13 GMT",
				headerXmsVersion: "2009-09-19",
			},
			canonicalizedResource: "/myaccount/ mycontainer\ncomp:metadata\nrestype:container\ntimeout:20",
			auth: sharedKey,
			out:  "GET\n\n\n\n\n\n\n\n\n\n\n\nx-ms-date:Sun, 11 Oct 2009 21:49:13 GMT\nx-ms-version:2009-09-19\n/myaccount/ mycontainer\ncomp:metadata\nrestype:container\ntimeout:20",
		},
		{
			// Shared Key for Tables
			verb: http.MethodPut,
			headers: map[string]string{
				headerContentType: "text/plain; charset=UTF-8",
				headerDate:        "Sun, 11 Oct 2009 19:52:39 GMT",
			},
			canonicalizedResource: "/testaccount1/Tables",
			auth: sharedKeyForTable,
			out:  "PUT\n\ntext/plain; charset=UTF-8\nSun, 11 Oct 2009 19:52:39 GMT\n/testaccount1/Tables",
		},
		{
			// Shared Key Lite
			verb: http.MethodPut,
			headers: map[string]string{
				headerContentType: "text/plain; charset=UTF-8",
				headerXmsDate:     "Sun, 20 Sep 2009 20:36:40 GMT",
				"x-ms-meta-m1":    "v1",
				"x-ms-meta-m2":    "v2",
			},
			canonicalizedResource: "/testaccount1/mycontainer/hello.txt",
			auth: sharedKeyLite,
			out:  "PUT\n\ntext/plain; charset=UTF-8\n\nx-ms-date:Sun, 20 Sep 2009 20:36:40 GMT\nx-ms-meta-m1:v1\nx-ms-meta-m2:v2\n/testaccount1/mycontainer/hello.txt",
		},
		{
			// Shared Key Lite for Tables
			verb: "",
			headers: map[string]string{
				headerDate: "Sun, 11 Oct 2009 19:52:39 GMT",
			},
			canonicalizedResource: "/testaccount1/Tables",
			auth: sharedKeyLiteForTable,
			out:  "Sun, 11 Oct 2009 19:52:39 GMT\n/testaccount1/Tables",
		},
	}

	for _, t := range tests {
		canonicalizedString, err := buildCanonicalizedString(t.verb, t.headers, t.canonicalizedResource, t.auth)
		c.Assert(err, chk.IsNil)
		c.Assert(canonicalizedString, chk.Equals, t.out)
	}
}

func (a *AuthorizationSuite) Test_buildCanonicalizedHeader(c *chk.C) {
	type test struct {
		headers  map[string]string
		expected string
	}
	tests := []test{
		{map[string]string{},
			""},
		{map[string]string{
			"x-ms-foo": "bar"},
			"x-ms-foo:bar"},
		{map[string]string{
			"foo:": "bar"},
			""},
		{map[string]string{
			"foo:":     "bar",
			"x-ms-foo": "bar"},
			"x-ms-foo:bar"},
		{map[string]string{
			"x-ms-version":   "9999-99-99",
			"x-ms-blob-type": "BlockBlob"},
			"x-ms-blob-type:BlockBlob\nx-ms-version:9999-99-99"}}

	for _, i := range tests {
		c.Assert(buildCanonicalizedHeader(i.headers), chk.Equals, i.expected)
	}
}

func (a *AuthorizationSuite) Test_createAuthorizationHeader(c *chk.C) {
	cli, err := NewBasicClient("foo", base64.StdEncoding.EncodeToString([]byte("bar")))
	c.Assert(err, chk.IsNil)

	canonicalizedString := `foobarzoo`

	c.Assert(cli.createAuthorizationHeader(canonicalizedString, sharedKey),
		chk.Equals, `SharedKey foo:h5U0ATVX6SpbFX1H6GNuxIMeXXCILLoIvhflPtuQZ30=`)
	c.Assert(cli.createAuthorizationHeader(canonicalizedString, sharedKeyLite),
		chk.Equals, `SharedKeyLite foo:h5U0ATVX6SpbFX1H6GNuxIMeXXCILLoIvhflPtuQZ30=`)
}

func (a *AuthorizationSuite) Test_allSharedKeys(c *chk.C) {
	cli := getBasicClient(c)

	blobCli := cli.GetBlobService()
	tableCli := cli.GetTableService()

	cnt1 := blobCli.GetContainerReference(randContainer())
	cnt2 := blobCli.GetContainerReference(randContainer())

	tn1 := AzureTable(randTable())
	tn2 := AzureTable(randTable())

	// Shared Key
	c.Assert(blobCli.auth, chk.Equals, sharedKey)
	c.Assert(cnt1.Create(), chk.IsNil)
	c.Assert(cnt1.Delete(), chk.IsNil)

	// Shared Key for Tables
	c.Assert(tableCli.auth, chk.Equals, sharedKeyForTable)
	c.Assert(tableCli.CreateTable(tn1), chk.IsNil)
	c.Assert(tableCli.DeleteTable(tn1), chk.IsNil)

	// Change to Lite
	cli.UseSharedKeyLite = true
	blobCli = cli.GetBlobService()
	tableCli = cli.GetTableService()

	// Shared Key Lite
	c.Assert(blobCli.auth, chk.Equals, sharedKeyLite)
	c.Assert(cnt2.Create(), chk.IsNil)
	c.Assert(cnt2.Delete(), chk.IsNil)

	// Shared Key Lite for Tables
	c.Assert(tableCli.auth, chk.Equals, sharedKeyLiteForTable)
	c.Assert(tableCli.CreateTable(tn2), chk.IsNil)
	c.Assert(tableCli.DeleteTable(tn2), chk.IsNil)
}
