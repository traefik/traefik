package plugin

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "gopkg.in/check.v1"
	"testing"
)

func TestMdw(t *testing.T) { TestingT(t) }

type MiddlewareSuite struct {
}

var _ = Suite(&MiddlewareSuite{})

func (s *MiddlewareSuite) TestVerifySignatureOK(c *C) {
	fn := func(TestMiddleware) (Middleware, error) { return nil, nil }
	c.Assert(verifySignature(fn), IsNil)
}

func (s *MiddlewareSuite) TestVerifySignatureIncompatibleFunctions(c *C) {
	// Not a function
	c.Assert(verifySignature(nil), NotNil)

	// Pointers are not ok
	fn := func(*TestMiddleware) (Middleware, error) { return nil, nil }
	c.Assert(verifySignature(fn), NotNil)

	// Just one input arg is needed
	fn1 := func(TestMiddleware, int) (Middleware, error) { return nil, nil }
	c.Assert(verifySignature(fn1), NotNil)

	// Return arguments are incorrect
	fn2 := func(TestMiddleware) Middleware { return nil }
	c.Assert(verifySignature(fn2), NotNil)

	// First return argument is not middleware
	fn3 := func(TestMiddleware) (int, error) { return 0, nil }
	c.Assert(verifySignature(fn3), NotNil)

	// Second return argument is not error
	fn4 := func(TestMiddleware) (Middleware, int) { return nil, 0 }
	c.Assert(verifySignature(fn4), NotNil)

}

func (s *MiddlewareSuite) TestFromJSONOk(c *C) {

	correct := TestMiddleware{Field: "hi"}
	bytes, err := json.Marshal(correct)
	c.Assert(err, IsNil)

	out, err := GetSpec().FromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, DeepEquals, out)
}

func (s *MiddlewareSuite) TestFromJSONBadInstance(c *C) {
	incorrect := TestMiddleware{}
	bytes, err := json.Marshal(incorrect)
	c.Assert(err, IsNil)

	out, err := GetSpec().FromJSON(bytes)
	c.Assert(err, NotNil)
	c.Assert(out, IsNil)
}

func (s *MiddlewareSuite) TestFromJSONBadJSON(c *C) {
	out, err := GetSpec().FromJSON([]byte(" what?"))
	c.Assert(err, NotNil)
	c.Assert(out, IsNil)
}

func (s *MiddlewareSuite) TestRegistrySetGet(c *C) {
	r := NewRegistry()
	c.Assert(r.AddSpec(GetSpec()), IsNil)
	c.Assert(r.GetSpec(GetSpec().Type).Type, DeepEquals, GetSpec().Type)
	c.Assert(len(r.GetSpecs()), Equals, 1)
}

func (s *MiddlewareSuite) TestRegistryAddSpecTwice(c *C) {
	r := NewRegistry()
	c.Assert(r.AddSpec(GetSpec()), IsNil)
	c.Assert(r.AddSpec(GetSpec()), NotNil)
}

func (s *MiddlewareSuite) TestRegistryAddNilSpec(c *C) {
	r := NewRegistry()
	c.Assert(r.AddSpec(nil), NotNil)
}

func (s *MiddlewareSuite) TestRegistryAddSpecBadSignature(c *C) {
	r := NewRegistry()
	c.Assert(r.AddSpec(&MiddlewareSpec{}), NotNil)
}

func (s *MiddlewareSuite) TestRegistryNotFoundMiddlewareGetSet(c *C) {
	r := NewRegistry()
	correct := &TestMiddleware{Field: "notfound"}
	r.AddNotFoundMiddleware(correct)
	c.Assert(r.GetNotFoundMiddleware(), Equals, correct)
}

type TestMiddleware struct {
	Field string
	next  http.Handler
}

func (*TestMiddleware) NewHandler(next http.Handler) (http.Handler, error) {
	return &TestMiddleware{next: next}, nil
}

func (*TestMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
}

func GetSpec() *MiddlewareSpec {
	return &MiddlewareSpec{
		Type: "test",
		FromOther: func(b TestMiddleware) (Middleware, error) {
			if b.Field == "" {
				return nil, fmt.Errorf("can not be empty")
			}
			return &TestMiddleware{Field: b.Field}, nil
		},
	}
}
