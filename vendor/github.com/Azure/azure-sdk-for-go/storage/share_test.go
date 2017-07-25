package storage

import (
	"math/rand"

	chk "gopkg.in/check.v1"
)

type StorageShareSuite struct{}

var _ = chk.Suite(&StorageShareSuite{})

const testSharePrefix = "zzzzztest"

func randShare() string {
	return testSharePrefix + randString(32-len(testSharePrefix))
}

func getFileClient(c *chk.C) FileServiceClient {
	return getBasicClient(c).GetFileService()
}

func (s *StorageShareSuite) TestCreateShareDeleteShare(c *chk.C) {
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())
	c.Assert(share.Create(), chk.IsNil)
	c.Assert(share.Delete(), chk.IsNil)
}

func (s *StorageShareSuite) TestCreateShareIfNotExists(c *chk.C) {
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	// First create
	ok, err := share.CreateIfNotExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)

	// Second create, should not give errors
	ok, err = share.CreateIfNotExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	// cleanup
	share.Delete()
}

func (s *StorageShareSuite) TestDeleteShareIfNotExists(c *chk.C) {
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	// delete non-existing share
	ok, err := share.DeleteIfExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	c.Assert(share.Create(), chk.IsNil)

	// delete existing share
	ok, err = share.DeleteIfExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)
}

func (s *StorageShareSuite) TestListShares(c *chk.C) {
	cli := getFileClient(c)
	c.Assert(deleteTestShares(cli), chk.IsNil)

	name := randShare()
	share := cli.GetShareReference(name)

	c.Assert(share.Create(), chk.IsNil)

	resp, err := cli.ListShares(ListSharesParameters{
		MaxResults: 5,
		Prefix:     testSharePrefix})
	c.Assert(err, chk.IsNil)

	c.Check(len(resp.Shares), chk.Equals, 1)
	c.Check(resp.Shares[0].Name, chk.Equals, name)

	// clean up via the retrieved share object
	resp.Shares[0].Delete()
}

func (s *StorageShareSuite) TestShareExists(c *chk.C) {
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	ok, err := share.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()

	ok, err = share.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)
}

func (s *StorageShareSuite) TestGetAndSetShareProperties(c *chk.C) {
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())
	quota := rand.Intn(5120)

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	c.Assert(share.Properties.LastModified, chk.Not(chk.Equals), "")

	share.Properties.Quota = quota
	err := share.SetProperties()
	c.Assert(err, chk.IsNil)

	err = share.FetchAttributes()
	c.Assert(err, chk.IsNil)

	c.Assert(share.Properties.Quota, chk.Equals, quota)
}

func (s *StorageShareSuite) TestGetAndSetShareMetadata(c *chk.C) {
	cli := getFileClient(c)
	share1 := cli.GetShareReference(randShare())

	c.Assert(share1.Create(), chk.IsNil)
	defer share1.Delete()

	// by default there should be no metadata
	c.Assert(share1.Metadata, chk.IsNil)
	c.Assert(share1.FetchAttributes(), chk.IsNil)
	c.Assert(share1.Metadata, chk.IsNil)

	share2 := cli.GetShareReference(randShare())
	c.Assert(share2.Create(), chk.IsNil)
	defer share2.Delete()

	c.Assert(share2.Metadata, chk.IsNil)

	mPut := map[string]string{
		"foo":     "bar",
		"bar_baz": "waz qux",
	}

	share2.Metadata = mPut
	c.Assert(share2.SetMetadata(), chk.IsNil)
	c.Check(share2.Metadata, chk.DeepEquals, mPut)

	c.Assert(share2.FetchAttributes(), chk.IsNil)
	c.Check(share2.Metadata, chk.DeepEquals, mPut)

	// Case munging

	mPutUpper := map[string]string{
		"Foo":     "different bar",
		"bar_BAZ": "different waz qux",
	}
	mExpectLower := map[string]string{
		"foo":     "different bar",
		"bar_baz": "different waz qux",
	}

	share2.Metadata = mPutUpper
	c.Assert(share2.SetMetadata(), chk.IsNil)

	c.Check(share2.Metadata, chk.DeepEquals, mPutUpper)
	c.Assert(share2.FetchAttributes(), chk.IsNil)
	c.Check(share2.Metadata, chk.DeepEquals, mExpectLower)
}

func deleteTestShares(cli FileServiceClient) error {
	for {
		resp, err := cli.ListShares(ListSharesParameters{Prefix: testSharePrefix})
		if err != nil {
			return err
		}
		if len(resp.Shares) == 0 {
			break
		}
		for _, c := range resp.Shares {
			share := cli.GetShareReference(c.Name)
			err = share.Delete()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
