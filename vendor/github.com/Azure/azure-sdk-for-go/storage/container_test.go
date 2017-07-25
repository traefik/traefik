package storage

import (
	"crypto/rand"
	"sort"
	"time"

	chk "gopkg.in/check.v1"
)

type ContainerSuite struct{}

var _ = chk.Suite(&ContainerSuite{})

func (s *ContainerSuite) Test_containerBuildPath(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference("foo")
	c.Assert(cnt.buildPath(), chk.Equals, "/foo")
}

func (s *ContainerSuite) TestListContainersPagination(c *chk.C) {
	cli := getBlobClient(c)
	c.Assert(deleteTestContainers(cli), chk.IsNil)

	const n = 5
	const pageSize = 2

	cntNames := []string{}
	for i := 0; i < n; i++ {
		cntNames = append(cntNames, randContainer())
	}
	sort.Strings(cntNames)

	// Create test containers
	created := []Container{}
	for i := 0; i < n; i++ {
		cnt := cli.GetContainerReference(cntNames[i])
		c.Assert(cnt.Create(), chk.IsNil)
		created = append(created, cnt)
		defer cnt.Delete()
	}

	// Paginate results
	seen := []Container{}
	marker := ""
	for {
		resp, err := cli.ListContainers(ListContainersParameters{
			Prefix:     testContainerPrefix,
			MaxResults: pageSize,
			Marker:     marker})

		c.Assert(err, chk.IsNil)

		if len(resp.Containers) > pageSize {
			c.Fatalf("Got a bigger page. Expected: %d, got: %d", pageSize, len(resp.Containers))
		}

		for _, c := range resp.Containers {
			seen = append(seen, c)
		}

		marker = resp.NextMarker
		if marker == "" || len(resp.Containers) == 0 {
			break
		}
	}

	for i := range created {
		c.Assert(seen[i].Name, chk.DeepEquals, created[i].Name)
	}
}

func (s *ContainerSuite) TestContainerExists(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	ok, err := cnt.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	ok, err = cnt.Exists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)
}

func (s *ContainerSuite) TestCreateContainerDeleteContainer(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	c.Assert(cnt.Delete(), chk.IsNil)
}

func (s *ContainerSuite) TestCreateContainerIfNotExists(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	defer cnt.Delete()

	// First create
	ok, err := cnt.CreateIfNotExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)

	// Second create, should not give errors
	ok, err = cnt.CreateIfNotExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)
}

func (s *ContainerSuite) TestDeleteContainerIfExists(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())

	// Nonexisting container
	c.Assert(cnt.Delete(), chk.NotNil)

	ok, err := cnt.DeleteIfExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, false)

	// Existing container
	c.Assert(cnt.Create(), chk.IsNil)
	ok, err = cnt.DeleteIfExists()
	c.Assert(err, chk.IsNil)
	c.Assert(ok, chk.Equals, true)
}

func (s *ContainerSuite) TestListBlobsPagination(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())

	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	blobs := []string{}
	const n = 5
	const pageSize = 2
	for i := 0; i < n; i++ {
		name := randName(5)
		c.Assert(cli.putSingleBlockBlob(cnt.Name, name, []byte("Hello, world!")), chk.IsNil)
		blobs = append(blobs, name)
	}
	sort.Strings(blobs)

	// Paginate
	seen := []string{}
	marker := ""
	for {
		resp, err := cnt.ListBlobs(ListBlobsParameters{
			MaxResults: pageSize,
			Marker:     marker})
		c.Assert(err, chk.IsNil)

		for _, v := range resp.Blobs {
			seen = append(seen, v.Name)
		}

		marker = resp.NextMarker
		if marker == "" || len(resp.Blobs) == 0 {
			break
		}
	}

	// Compare
	c.Assert(seen, chk.DeepEquals, blobs)
}

