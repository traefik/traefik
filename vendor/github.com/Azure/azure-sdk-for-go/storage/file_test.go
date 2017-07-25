package storage

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"io"

	chk "gopkg.in/check.v1"
)

type StorageFileSuite struct{}

var _ = chk.Suite(&StorageFileSuite{})

func (s *StorageFileSuite) TestCreateFile(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// create directory structure
	dir1 := root.GetDirectoryReference("one")
	c.Assert(dir1.Create(), chk.IsNil)
	dir2 := dir1.GetDirectoryReference("two")
	c.Assert(dir2.Create(), chk.IsNil)

	// verify file doesn't exist
	file := dir2.GetFileReference("some.file")
	exists, err := file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// create file
	c.Assert(file.Create(1024), chk.IsNil)
	exists, err = file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, true)

	// delete file and verify
	c.Assert(file.Delete(), chk.IsNil)
	exists, err = file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)
}

func (s *StorageFileSuite) TestGetFile(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// create file
	const size = uint64(1024)
	byteStream, _ := newByteStream(size)
	file := root.GetFileReference("some.file")
	c.Assert(file.Create(size), chk.IsNil)

	// fill file with some data
	c.Assert(file.WriteRange(byteStream, FileRange{End: size - 1}, nil), chk.IsNil)

	// set some metadata
	md := map[string]string{
		"something": "somethingvalue",
		"another":   "anothervalue",
	}
	file.Metadata = md
	c.Assert(file.SetMetadata(), chk.IsNil)

	// retrieve full file content and verify
	stream, err := file.DownloadRangeToStream(FileRange{Start: 0, End: size - 1}, false)
	c.Assert(err, chk.IsNil)
	defer stream.Body.Close()
	var b1 [size]byte
	count, _ := stream.Body.Read(b1[:])
	c.Assert(count, chk.Equals, int(size))
	var c1 [size]byte
	bs, _ := newByteStream(size)
	bs.Read(c1[:])
	c.Assert(b1, chk.DeepEquals, c1)

	// retrieve partial file content and verify
	stream, err = file.DownloadRangeToStream(FileRange{Start: size / 2, End: size - 1}, false)
	c.Assert(err, chk.IsNil)
	defer stream.Body.Close()
	var b2 [size / 2]byte
	count, _ = stream.Body.Read(b2[:])
	c.Assert(count, chk.Equals, int(size)/2)
	var c2 [size / 2]byte
	bs, _ = newByteStream(size / 2)
	bs.Read(c2[:])
	c.Assert(b2, chk.DeepEquals, c2)
}

func (s *StorageFileSuite) TestFileRanges(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// create file
	fileSize := uint64(4096)
	byteStream, _ := newByteStream(fileSize)
	file := root.GetFileReference("test.dat")
	c.Assert(file.Create(fileSize), chk.IsNil)

	// verify there are no valid ranges
	ranges, err := file.ListRanges(nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.ContentLength, chk.Equals, fileSize)
	c.Assert(ranges.FileRanges, chk.IsNil)

	// fill entire range and validate
	c.Assert(file.WriteRange(byteStream, FileRange{End: fileSize - 1}, nil), chk.IsNil)
	ranges, err = file.ListRanges(nil)
	c.Assert(err, chk.IsNil)
	c.Assert(len(ranges.FileRanges), chk.Equals, 1)
	c.Assert((ranges.FileRanges[0].End-ranges.FileRanges[0].Start)+1, chk.Equals, fileSize)

	// clear entire range and validate
	c.Assert(file.ClearRange(FileRange{End: fileSize - 1}), chk.IsNil)
	ranges, err = file.ListRanges(nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.FileRanges, chk.IsNil)

	// put partial ranges on 512 byte aligned boundaries
	putRanges := []FileRange{
		{End: 511},
		{Start: 1024, End: 1535},
		{Start: 2048, End: 2559},
		{Start: 3072, End: 3583},
	}

	for _, r := range putRanges {
		byteStream, _ = newByteStream(512)
		err = file.WriteRange(byteStream, r, nil)
		c.Assert(err, chk.IsNil)
	}

	// validate all ranges
	ranges, err = file.ListRanges(nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.FileRanges, chk.DeepEquals, putRanges)

	// validate sub-ranges
	ranges, err = file.ListRanges(&FileRange{Start: 1000, End: 3000})
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.FileRanges, chk.DeepEquals, putRanges[1:3])

	// clear partial range and validate
	c.Assert(file.ClearRange(putRanges[0]), chk.IsNil)
	c.Assert(file.ClearRange(putRanges[2]), chk.IsNil)
	ranges, err = file.ListRanges(nil)
	c.Assert(err, chk.IsNil)
	c.Assert(ranges.FileRanges, chk.HasLen, 2)
	c.Assert(ranges.FileRanges[0], chk.DeepEquals, putRanges[1])
	c.Assert(ranges.FileRanges[1], chk.DeepEquals, putRanges[3])
}

