package storage

import chk "gopkg.in/check.v1"

type StorageDirSuite struct{}

var _ = chk.Suite(&StorageDirSuite{})

func (s *StorageDirSuite) TestListDirsAndFiles(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// list contents, should be empty
	resp, err := root.ListDirsAndFiles(ListDirsAndFilesParameters{})
	c.Assert(err, chk.IsNil)
	c.Assert(resp.Directories, chk.IsNil)
	c.Assert(resp.Files, chk.IsNil)

	// create a directory and a file
	dir := root.GetDirectoryReference("SomeDirectory")
	file := root.GetFileReference("foo.file")
	c.Assert(dir.Create(), chk.IsNil)
	c.Assert(file.Create(512), chk.IsNil)

	// list contents
	resp, err = root.ListDirsAndFiles(ListDirsAndFilesParameters{})
	c.Assert(err, chk.IsNil)
	c.Assert(len(resp.Directories), chk.Equals, 1)
	c.Assert(len(resp.Files), chk.Equals, 1)
	c.Assert(resp.Directories[0].Name, chk.Equals, dir.Name)
	c.Assert(resp.Files[0].Name, chk.Equals, file.Name)

	// delete file
	del, err := file.DeleteIfExists()
	c.Assert(err, chk.IsNil)
	c.Assert(del, chk.Equals, true)

	// attempt to delete again
	del, err = file.DeleteIfExists()
	c.Assert(err, chk.IsNil)
	c.Assert(del, chk.Equals, false)
}

func (s *StorageDirSuite) TestCreateDirectory(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// directory shouldn't exist
	dir := root.GetDirectoryReference("SomeDirectory")
	exists, err := dir.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// create directory
	exists, err = dir.CreateIfNotExists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, true)

	// try to create again, should fail
	c.Assert(dir.Create(), chk.NotNil)
	exists, err = dir.CreateIfNotExists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// check properties
	c.Assert(dir.Properties.Etag, chk.Not(chk.Equals), "")
	c.Assert(dir.Properties.LastModified, chk.Not(chk.Equals), "")

	// delete directory and verify
	c.Assert(dir.Delete(), chk.IsNil)
	exists, err = dir.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)
}

func (s *StorageDirSuite) TestDirectoryMetadata(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	dir := root.GetDirectoryReference("testdir")
	c.Assert(dir.Create(), chk.IsNil)

	// get metadata, shouldn't be any
	c.Assert(dir.Metadata, chk.IsNil)

	// set some custom metadata
	md := map[string]string{
		"something": "somethingvalue",
		"another":   "anothervalue",
	}
	dir.Metadata = md
	c.Assert(dir.SetMetadata(), chk.IsNil)

	// retrieve and verify
	c.Assert(dir.FetchAttributes(), chk.IsNil)
	c.Assert(dir.Metadata, chk.DeepEquals, md)
}