// listBlobsAsFiles is a helper function to list blobs as "folders" and "files".
func listBlobsAsFiles(cli BlobStorageClient, cnt Container, parentDir string) (folders []string, files []string, err error) {
	var blobParams ListBlobsParameters
	var blobListResponse BlobListResponse

	// Top level "folders"
	blobParams = ListBlobsParameters{
		Delimiter: "/",
		Prefix:    parentDir,
	}

	blobListResponse, err = cnt.ListBlobs(blobParams)
	if err != nil {
		return nil, nil, err
	}

	// These are treated as "folders" under the parentDir.
	folders = blobListResponse.BlobPrefixes

	// "Files"" are blobs which are under the parentDir.
	files = make([]string, len(blobListResponse.Blobs))
	for i := range blobListResponse.Blobs {
		files[i] = blobListResponse.Blobs[i].Name
	}

	return folders, files, nil
}

// TestListBlobsTraversal tests that we can correctly traverse
// blobs in blob storage as if it were a file system by using
// a combination of Prefix, Delimiter, and BlobPrefixes.
//
// Blob storage is flat, but we can *simulate* the file
// system with folders and files using conventions in naming.
// With the blob namedd "/usr/bin/ls", when we use delimiter '/',
// the "ls" would be a "file"; with "/", /usr" and "/usr/bin" being
// the "folders"
//
// NOTE: The use of delimiter (eg forward slash) is extremely fiddly
// and difficult to get right so some discipline in naming and rules
// when using the API is required to get everything to work as expected.
//
// Assuming our delimiter is a forward slash, the rules are:
//
//  - Do use a leading forward slash in blob names to make things
//    consistent and simpler (see further).
//    Note that doing so will show "<no name>" as the only top-level
//    folder in the container in Azure portal, which may look strange.
//
//  - The "folder names" are returned *with trailing forward slash* as per MSDN.
//
//  - The "folder names" will be "absolute paths", e.g. listing things under "/usr/"
//    will return folder names "/usr/bin/".
//
//  - The "file names" are returned as full blob names, e.g. when listing
//    things under "/usr/bin/", the file names will be "/usr/bin/ls" and
//    "/usr/bin/cat".
//
//  - Everything is returned with case-sensitive order as expected in real file system
//    as per MSDN.
//
//  - To list things under a "folder" always use trailing forward slash.
//
//    Example: to list top level folders we use root folder named "" with
//    trailing forward slash, so we use "/".
//
//    Example: to list folders under "/usr", we again append forward slash and
//    so we use "/usr/".
//
//    Because we use leading forward slash we don't need to have different
//    treatment of "get top-level folders" and "get non-top-level folders"
//    scenarios.
func (s *ContainerSuite) TestListBlobsTraversal(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	// Note use of leading forward slash as per naming rules.
	blobsToCreate := []string{
		"/usr/bin/ls",
		"/usr/bin/cat",
		"/usr/lib64/libc.so",
		"/etc/hosts",
		"/etc/init.d/iptables",
	}

	// Create the above blobs
	for _, blobName := range blobsToCreate {
		err := cli.CreateBlockBlob(cnt.Name, blobName)
		c.Assert(err, chk.IsNil)
	}

	var folders []string
	var files []string
	var err error

	// Top level folders and files.
	folders, files, err = listBlobsAsFiles(cli, cnt, "/")
	c.Assert(err, chk.IsNil)
	c.Assert(folders, chk.DeepEquals, []string{"/etc/", "/usr/"})
	c.Assert(files, chk.DeepEquals, []string{})

	// Things under /etc/. Note use of trailing forward slash here as per rules.
	folders, files, err = listBlobsAsFiles(cli, cnt, "/etc/")
	c.Assert(err, chk.IsNil)
	c.Assert(folders, chk.DeepEquals, []string{"/etc/init.d/"})
	c.Assert(files, chk.DeepEquals, []string{"/etc/hosts"})

	// Things under /etc/init.d/
	folders, files, err = listBlobsAsFiles(cli, cnt, "/etc/init.d/")
	c.Assert(err, chk.IsNil)
	c.Assert(folders, chk.DeepEquals, []string(nil))
	c.Assert(files, chk.DeepEquals, []string{"/etc/init.d/iptables"})

	// Things under /usr/
	folders, files, err = listBlobsAsFiles(cli, cnt, "/usr/")
	c.Assert(err, chk.IsNil)
	c.Assert(folders, chk.DeepEquals, []string{"/usr/bin/", "/usr/lib64/"})
	c.Assert(files, chk.DeepEquals, []string{})

	// Things under /usr/bin/
	folders, files, err = listBlobsAsFiles(cli, cnt, "/usr/bin/")
	c.Assert(err, chk.IsNil)
	c.Assert(folders, chk.DeepEquals, []string(nil))
	c.Assert(files, chk.DeepEquals, []string{"/usr/bin/cat", "/usr/bin/ls"})
}

