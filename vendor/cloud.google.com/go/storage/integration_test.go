// Copyright 2014 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package storage

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	gax "github.com/googleapis/gax-go"

	"golang.org/x/net/context"

	"cloud.google.com/go/internal"
	"cloud.google.com/go/internal/testutil"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	itesting "google.golang.org/api/iterator/testing"
	"google.golang.org/api/option"
)

const testPrefix = "-go-cloud-storage-test"

// suffix is a timestamp-based suffix which is added to all buckets created by
// tests. This reduces flakiness when the tests are run in parallel and allows
// automatic cleaning up of artifacts left when tests fail.
var suffix = fmt.Sprintf("%s-%d", testPrefix, time.Now().UnixNano())

func TestMain(m *testing.M) {
	integrationTest := initIntegrationTest()
	exit := m.Run()
	if integrationTest {
		if err := cleanup(); err != nil {
			// No need to be loud if cleanup() fails; we'll get
			// any undeleted buckets next time.
			log.Printf("Post-test cleanup failed: %v\n", err)
		}
	}
	os.Exit(exit)
}

// If integration tests will be run, create a unique bucket for them.
func initIntegrationTest() bool {
	flag.Parse() // needed for testing.Short()
	ctx := context.Background()
	if testing.Short() {
		return false
	}
	client, bucket := config(ctx)
	if client == nil {
		return false
	}
	defer client.Close()
	if err := client.Bucket(bucket).Create(ctx, testutil.ProjID(), nil); err != nil {
		log.Fatalf("creating bucket %q: %v", bucket, err)
	}
	return true
}

// testConfig returns the Client used to access GCS and the default bucket
// name to use. testConfig skips the current test if credentials are not
// available or when being run in Short mode.
func testConfig(ctx context.Context, t *testing.T) (*Client, string) {
	if testing.Short() {
		t.Skip("Integration tests skipped in short mode")
	}
	client, bucket := config(ctx)
	if client == nil {
		t.Skip("Integration tests skipped. See CONTRIBUTING.md for details")
	}
	return client, bucket
}

// config is like testConfig, but it doesn't need a *testing.T.
func config(ctx context.Context) (*Client, string) {
	ts := testutil.TokenSource(ctx, ScopeFullControl)
	if ts == nil {
		return nil, ""
	}
	p := testutil.ProjID()
	if p == "" {
		log.Fatal("The project ID must be set. See CONTRIBUTING.md for details")
	}
	client, err := NewClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	return client, p + suffix
}

func TestBucketMethods(t *testing.T) {
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	projectID := testutil.ProjID()
	newBucket := bucket + "-new"
	// Test Create and Delete.
	if err := client.Bucket(newBucket).Create(ctx, projectID, nil); err != nil {
		t.Errorf("Bucket(%v).Create(%v, %v) failed: %v", newBucket, projectID, nil, err)
	}
	if err := client.Bucket(newBucket).Delete(ctx); err != nil {
		t.Errorf("Bucket(%v).Delete failed: %v", newBucket, err)
	}

	// Test Create and Delete with attributes.
	attrs := BucketAttrs{
		DefaultObjectACL: []ACLRule{{Entity: "domain-google.com", Role: RoleReader}},
	}
	if err := client.Bucket(newBucket).Create(ctx, projectID, &attrs); err != nil {
		t.Errorf("Bucket(%v).Create(%v, %v) failed: %v", newBucket, projectID, attrs, err)
	}
	if err := client.Bucket(newBucket).Delete(ctx); err != nil {
		t.Errorf("Bucket(%v).Delete failed: %v", newBucket, err)
	}
}

func TestIntegration_ConditionalDelete(t *testing.T) {
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	o := client.Bucket(bucket).Object("conddel")

	wc := o.NewWriter(ctx)
	wc.ContentType = "text/plain"
	if _, err := wc.Write([]byte("foo")); err != nil {
		t.Fatal(err)
	}
	if err := wc.Close(); err != nil {
		t.Fatal(err)
	}

	gen := wc.Attrs().Generation
	metaGen := wc.Attrs().MetaGeneration

	if err := o.Generation(gen - 1).Delete(ctx); err == nil {
		t.Fatalf("Unexpected successful delete with Generation")
	}
	if err := o.If(Conditions{MetagenerationMatch: metaGen + 1}).Delete(ctx); err == nil {
		t.Fatalf("Unexpected successful delete with IfMetaGenerationMatch")
	}
	if err := o.If(Conditions{MetagenerationNotMatch: metaGen}).Delete(ctx); err == nil {
		t.Fatalf("Unexpected successful delete with IfMetaGenerationNotMatch")
	}
	if err := o.Generation(gen).Delete(ctx); err != nil {
		t.Fatalf("final delete failed: %v", err)
	}
}

