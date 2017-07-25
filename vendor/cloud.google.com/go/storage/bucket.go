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
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/iterator"
	raw "google.golang.org/api/storage/v1"
)

// Create creates the Bucket in the project.
// If attrs is nil the API defaults will be used.
func (b *BucketHandle) Create(ctx context.Context, projectID string, attrs *BucketAttrs) error {
	var bkt *raw.Bucket
	if attrs != nil {
		bkt = attrs.toRawBucket()
	} else {
		bkt = &raw.Bucket{}
	}
	bkt.Name = b.name
	req := b.c.raw.Buckets.Insert(projectID, bkt)
	return runWithRetry(ctx, func() error { _, err := req.Context(ctx).Do(); return err })
}

// Delete deletes the Bucket.
func (b *BucketHandle) Delete(ctx context.Context) error {
	req := b.c.raw.Buckets.Delete(b.name)
	return runWithRetry(ctx, func() error { return req.Context(ctx).Do() })
}

// ACL returns an ACLHandle, which provides access to the bucket's access control list.
// This controls who can list, create or overwrite the objects in a bucket.
// This call does not perform any network operations.
func (b *BucketHandle) ACL() *ACLHandle {
	return &b.acl
}

// DefaultObjectACL returns an ACLHandle, which provides access to the bucket's default object ACLs.
// These ACLs are applied to newly created objects in this bucket that do not have a defined ACL.
// This call does not perform any network operations.
func (b *BucketHandle) DefaultObjectACL() *ACLHandle {
	return &b.defaultObjectACL
}

// Object returns an ObjectHandle, which provides operations on the named object.
// This call does not perform any network operations.
//
// name must consist entirely of valid UTF-8-encoded runes. The full specification
// for valid object names can be found at:
//   https://cloud.google.com/storage/docs/bucket-naming
func (b *BucketHandle) Object(name string) *ObjectHandle {
	return &ObjectHandle{
		c:      b.c,
		bucket: b.name,
		object: name,
		acl: ACLHandle{
			c:      b.c,
			bucket: b.name,
			object: name,
		},
		gen: -1,
	}
}

// Attrs returns the metadata for the bucket.
func (b *BucketHandle) Attrs(ctx context.Context) (*BucketAttrs, error) {
	var resp *raw.Bucket
	var err error
	err = runWithRetry(ctx, func() error {
		resp, err = b.c.raw.Buckets.Get(b.name).Projection("full").Context(ctx).Do()
		return err
	})
	if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusNotFound {
		return nil, ErrBucketNotExist
	}
	if err != nil {
		return nil, err
	}
	return newBucket(resp), nil
}

// BucketAttrs represents the metadata for a Google Cloud Storage bucket.
type BucketAttrs struct {
	// Name is the name of the bucket.
	Name string

	// ACL is the list of access control rules on the bucket.
	ACL []ACLRule

	// DefaultObjectACL is the list of access controls to
	// apply to new objects when no object ACL is provided.
	DefaultObjectACL []ACLRule

	// Location is the location of the bucket. It defaults to "US".
	Location string

	// MetaGeneration is the metadata generation of the bucket.
	MetaGeneration int64

	// StorageClass is the default storage class of the bucket. This defines
	// how objects in the bucket are stored and determines the SLA
	// and the cost of storage. Typical values are "MULTI_REGIONAL",
	// "REGIONAL", "NEARLINE", "COLDLINE", "STANDARD" and
	// "DURABLE_REDUCED_AVAILABILITY". Defaults to "STANDARD", which
	// is equivalent to "MULTI_REGIONAL" or "REGIONAL" depending on
	// the bucket's location settings.
	StorageClass string

	// Created is the creation time of the bucket.
	Created time.Time

	// VersioningEnabled reports whether this bucket has versioning enabled.
	// This field is read-only.
	VersioningEnabled bool
}

func newBucket(b *raw.Bucket) *BucketAttrs {
	if b == nil {
		return nil
	}
	bucket := &BucketAttrs{
		Name:              b.Name,
		Location:          b.Location,
		MetaGeneration:    b.Metageneration,
		StorageClass:      b.StorageClass,
		Created:           convertTime(b.TimeCreated),
		VersioningEnabled: b.Versioning != nil && b.Versioning.Enabled,
	}
	acl := make([]ACLRule, len(b.Acl))
	for i, rule := range b.Acl {
		acl[i] = ACLRule{
			Entity: ACLEntity(rule.Entity),
			Role:   ACLRole(rule.Role),
		}
	}
	bucket.ACL = acl
	objACL := make([]ACLRule, len(b.DefaultObjectAcl))
	for i, rule := range b.DefaultObjectAcl {
		objACL[i] = ACLRule{
			Entity: ACLEntity(rule.Entity),
			Role:   ACLRole(rule.Role),
		}
	}
	bucket.DefaultObjectACL = objACL
	return bucket
}