func (s *StorageFileSuite) TestFileProperties(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	fileSize := uint64(512)
	file := root.GetFileReference("test.dat")
	c.Assert(file.Create(fileSize), chk.IsNil)

	// get initial set of properties
	c.Assert(file.Properties.Length, chk.Equals, fileSize)
	c.Assert(file.Properties.Etag, chk.NotNil)

	// set some file properties
	cc := "cachecontrol"
	ct := "mytype"
	enc := "noencoding"
	lang := "neutral"
	disp := "friendly"
	file.Properties.CacheControl = cc
	file.Properties.Type = ct
	file.Properties.Disposition = disp
	file.Properties.Encoding = enc
	file.Properties.Language = lang
	c.Assert(file.SetProperties(), chk.IsNil)

	// retrieve and verify
	c.Assert(file.FetchAttributes(), chk.IsNil)
	c.Assert(file.Properties.CacheControl, chk.Equals, cc)
	c.Assert(file.Properties.Type, chk.Equals, ct)
	c.Assert(file.Properties.Disposition, chk.Equals, disp)
	c.Assert(file.Properties.Encoding, chk.Equals, enc)
	c.Assert(file.Properties.Language, chk.Equals, lang)
}

func (s *StorageFileSuite) TestFileMetadata(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	fileSize := uint64(512)
	file := root.GetFileReference("test.dat")
	c.Assert(file.Create(fileSize), chk.IsNil)

	// get metadata, shouldn't be any
	c.Assert(file.Metadata, chk.HasLen, 0)

	// set some custom metadata
	md := map[string]string{
		"something": "somethingvalue",
		"another":   "anothervalue",
	}
	file.Metadata = md
	c.Assert(file.SetMetadata(), chk.IsNil)

	// retrieve and verify
	c.Assert(file.FetchAttributes(), chk.IsNil)
	c.Assert(file.Metadata, chk.DeepEquals, md)
}

func (s *StorageFileSuite) TestFileMD5(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// create file
	const size = uint64(1024)
	fileSize := uint64(size)
	file := root.GetFileReference("test.dat")
	c.Assert(file.Create(fileSize), chk.IsNil)

	// fill file with some data and MD5 hash
	byteStream, contentMD5 := newByteStream(size)
	c.Assert(file.WriteRange(byteStream, FileRange{End: size - 1}, &contentMD5), chk.IsNil)

	// download file and verify
	stream, err := file.DownloadRangeToStream(FileRange{Start: 0, End: size - 1}, true)
	c.Assert(err, chk.IsNil)
	defer stream.Body.Close()
	c.Assert(stream.ContentMD5, chk.Equals, contentMD5)
}

// returns a byte stream along with a base-64 encoded MD5 hash of its contents
func newByteStream(count uint64) (io.Reader, string) {
	b := make([]uint8, count)
	for i := uint64(0); i < count; i++ {
		b[i] = 0xff
	}

	// create an MD5 hash of the array
	hash := md5.Sum(b)

	return bytes.NewReader(b), base64.StdEncoding.EncodeToString(hash[:])
}

func (s *StorageFileSuite) TestCopyFileSameAccountNoMetaData(c *chk.C) {

	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// create directory structure
	dir1 := root.GetDirectoryReference("one")
	c.Assert(dir1.Create(), chk.IsNil)
	dir2 := dir1.GetDirectoryReference("two")
	c.Assert(dir2.Create(), chk.IsNil)

	// verify file doesn't exist
	file := dir2.GetFileReference("some.file")
	exists, err := file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// create file
	c.Assert(file.Create(1024), chk.IsNil)
	exists, err = file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, true)

	otherFile := dir2.GetFileReference("someother.file")

	// copy the file, no timeout parameter
	err = otherFile.CopyFile(file.URL(), nil)
	c.Assert(err, chk.IsNil)

	// delete file and verify
	c.Assert(file.Delete(), chk.IsNil)
	c.Assert(otherFile.Delete(), chk.IsNil)
	exists, err = file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	exists, err = otherFile.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

}

func (s *StorageFileSuite) TestCopyFileSameAccountTimeout(c *chk.C) {
	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// create directory structure
	dir1 := root.GetDirectoryReference("one")
	c.Assert(dir1.Create(), chk.IsNil)
	dir2 := dir1.GetDirectoryReference("two")
	c.Assert(dir2.Create(), chk.IsNil)

	// verify file doesn't exist
	file := dir2.GetFileReference("some.file")
	exists, err := file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	// create file
	c.Assert(file.Create(1024), chk.IsNil)
	exists, err = file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, true)

	otherFile := dir2.GetFileReference("someother.file")

	options := FileRequestOptions{}
	options.Timeout = 60

	// copy the file, 60 second timeout.
	err = otherFile.CopyFile(file.URL(), &options)
	c.Assert(err, chk.IsNil)

	// delete file and verify
	c.Assert(file.Delete(), chk.IsNil)
	c.Assert(otherFile.Delete(), chk.IsNil)
	exists, err = file.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

	exists, err = otherFile.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(exists, chk.Equals, false)

}

func (s *StorageFileSuite) TestCopyFileMissingFile(c *chk.C) {

	// create share
	cli := getFileClient(c)
	share := cli.GetShareReference(randShare())

	c.Assert(share.Create(), chk.IsNil)
	defer share.Delete()
	root := share.GetRootDirectoryReference()

	// create directory structure
	dir1 := root.GetDirectoryReference("one")
	c.Assert(dir1.Create(), chk.IsNil)

	otherFile := dir1.GetFileReference("someother.file")

	// copy the file, no timeout parameter
	err := otherFile.CopyFile("", nil)
	c.Assert(err, chk.NotNil)
}