func TestObjects(t *testing.T) {
	// TODO(djd): there are a lot of closely-related tests here which share
	// a common setup. Once we can depend on Go 1.7 features, we should refactor
	// this test to use the sub-test feature. This will increase the readability
	// of this test, and should also reduce the time it takes to execute.
	// https://golang.org/pkg/testing/#hdr-Subtests_and_Sub_benchmarks
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	bkt := client.Bucket(bucket)

	const defaultType = "text/plain"

	// Populate object names and make a map for their contents.
	objects := []string{
		"obj1",
		"obj2",
		"obj/with/slashes",
	}
	contents := make(map[string][]byte)

	// Test Writer.
	for _, obj := range objects {
		c := randomContents()
		if err := writeObject(ctx, bkt.Object(obj), defaultType, c); err != nil {
			t.Errorf("Write for %v failed with %v", obj, err)
		}
		contents[obj] = c
	}

	testObjectIterator(t, bkt, objects)

	// Test Reader.
	for _, obj := range objects {
		rc, err := bkt.Object(obj).NewReader(ctx)
		if err != nil {
			t.Errorf("Can't create a reader for %v, errored with %v", obj, err)
			continue
		}
		slurp, err := ioutil.ReadAll(rc)
		if err != nil {
			t.Errorf("Can't ReadAll object %v, errored with %v", obj, err)
		}
		if got, want := slurp, contents[obj]; !bytes.Equal(got, want) {
			t.Errorf("Contents (%q) = %q; want %q", obj, got, want)
		}
		if got, want := rc.Size(), len(contents[obj]); got != int64(want) {
			t.Errorf("Size (%q) = %d; want %d", obj, got, want)
		}
		if got, want := rc.ContentType(), "text/plain"; got != want {
			t.Errorf("ContentType (%q) = %q; want %q", obj, got, want)
		}
		rc.Close()

		// Test SignedURL
		opts := &SignedURLOptions{
			GoogleAccessID: "xxx@clientid",
			PrivateKey:     dummyKey("rsa"),
			Method:         "GET",
			MD5:            []byte("202cb962ac59075b964b07152d234b70"),
			Expires:        time.Date(2020, time.October, 2, 10, 0, 0, 0, time.UTC),
			ContentType:    "application/json",
			Headers:        []string{"x-header1", "x-header2"},
		}
		u, err := SignedURL(bucket, obj, opts)
		if err != nil {
			t.Fatalf("SignedURL(%q, %q) errored with %v", bucket, obj, err)
		}
		res, err := client.hc.Get(u)
		if err != nil {
			t.Fatalf("Can't get URL %q: %v", u, err)
		}
		slurp, err = ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("Can't ReadAll signed object %v, errored with %v", obj, err)
		}
		if got, want := slurp, contents[obj]; !bytes.Equal(got, want) {
			t.Errorf("Contents (%v) = %q; want %q", obj, got, want)
		}
		res.Body.Close()
	}

	obj := objects[0]
	objlen := int64(len(contents[obj]))
	// Test Range Reader.
	for i, r := range []struct {
		offset, length, want int64
	}{
		{0, objlen, objlen},
		{0, objlen / 2, objlen / 2},
		{objlen / 2, objlen, objlen / 2},
		{0, 0, 0},
		{objlen / 2, 0, 0},
		{objlen / 2, -1, objlen / 2},
		{0, objlen * 2, objlen},
	} {
		rc, err := bkt.Object(obj).NewRangeReader(ctx, r.offset, r.length)
		if err != nil {
			t.Errorf("%d: Can't create a range reader for %v, errored with %v", i, obj, err)
			continue
		}
		if rc.Size() != objlen {
			t.Errorf("%d: Reader has a content-size of %d, want %d", i, rc.Size(), objlen)
		}
		if rc.Remain() != r.want {
			t.Errorf("%d: Reader's available bytes reported as %d, want %d", i, rc.Remain(), r.want)
		}
		slurp, err := ioutil.ReadAll(rc)
		if err != nil {
			t.Errorf("%d:Can't ReadAll object %v, errored with %v", i, obj, err)
			continue
		}
		if len(slurp) != int(r.want) {
			t.Errorf("%d:RangeReader (%d, %d): Read %d bytes, wanted %d bytes", i, r.offset, r.length, len(slurp), r.want)
			continue
		}
		if got, want := slurp, contents[obj][r.offset:r.offset+r.want]; !bytes.Equal(got, want) {
			t.Errorf("RangeReader (%d, %d) = %q; want %q", r.offset, r.length, got, want)
		}
		rc.Close()
	}

	// Test content encoding
	const zeroCount = 20 << 20
	w := bkt.Object("gzip-test").NewWriter(ctx)
	w.ContentEncoding = "gzip"
	gw := gzip.NewWriter(w)
	if _, err := io.Copy(gw, io.LimitReader(zeros{}, zeroCount)); err != nil {
		t.Fatalf("io.Copy, upload: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Errorf("gzip.Close(): %v", err)
	}
	if err := w.Close(); err != nil {
		t.Errorf("w.Close(): %v", err)
	}
	r, err := bkt.Object("gzip-test").NewReader(ctx)
	if err != nil {
		t.Fatalf("NewReader(gzip-test): %v", err)
	}
	n, err := io.Copy(ioutil.Discard, r)
	if err != nil {
		t.Errorf("io.Copy, download: %v", err)
	}
	if n != zeroCount {
		t.Errorf("downloaded bad data: got %d bytes, want %d", n, zeroCount)
	}

	// Test NotFound.
	_, err = bkt.Object("obj-not-exists").NewReader(ctx)
	if err != ErrObjectNotExist {
		t.Errorf("Object should not exist, err found to be %v", err)
	}

	objName := objects[0]

	// Test NewReader googleapi.Error.
	// Since a 429 or 5xx is hard to cause, we trigger a 416.
	realLen := len(contents[objName])
	_, err = bkt.Object(objName).NewRangeReader(ctx, int64(realLen*2), 10)
	if err, ok := err.(*googleapi.Error); !ok {
		t.Error("NewRangeReader did not return a googleapi.Error")
	} else {
		if err.Code != 416 {
			t.Errorf("Code = %d; want %d", err.Code, 416)
		}
		if len(err.Header) == 0 {
			t.Error("Missing googleapi.Error.Header")
		}
		if len(err.Body) == 0 {
			t.Error("Missing googleapi.Error.Body")
		}
	}

	// Test StatObject.
	o, err := bkt.Object(objName).Attrs(ctx)
	if err != nil {
		t.Error(err)
	}
	if got, want := o.Name, objName; got != want {
		t.Errorf("Name (%v) = %q; want %q", objName, got, want)
	}
	if got, want := o.ContentType, defaultType; got != want {
		t.Errorf("ContentType (%v) = %q; want %q", objName, got, want)
	}
	created := o.Created
	// Check that the object is newer than its containing bucket.
	bAttrs, err := bkt.Attrs(ctx)
	if err != nil {
		t.Error(err)
	}
	if o.Created.Before(bAttrs.Created) {
		t.Errorf("Object %v is older than its containing bucket, %v", o, bAttrs)
	}

	// Test object copy.
	copyName := "copy-" + objName
	copyObj, err := bkt.Object(copyName).CopierFrom(bkt.Object(objName)).Run(ctx)
	if err != nil {
		t.Errorf("Copier.Run failed with %v", err)
	} else if !namesEqual(copyObj, bucket, copyName) {
		t.Errorf("Copy object bucket, name: got %q.%q, want %q.%q",
			copyObj.Bucket, copyObj.Name, bucket, copyName)
	}

	// Copying with attributes.
	const contentEncoding = "identity"
	copier := bkt.Object(copyName).CopierFrom(bkt.Object(objName))
	copier.ContentEncoding = contentEncoding
	copyObj, err = copier.Run(ctx)
	if err != nil {
		t.Errorf("Copier.Run failed with %v", err)
	} else {
		if !namesEqual(copyObj, bucket, copyName) {
			t.Errorf("Copy object bucket, name: got %q.%q, want %q.%q",
				copyObj.Bucket, copyObj.Name, bucket, copyName)
		}
		if copyObj.ContentEncoding != contentEncoding {
			t.Errorf("Copy ContentEncoding: got %q, want %q", copyObj.ContentEncoding, contentEncoding)
		}
	}

	// Test UpdateAttrs.
	metadata := map[string]string{"key": "value"}
	updated, err := bkt.Object(objName).Update(ctx, ObjectAttrsToUpdate{
		ContentType:     "text/html",
		ContentLanguage: "en",
		Metadata:        metadata,
		ACL:             []ACLRule{{Entity: "domain-google.com", Role: RoleReader}},
	})
	if err != nil {
		t.Errorf("UpdateAttrs failed with %v", err)
	} else {
		if got, want := updated.ContentType, "text/html"; got != want {
			t.Errorf("updated.ContentType == %q; want %q", got, want)
		}
		if got, want := updated.ContentLanguage, "en"; got != want {
			t.Errorf("updated.ContentLanguage == %q; want %q", updated.ContentLanguage, want)
		}
		if got, want := updated.Metadata, metadata; !reflect.DeepEqual(got, want) {
			t.Errorf("updated.Metadata == %+v; want %+v", updated.Metadata, want)
		}
		if got, want := updated.Created, created; got != want {
			t.Errorf("updated.Created == %q; want %q", got, want)
		}
		if !updated.Created.Before(updated.Updated) {
			t.Errorf("updated.Updated should be newer than update.Created")
		}
	}
	// Delete ContentType and ContentLanguage.
	updated, err = bkt.Object(objName).Update(ctx, ObjectAttrsToUpdate{
		ContentType:     "",
		ContentLanguage: "",
		Metadata:        map[string]string{},
	})
	if err != nil {
		t.Errorf("UpdateAttrs failed with %v", err)
	} else {
		if got, want := updated.ContentType, ""; got != want {
			t.Errorf("updated.ContentType == %q; want %q", got, want)
		}
		if got, want := updated.ContentLanguage, ""; got != want {
			t.Errorf("updated.ContentLanguage == %q; want %q", updated.ContentLanguage, want)
		}
		if updated.Metadata != nil {
			t.Errorf("updated.Metadata == %+v; want nil", updated.Metadata)
		}
		if got, want := updated.Created, created; got != want {
			t.Errorf("updated.Created == %q; want %q", got, want)
		}
		if !updated.Created.Before(updated.Updated) {
			t.Errorf("updated.Updated should be newer than update.Created")
		}
	}

	// Test checksums.
	checksumCases := []struct {
		name     string
		contents [][]byte
		size     int64
		md5      string
		crc32c   uint32
	}{
		{
			name:     "checksum-object",
			contents: [][]byte{[]byte("hello"), []byte("world")},
			size:     10,
			md5:      "fc5e038d38a57032085441e7fe7010b0",
			crc32c:   1456190592,
		},
		{
			name:     "zero-object",
			contents: [][]byte{},
			size:     0,
			md5:      "d41d8cd98f00b204e9800998ecf8427e",
			crc32c:   0,
		},
	}
	for _, c := range checksumCases {
		wc := bkt.Object(c.name).NewWriter(ctx)
		for _, data := range c.contents {
			if _, err := wc.Write(data); err != nil {
				t.Errorf("Write(%q) failed with %q", data, err)
			}
		}
		if err = wc.Close(); err != nil {
			t.Errorf("%q: close failed with %q", c.name, err)
		}
		obj := wc.Attrs()
		if got, want := obj.Size, c.size; got != want {
			t.Errorf("Object (%q) Size = %v; want %v", c.name, got, want)
		}
		if got, want := fmt.Sprintf("%x", obj.MD5), c.md5; got != want {
			t.Errorf("Object (%q) MD5 = %q; want %q", c.name, got, want)
		}
		if got, want := obj.CRC32C, c.crc32c; got != want {
			t.Errorf("Object (%q) CRC32C = %v; want %v", c.name, got, want)
		}
	}

	// Test public ACL.
	publicObj := objects[0]
	if err = bkt.Object(publicObj).ACL().Set(ctx, AllUsers, RoleReader); err != nil {
		t.Errorf("PutACLEntry failed with %v", err)
	}
	publicClient, err := NewClient(ctx, option.WithHTTPClient(http.DefaultClient))
	if err != nil {
		t.Fatal(err)
	}

	slurp, err := readObject(ctx, publicClient.Bucket(bucket).Object(publicObj))
	if err != nil {
		t.Errorf("readObject failed with %v", err)
	} else if !bytes.Equal(slurp, contents[publicObj]) {
		t.Errorf("Public object's content: got %q, want %q", slurp, contents[publicObj])
	}

	// Test writer error handling.
	wc := publicClient.Bucket(bucket).Object(publicObj).NewWriter(ctx)
	if _, err := wc.Write([]byte("hello")); err != nil {
		t.Errorf("Write unexpectedly failed with %v", err)
	}
	if err = wc.Close(); err == nil {
		t.Error("Close expected an error, found none")
	}

	// Test deleting the copy object.
	if err := bkt.Object(copyName).Delete(ctx); err != nil {
		t.Errorf("Deletion of %v failed with %v", copyName, err)
	}
	// Deleting it a second time should return ErrObjectNotExist.
	if err := bkt.Object(copyName).Delete(ctx); err != ErrObjectNotExist {
		t.Errorf("second deletion of %v = %v; want ErrObjectNotExist", copyName, err)
	}
	_, err = bkt.Object(copyName).Attrs(ctx)
	if err != ErrObjectNotExist {
		t.Errorf("Copy is expected to be deleted, stat errored with %v", err)
	}

	// Test object composition.
	var compSrcs []*ObjectHandle
	var wantContents []byte
	for _, obj := range objects {
		compSrcs = append(compSrcs, bkt.Object(obj))
		wantContents = append(wantContents, contents[obj]...)
	}
	checkCompose := func(obj *ObjectHandle, wantContentType string) {
		rc, err := obj.NewReader(ctx)
		if err != nil {
			t.Fatalf("NewReader: %v", err)
		}
		slurp, err = ioutil.ReadAll(rc)
		if err != nil {
			t.Fatalf("ioutil.ReadAll: %v", err)
		}
		defer rc.Close()
		if !bytes.Equal(slurp, wantContents) {
			t.Errorf("Composed object contents\ngot:  %q\nwant: %q", slurp, wantContents)
		}
		if got := rc.ContentType(); got != wantContentType {
			t.Errorf("Composed object content-type = %q, want %q", got, wantContentType)
		}
	}

	// Compose should work even if the user sets no destination attributes.
	compDst := bkt.Object("composed1")
	c := compDst.ComposerFrom(compSrcs...)
	if _, err := c.Run(ctx); err != nil {
		t.Fatalf("ComposeFrom error: %v", err)
	}
	checkCompose(compDst, "application/octet-stream")

	// It should also work if we do.
	compDst = bkt.Object("composed2")
	c = compDst.ComposerFrom(compSrcs...)
	c.ContentType = "text/json"
	if _, err := c.Run(ctx); err != nil {
		t.Fatalf("ComposeFrom error: %v", err)
	}
	checkCompose(compDst, "text/json")
}