func (s *ContainerSuite) TestListBlobsWithMetadata(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	expectMeta := make(map[string]BlobMetadata)

	// Put 4 blobs with metadata
	for i := 0; i < 4; i++ {
		name := randName(5)
		c.Assert(cli.putSingleBlockBlob(cnt.Name, name, []byte("Hello, world!")), chk.IsNil)
		c.Assert(cli.SetBlobMetadata(cnt.Name, name, map[string]string{
			"Foo":     name,
			"Bar_BAZ": "Waz Qux",
		}, nil), chk.IsNil)
		expectMeta[name] = BlobMetadata{
			"foo":     name,
			"bar_baz": "Waz Qux",
		}
	}

	// Put one more blob with no metadata
	blobWithoutMetadata := randName(5)
	c.Assert(cli.putSingleBlockBlob(cnt.Name, blobWithoutMetadata, []byte("Hello, world!")), chk.IsNil)
	expectMeta[blobWithoutMetadata] = nil

	// Get ListBlobs with include:"metadata"
	resp, err := cnt.ListBlobs(ListBlobsParameters{
		MaxResults: 5,
		Include:    "metadata"})
	c.Assert(err, chk.IsNil)

	respBlobs := make(map[string]Blob)
	for _, v := range resp.Blobs {
		respBlobs[v.Name] = v
	}

	// Verify the metadata is as expected
	for name := range expectMeta {
		c.Check(respBlobs[name].Metadata, chk.DeepEquals, expectMeta[name])
	}
}

func appendContainerPermission(perms ContainerPermissions, accessType ContainerAccessType,
	ID string, start time.Time, expiry time.Time,
	canRead bool, canWrite bool, canDelete bool) ContainerPermissions {

	perms.AccessType = accessType

	if ID != "" {
		capd := ContainerAccessPolicy{
			ID:         ID,
			StartTime:  start,
			ExpiryTime: expiry,
			CanRead:    canRead,
			CanWrite:   canWrite,
			CanDelete:  canDelete,
		}
		perms.AccessPolicies = append(perms.AccessPolicies, capd)
	}
	return perms
}

func (s *ContainerSuite) TestSetContainerPermissionsWithTimeoutSuccessfully(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	perms := ContainerPermissions{}
	perms = appendContainerPermission(perms, ContainerAccessTypeBlob, "GolangRocksOnAzure", now, now.Add(10*time.Hour), true, true, true)

	err := cnt.SetPermissions(perms, 30, "")
	c.Assert(err, chk.IsNil)
}

func (s *ContainerSuite) TestSetContainerPermissionsSuccessfully(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	perms := ContainerPermissions{}
	perms = appendContainerPermission(perms, ContainerAccessTypeBlob, "GolangRocksOnAzure", now, now.Add(10*time.Hour), true, true, true)

	err := cnt.SetPermissions(perms, 0, "")
	c.Assert(err, chk.IsNil)
}

