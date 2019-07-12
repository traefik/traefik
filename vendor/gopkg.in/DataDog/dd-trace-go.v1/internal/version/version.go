package version

// Tag specifies the current release tag. It needs to be manually
// updated. A test checks that the value of Tag never points to a
// git tag that is older than HEAD.
const Tag = "v1.15.0"