func namesEqual(obj *ObjectAttrs, bucketName, objectName string) bool {
	return obj.Bucket == bucketName && obj.Name == objectName
}

func testObjectIterator(t *testing.T, bkt *BucketHandle, objects []string) {
	ctx := context.Background()
	// Collect the list of items we expect: ObjectAttrs in lexical order by name.
	names := make([]string, len(objects))
	copy(names, objects)
	sort.Strings(names)
	var attrs []*ObjectAttrs
	for _, name := range names {
		attr, err := bkt.Object(name).Attrs(ctx)
		if err != nil {
			t.Errorf("Object(%q).Attrs: %v", name, err)
			return
		}
		attrs = append(attrs, attr)
	}
	// The following iterator test fails occasionally, probably because the
	// underlying Objects.List operation is eventually consistent. So we retry
	// it.
	tctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	var msg string
	var ok bool
	err := internal.Retry(tctx, gax.Backoff{}, func() (stop bool, err error) {
		msg, ok = itesting.TestIterator(attrs,
			func() interface{} { return bkt.Objects(ctx, &Query{Prefix: "obj"}) },
			func(it interface{}) (interface{}, error) { return it.(*ObjectIterator).Next() })
		if ok {
			return true, nil
		} else {
			t.Logf("TestIterator failed, trying again: %s", msg)
			return false, nil
		}
	})
	if !ok {
		t.Errorf("ObjectIterator.Next: %s (err=%v)", msg, err)
	}
	// TODO(jba): test query.Delimiter != ""
}