func (s *ContainerSuite) TestSetThenGetContainerPermissionsSuccessfully(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.delete()

	perms := ContainerPermissions{}
	perms = appendContainerPermission(perms, ContainerAccessTypeBlob, "AutoRestIsSuperCool", now, now.Add(10*time.Hour), true, true, true)
	perms = appendContainerPermission(perms, ContainerAccessTypeBlob, "GolangRocksOnAzure", now.Add(20*time.Hour), now.Add(30*time.Hour), true, false, false)
	c.Assert(perms.AccessPolicies, chk.HasLen, 2)

	err := cnt.SetPermissions(perms, 0, "")
	c.Assert(err, chk.IsNil)

	newPerms, err := cnt.GetPermissions(0, "")
	c.Assert(err, chk.IsNil)

	// check container permissions itself.
	c.Assert(newPerms.AccessType, chk.Equals, perms.AccessType)

	// now check policy set.
	c.Assert(newPerms.AccessPolicies, chk.HasLen, 2)

	for i := range perms.AccessPolicies {
		c.Assert(newPerms.AccessPolicies[i].ID, chk.Equals, perms.AccessPolicies[i].ID)

		// test timestamps down the second
		// rounding start/expiry time original perms since the returned perms would have been rounded.
		// so need rounded vs rounded.
		c.Assert(newPerms.AccessPolicies[i].StartTime.UTC().Round(time.Second).Format(time.RFC1123),
			chk.Equals, perms.AccessPolicies[i].StartTime.UTC().Round(time.Second).Format(time.RFC1123))

		c.Assert(newPerms.AccessPolicies[i].ExpiryTime.UTC().Round(time.Second).Format(time.RFC1123),
			chk.Equals, perms.AccessPolicies[i].ExpiryTime.UTC().Round(time.Second).Format(time.RFC1123))

		c.Assert(newPerms.AccessPolicies[i].CanRead, chk.Equals, perms.AccessPolicies[i].CanRead)
		c.Assert(newPerms.AccessPolicies[i].CanWrite, chk.Equals, perms.AccessPolicies[i].CanWrite)
		c.Assert(newPerms.AccessPolicies[i].CanDelete, chk.Equals, perms.AccessPolicies[i].CanDelete)
	}
}

func (s *ContainerSuite) TestSetContainerPermissionsOnlySuccessfully(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	perms := ContainerPermissions{}
	perms = appendContainerPermission(perms, ContainerAccessTypeBlob, "GolangRocksOnAzure", now, now.Add(10*time.Hour), true, true, true)

	err := cnt.SetPermissions(perms, 0, "")
	c.Assert(err, chk.IsNil)
}

func (s *ContainerSuite) TestSetThenGetContainerPermissionsOnlySuccessfully(c *chk.C) {
	cli := getBlobClient(c)
	cnt := cli.GetContainerReference(randContainer())
	c.Assert(cnt.Create(), chk.IsNil)
	defer cnt.Delete()

	perms := ContainerPermissions{}
	perms = appendContainerPermission(perms, ContainerAccessTypeBlob, "", now, now.Add(10*time.Hour), true, true, true)

	err := cnt.SetPermissions(perms, 0, "")
	c.Assert(err, chk.IsNil)

	newPerms, err := cnt.GetPermissions(0, "")
	c.Assert(err, chk.IsNil)

	// check container permissions itself.
	c.Assert(newPerms.AccessType, chk.Equals, perms.AccessType)

	// now check there are NO policies set
	c.Assert(newPerms.AccessPolicies, chk.HasLen, 0)
}

func deleteTestContainers(cli BlobStorageClient) error {
	for {
		resp, err := cli.ListContainers(ListContainersParameters{Prefix: testContainerPrefix})
		if err != nil {
			return err
		}
		if len(resp.Containers) == 0 {
			break
		}
		for _, c := range resp.Containers {
			err = c.Delete()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func randContainer() string {
	return testContainerPrefix + randString(32-len(testContainerPrefix))
}

func randString(n int) string {
	if n <= 0 {
		panic("negative number")
	}
	const alphanum = "0123456789abcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
