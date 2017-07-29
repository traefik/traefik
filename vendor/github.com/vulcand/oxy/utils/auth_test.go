package utils

import (
	. "gopkg.in/check.v1"
)

type AuthSuite struct {
}

var _ = Suite(&AuthSuite{})

//Just to make sure we don't panic, return err and not
//username and pass and cover the function
func (s *AuthSuite) TestParseBadHeaders(c *C) {
	headers := []string{
		//just empty string
		"",
		//missing auth type
		"justplainstring",
		//unknown auth type
		"Whut justplainstring",
		//invalid base64
		"Basic Shmasic",
		//random encoded string
		"Basic YW55IGNhcm5hbCBwbGVhcw==",
	}
	for _, h := range headers {
		_, err := ParseAuthHeader(h)
		c.Assert(err, NotNil)
	}
}

//Just to make sure we don't panic, return err and not
//username and pass and cover the function
func (s *AuthSuite) TestParseSuccess(c *C) {
	headers := []struct {
		Header   string
		Expected BasicAuth
	}{
		{
			"Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==",
			BasicAuth{Username: "Aladdin", Password: "open sesame"},
		},
		// Make sure that String() produces valid header
		{
			(&BasicAuth{Username: "Alice", Password: "Here's bob"}).String(),
			BasicAuth{Username: "Alice", Password: "Here's bob"},
		},
		//empty pass
		{
			"Basic QWxhZGRpbjo=",
			BasicAuth{Username: "Aladdin", Password: ""},
		},
	}
	for _, h := range headers {
		request, err := ParseAuthHeader(h.Header)
		c.Assert(err, IsNil)
		c.Assert(request.Username, Equals, h.Expected.Username)
		c.Assert(request.Password, Equals, h.Expected.Password)

	}
}