func TestACL(t *testing.T) {
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	bkt := client.Bucket(bucket)

	entity := ACLEntity("domain-google.com")
	rule := ACLRule{Entity: entity, Role: RoleReader}
	if err := bkt.DefaultObjectACL().Set(ctx, entity, RoleReader); err != nil {
		t.Errorf("Can't put default ACL rule for the bucket, errored with %v", err)
	}
	acl, err := bkt.DefaultObjectACL().List(ctx)
	if err != nil {
		t.Errorf("DefaultObjectACL.List for bucket %q: %v", bucket, err)
	} else if !hasRule(acl, rule) {
		t.Errorf("default ACL missing %#v", rule)
	}
	aclObjects := []string{"acl1", "acl2"}
	for _, obj := range aclObjects {
		c := randomContents()
		if err := writeObject(ctx, bkt.Object(obj), "", c); err != nil {
			t.Errorf("Write for %v failed with %v", obj, err)
		}
	}
	name := aclObjects[0]
	o := bkt.Object(name)
	acl, err = o.ACL().List(ctx)
	if err != nil {
		t.Errorf("Can't retrieve ACL of %v", name)
	} else if !hasRule(acl, rule) {
		t.Errorf("object ACL missing %+v", rule)
	}
	if err := o.ACL().Delete(ctx, entity); err != nil {
		t.Errorf("object ACL: could not delete entity %s", entity)
	}
	// Delete the default ACL rule. We can't move this code earlier in the
	// test, because the test depends on the fact that the object ACL inherits
	// it.
	if err := bkt.DefaultObjectACL().Delete(ctx, entity); err != nil {
		t.Errorf("default ACL: could not delete entity %s", entity)
	}

	entity2 := ACLEntity("user-jbd@google.com")
	rule2 := ACLRule{Entity: entity2, Role: RoleReader}
	if err := bkt.ACL().Set(ctx, entity2, RoleReader); err != nil {
		t.Errorf("Error while putting bucket ACL rule: %v", err)
	}
	bACL, err := bkt.ACL().List(ctx)
	if err != nil {
		t.Errorf("Error while getting the ACL of the bucket: %v", err)
	} else if !hasRule(bACL, rule2) {
		t.Errorf("bucket ACL missing %+v", rule2)
	}
	if err := bkt.ACL().Delete(ctx, entity2); err != nil {
		t.Errorf("Error while deleting bucket ACL rule: %v", err)
	}

}