// toRawBucket copies the editable attribute from b to the raw library's Bucket type.
func (b *BucketAttrs) toRawBucket() *raw.Bucket {
	var acl []*raw.BucketAccessControl
	if len(b.ACL) > 0 {
		acl = make([]*raw.BucketAccessControl, len(b.ACL))
		for i, rule := range b.ACL {
			acl[i] = &raw.BucketAccessControl{
				Entity: string(rule.Entity),
				Role:   string(rule.Role),
			}
		}
	}
	dACL := toRawObjectACL(b.DefaultObjectACL)
	return &raw.Bucket{
		Name:             b.Name,
		DefaultObjectAcl: dACL,
		Location:         b.Location,
		StorageClass:     b.StorageClass,
		Acl:              acl,
	}
}

// Objects returns an iterator over the objects in the bucket that match the Query q.
// If q is nil, no filtering is done.
func (b *BucketHandle) Objects(ctx context.Context, q *Query) *ObjectIterator {
	it := &ObjectIterator{
		ctx:    ctx,
		bucket: b,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.items) },
		func() interface{} { b := it.items; it.items = nil; return b })
	if q != nil {
		it.query = *q
	}
	return it
}

// An ObjectIterator is an iterator over ObjectAttrs.
type ObjectIterator struct {
	ctx      context.Context
	bucket   *BucketHandle
	query    Query
	pageInfo *iterator.PageInfo
	nextFunc func() error
	items    []*ObjectAttrs
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *ObjectIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

// Next returns the next result. Its second return value is iterator.Done if
// there are no more results. Once Next returns iterator.Done, all subsequent
// calls will return iterator.Done.
//
// If Query.Delimiter is non-empty, some of the ObjectAttrs returned by Next will
// have a non-empty Prefix field, and a zero value for all other fields. These
// represent prefixes.
func (it *ObjectIterator) Next() (*ObjectAttrs, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	item := it.items[0]
	it.items = it.items[1:]
	return item, nil
}

func (it *ObjectIterator) fetch(pageSize int, pageToken string) (string, error) {
	req := it.bucket.c.raw.Objects.List(it.bucket.name)
	req.Projection("full")
	req.Delimiter(it.query.Delimiter)
	req.Prefix(it.query.Prefix)
	req.Versions(it.query.Versions)
	req.PageToken(pageToken)
	if pageSize > 0 {
		req.MaxResults(int64(pageSize))
	}
	var resp *raw.Objects
	var err error
	err = runWithRetry(it.ctx, func() error {
		resp, err = req.Context(it.ctx).Do()
		return err
	})
	if err != nil {
		if e, ok := err.(*googleapi.Error); ok && e.Code == http.StatusNotFound {
			err = ErrBucketNotExist
		}
		return "", err
	}
	for _, item := range resp.Items {
		it.items = append(it.items, newObject(item))
	}
	for _, prefix := range resp.Prefixes {
		it.items = append(it.items, &ObjectAttrs{Prefix: prefix})
	}
	return resp.NextPageToken, nil
}

// TODO(jbd): Add storage.buckets.update.

// Buckets returns an iterator over the buckets in the project. You may
// optionally set the iterator's Prefix field to restrict the list to buckets
// whose names begin with the prefix. By default, all buckets in the project
// are returned.
func (c *Client) Buckets(ctx context.Context, projectID string) *BucketIterator {
	it := &BucketIterator{
		ctx:       ctx,
		client:    c,
		projectID: projectID,
	}
	it.pageInfo, it.nextFunc = iterator.NewPageInfo(
		it.fetch,
		func() int { return len(it.buckets) },
		func() interface{} { b := it.buckets; it.buckets = nil; return b })
	return it
}

// A BucketIterator is an iterator over BucketAttrs.
type BucketIterator struct {
	// Prefix restricts the iterator to buckets whose names begin with it.
	Prefix string

	ctx       context.Context
	client    *Client
	projectID string
	buckets   []*BucketAttrs
	pageInfo  *iterator.PageInfo
	nextFunc  func() error
}

// Next returns the next result. Its second return value is iterator.Done if
// there are no more results. Once Next returns iterator.Done, all subsequent
// calls will return iterator.Done.
func (it *BucketIterator) Next() (*BucketAttrs, error) {
	if err := it.nextFunc(); err != nil {
		return nil, err
	}
	b := it.buckets[0]
	it.buckets = it.buckets[1:]
	return b, nil
}

// PageInfo supports pagination. See the google.golang.org/api/iterator package for details.
func (it *BucketIterator) PageInfo() *iterator.PageInfo { return it.pageInfo }

func (it *BucketIterator) fetch(pageSize int, pageToken string) (string, error) {
	req := it.client.raw.Buckets.List(it.projectID)
	req.Projection("full")
	req.Prefix(it.Prefix)
	req.PageToken(pageToken)
	if pageSize > 0 {
		req.MaxResults(int64(pageSize))
	}
	var resp *raw.Buckets
	var err error
	err = runWithRetry(it.ctx, func() error {
		resp, err = req.Context(it.ctx).Do()
		return err
	})
	if err != nil {
		return "", err
	}
	for _, item := range resp.Items {
		it.buckets = append(it.buckets, newBucket(item))
	}
	return resp.NextPageToken, nil
}