func hasRule(acl []ACLRule, rule ACLRule) bool {
	for _, r := range acl {
		if r == rule {
			return true
		}
	}
	return false
}

func TestValidObjectNames(t *testing.T) {
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	bkt := client.Bucket(bucket)

	validNames := []string{
		"gopher",
		"Гоферови",
		"a",
		strings.Repeat("a", 1024),
	}
	for _, name := range validNames {
		if err := writeObject(ctx, bkt.Object(name), "", []byte("data")); err != nil {
			t.Errorf("Object %q write failed: %v. Want success", name, err)
			continue
		}
		defer bkt.Object(name).Delete(ctx)
	}

	invalidNames := []string{
		"", // Too short.
		strings.Repeat("a", 1025), // Too long.
		"new\nlines",
		"bad\xffunicode",
	}
	for _, name := range invalidNames {
		// Invalid object names will either cause failure during Write or Close.
		if err := writeObject(ctx, bkt.Object(name), "", []byte("data")); err != nil {
			continue
		}
		defer bkt.Object(name).Delete(ctx)
		t.Errorf("%q should have failed. Didn't", name)
	}
}

func TestWriterContentType(t *testing.T) {
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	obj := client.Bucket(bucket).Object("content")
	testCases := []struct {
		content           string
		setType, wantType string
	}{
		{
			content:  "It was the best of times, it was the worst of times.",
			wantType: "text/plain; charset=utf-8",
		},
		{
			content:  "<html><head><title>My first page</title></head></html>",
			wantType: "text/html; charset=utf-8",
		},
		{
			content:  "<html><head><title>My first page</title></head></html>",
			setType:  "text/html",
			wantType: "text/html",
		},
		{
			content:  "<html><head><title>My first page</title></head></html>",
			setType:  "image/jpeg",
			wantType: "image/jpeg",
		},
	}
	for i, tt := range testCases {
		if err := writeObject(ctx, obj, tt.setType, []byte(tt.content)); err != nil {
			t.Errorf("writing #%d: %v", i, err)
		}
		attrs, err := obj.Attrs(ctx)
		if err != nil {
			t.Errorf("obj.Attrs: %v", err)
			continue
		}
		if got := attrs.ContentType; got != tt.wantType {
			t.Errorf("Content-Type = %q; want %q\nContent: %q\nSet Content-Type: %q", got, tt.wantType, tt.content, tt.setType)
		}
	}
}

func TestZeroSizedObject(t *testing.T) {
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	obj := client.Bucket(bucket).Object("zero")

	// Check writing it works as expected.
	w := obj.NewWriter(ctx)
	if err := w.Close(); err != nil {
		t.Fatalf("Writer.Close: %v", err)
	}
	defer obj.Delete(ctx)

	// Check we can read it too.
	body, err := readObject(ctx, obj)
	if err != nil {
		t.Fatalf("readObject: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("Body is %v, want empty []byte{}", body)
	}
}

func TestIntegration_Encryption(t *testing.T) {
	// This function tests customer-supplied encryption keys for all operations
	// involving objects. Bucket and ACL operations aren't tested because they
	// aren't affected customer encryption.
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	obj := client.Bucket(bucket).Object("customer-encryption")
	key := []byte("my-secret-AES-256-encryption-key")
	keyHash := sha256.Sum256(key)
	keyHashB64 := base64.StdEncoding.EncodeToString(keyHash[:])
	key2 := []byte("My-Secret-AES-256-Encryption-Key")
	contents := "top secret."

	checkMetadataCall := func(msg string, f func(o *ObjectHandle) (*ObjectAttrs, error)) {
		// Performing a metadata operation without the key should succeed.
		attrs, err := f(obj)
		if err != nil {
			t.Fatalf("%s: %v", msg, err)
		}
		// The key hash should match...
		if got, want := attrs.CustomerKeySHA256, keyHashB64; got != want {
			t.Errorf("%s: key hash: got %q, want %q", msg, got, want)
		}
		// ...but CRC and MD5 should not be present.
		if attrs.CRC32C != 0 {
			t.Errorf("%s: CRC: got %v, want 0", msg, attrs.CRC32C)
		}
		if len(attrs.MD5) > 0 {
			t.Errorf("%s: MD5: got %v, want len == 0", msg, attrs.MD5)
		}

		// Performing a metadata operation with the key should succeed.
		attrs, err = f(obj.Key(key))
		if err != nil {
			t.Fatalf("%s: %v", msg, err)
		}
		// Check the key and content hashes.
		if got, want := attrs.CustomerKeySHA256, keyHashB64; got != want {
			t.Errorf("%s: key hash: got %q, want %q", msg, got, want)
		}
		if attrs.CRC32C == 0 {
			t.Errorf("%s: CRC: got 0, want non-zero", msg)
		}
		if len(attrs.MD5) == 0 {
			t.Errorf("%s: MD5: got len == 0, want len > 0", msg)
		}
	}

	checkRead := func(msg string, o *ObjectHandle, k []byte, wantContents string) {
		// Reading the object without the key should fail.
		if _, err := readObject(ctx, o); err == nil {
			t.Errorf("%s: reading without key: want error, got nil", msg)
		}
		// Reading the object with the key should succeed.
		got, err := readObject(ctx, o.Key(k))
		if err != nil {
			t.Fatalf("%s: %v", msg, err)
		}
		gotContents := string(got)
		// And the contents should match what we wrote.
		if gotContents != wantContents {
			t.Errorf("%s: contents: got %q, want %q", msg, gotContents, wantContents)
		}
	}

	checkReadUnencrypted := func(msg string, obj *ObjectHandle, wantContents string) {
		got, err := readObject(ctx, obj)
		if err != nil {
			t.Fatalf("%s: %v", msg, err)
		}
		gotContents := string(got)
		if gotContents != wantContents {
			t.Errorf("%s: got %q, want %q", gotContents, wantContents)
		}
	}

	// Write to obj using our own encryption key, which is a valid 32-byte
	// AES-256 key.
	w := obj.Key(key).NewWriter(ctx)
	w.Write([]byte(contents))
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	checkMetadataCall("Attrs", func(o *ObjectHandle) (*ObjectAttrs, error) {
		return o.Attrs(ctx)
	})

	checkMetadataCall("Update", func(o *ObjectHandle) (*ObjectAttrs, error) {
		return o.Update(ctx, ObjectAttrsToUpdate{ContentLanguage: "en"})
	})

	checkRead("first object", obj, key, contents)

	obj2 := client.Bucket(bucket).Object("customer-encryption-2")
	// Copying an object without the key should fail.
	if _, err := obj2.CopierFrom(obj).Run(ctx); err == nil {
		t.Fatal("want error, got nil")
	}
	// Copying an object with the key should succeed.
	if _, err := obj2.CopierFrom(obj.Key(key)).Run(ctx); err != nil {
		t.Fatal(err)
	}
	// The destination object is not encrypted; we can read it without a key.
	checkReadUnencrypted("copy dest", obj2, contents)

	// Providing a key on the destination but not the source should fail,
	// since the source is encrypted.
	if _, err := obj2.Key(key2).CopierFrom(obj).Run(ctx); err == nil {
		t.Fatal("want error, got nil")
	}

	// But copying with keys for both source and destination should succeed.
	if _, err := obj2.Key(key2).CopierFrom(obj.Key(key)).Run(ctx); err != nil {
		t.Fatal(err)
	}
	// And the destination should be encrypted, meaning we can only read it
	// with a key.
	checkRead("copy destination", obj2, key2, contents)

	// Change obj2's key to prepare for compose, where all objects must have
	// the same key. Also illustrates key rotation: copy an object to itself
	// with a different key.
	if _, err := obj2.Key(key).CopierFrom(obj2.Key(key2)).Run(ctx); err != nil {
		t.Fatal(err)
	}
	obj3 := client.Bucket(bucket).Object("customer-encryption-3")
	// Composing without keys should fail.
	if _, err := obj3.ComposerFrom(obj, obj2).Run(ctx); err == nil {
		t.Fatal("want error, got nil")
	}
	// Keys on the source objects result in an error.
	if _, err := obj3.ComposerFrom(obj.Key(key), obj2).Run(ctx); err == nil {
		t.Fatal("want error, got nil")
	}
	// A key on the destination object both decrypts the source objects
	// and encrypts the destination.
	if _, err := obj3.Key(key).ComposerFrom(obj, obj2).Run(ctx); err != nil {
		t.Fatalf("got %v, want nil", err)
	}
	// Check that the destination in encrypted.
	checkRead("compose destination", obj3, key, contents+contents)

	// You can't compose one or more unencrypted source objects into an
	// encrypted destination object.
	_, err := obj2.CopierFrom(obj2.Key(key)).Run(ctx) // unencrypt obj2
	if err != nil {
		t.Fatal(err)
	}
	if _, err := obj3.Key(key).ComposerFrom(obj2).Run(ctx); err == nil {
		t.Fatal("got nil, want error")
	}
}

func TestIntegration_NonexistentBucket(t *testing.T) {
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	bkt := client.Bucket(bucket + "-nonexistent")
	if _, err := bkt.Attrs(ctx); err != ErrBucketNotExist {
		t.Errorf("Attrs: got %v, want ErrBucketNotExist", err)
	}
	it := bkt.Objects(ctx, nil)
	if _, err := it.Next(); err != ErrBucketNotExist {
		t.Errorf("Objects: got %v, want ErrBucketNotExist", err)
	}
}

func TestIntegration_PerObjectStorageClass(t *testing.T) {
	const (
		defaultStorageClass = "STANDARD"
		newStorageClass     = "MULTI_REGIONAL"
	)
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	bkt := client.Bucket(bucket)

	// The bucket should have the default storage class.
	battrs, err := bkt.Attrs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if battrs.StorageClass != defaultStorageClass {
		t.Fatalf("bucket storage class: got %q, want %q",
			battrs.StorageClass, defaultStorageClass)
	}
	// Write an object; it should start with the bucket's storage class.
	obj := bkt.Object("posc")
	if err := writeObject(ctx, obj, "", []byte("foo")); err != nil {
		t.Fatal(err)
	}
	oattrs, err := obj.Attrs(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if oattrs.StorageClass != defaultStorageClass {
		t.Fatalf("object storage class: got %q, want %q",
			oattrs.StorageClass, defaultStorageClass)
	}
	// Now use Copy to change the storage class.
	copier := obj.CopierFrom(obj)
	copier.StorageClass = newStorageClass
	oattrs2, err := copier.Run(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if oattrs2.StorageClass != newStorageClass {
		t.Fatalf("new object storage class: got %q, want %q",
			oattrs2.StorageClass, newStorageClass)
	}

	// We can also write a new object using a non-default storage class.
	obj2 := bkt.Object("posc2")
	w := obj2.NewWriter(ctx)
	w.StorageClass = newStorageClass
	if _, err := w.Write([]byte("xxx")); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if w.Attrs().StorageClass != newStorageClass {
		t.Fatalf("new object storage class: got %q, want %q",
			w.Attrs().StorageClass, newStorageClass)
	}
}

func TestIntegration_BucketInCopyAttrs(t *testing.T) {
	// Confirm that if bucket is included in the object attributes of a rewrite
	// call, but object name and content-type aren't, then we get an error. See
	// the comment in Copier.Run.
	ctx := context.Background()
	client, bucket := testConfig(ctx, t)
	defer client.Close()

	bkt := client.Bucket(bucket)
	obj := bkt.Object("bucketInCopyAttrs")
	if err := writeObject(ctx, obj, "", []byte("foo")); err != nil {
		t.Fatal(err)
	}
	copier := obj.CopierFrom(obj)
	rawObject := copier.ObjectAttrs.toRawObject(bucket)
	_, err := copier.callRewrite(ctx, obj, rawObject)
	if err == nil {
		t.Errorf("got nil, want error")
	}
}

func writeObject(ctx context.Context, obj *ObjectHandle, contentType string, contents []byte) error {
	w := obj.NewWriter(ctx)
	w.ContentType = contentType
	if contents != nil {
		if _, err := w.Write(contents); err != nil {
			_ = w.Close()
			return err
		}
	}
	return w.Close()
}

func readObject(ctx context.Context, obj *ObjectHandle) ([]byte, error) {
	r, err := obj.NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return ioutil.ReadAll(r)
}

// cleanup deletes the bucket used for testing, as well as old
// testing buckets that weren't cleaned previously.
func cleanup() error {
	if testing.Short() {
		return nil // Don't clean up in short mode.
	}
	ctx := context.Background()
	client, bucket := config(ctx)
	if client == nil {
		return nil // Don't cleanup if we're not configured correctly.
	}
	defer client.Close()
	if err := killBucket(ctx, client, bucket); err != nil {
		return err
	}

	// Delete buckets whose name begins with our test prefix, and which were
	// created a while ago. (Unfortunately GCS doesn't provide last-modified
	// time, which would be a better way to check for staleness.)
	const expireAge = 24 * time.Hour
	projectID := testutil.ProjID()
	it := client.Buckets(ctx, projectID)
	it.Prefix = projectID + testPrefix
	for {
		bktAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if time.Since(bktAttrs.Created) > expireAge {
			log.Printf("deleting bucket %q, which is more than %s old", bktAttrs.Name, expireAge)
			if err := killBucket(ctx, client, bktAttrs.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

// killBucket deletes a bucket and all its objects.
func killBucket(ctx context.Context, client *Client, bucketName string) error {
	bkt := client.Bucket(bucketName)
	// Bucket must be empty to delete.
	it := bkt.Objects(ctx, nil)
	for {
		objAttrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if err := bkt.Object(objAttrs.Name).Delete(ctx); err != nil {
			return fmt.Errorf("deleting %q: %v", bucketName+"/"+objAttrs.Name, err)
		}
	}
	// GCS is eventually consistent, so this delete may fail because the
	// replica still sees an object in the bucket. We log the error and expect
	// a later test run to delete the bucket.
	if err := bkt.Delete(ctx); err != nil {
		log.Printf("deleting %q: %v", bucketName, err)
	}
	return nil
}

func randomContents() []byte {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("hello world%d", rand.Intn(100000)))
	return h.Sum(nil)
}

type zeros struct{}

func (zeros) Read(p []byte) (int, error) { return len(p), nil }
