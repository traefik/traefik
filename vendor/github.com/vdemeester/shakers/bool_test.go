package shakers

import (
	"github.com/go-check/check"
)

func init() {
	check.Suite(&BoolCheckerS{})
}

type BoolCheckerS struct{}

func (s *BoolCheckerS) TestTrueInfo(c *check.C) {
	testInfo(c, True, "True", []string{"obtained"})
}

func (s *BoolCheckerS) TestFalseInfo(c *check.C) {
	testInfo(c, False, "False", []string{"obtained"})
}

func (s *BoolCheckerS) TestTrue(c *check.C) {
	testCheck(c, True, false, "obtained value must be a bool.", nil)
	testCheck(c, True, false, "obtained value must be a bool.", "a string")
	testCheck(c, True, false, "obtained value must be a bool.", 1)
	testCheck(c, True, false, "obtained value must be a bool.", struct{}{})

	trueBool := true
	falseBool := false

	testCheck(c, True, false, "", falseBool)
	testCheck(c, True, true, "", trueBool)
}

func (s *BoolCheckerS) TestFalse(c *check.C) {
	testCheck(c, False, false, "obtained value must be a bool.", nil)
	testCheck(c, False, false, "obtained value must be a bool.", "a string")
	testCheck(c, False, false, "obtained value must be a bool.", 1)
	testCheck(c, False, false, "obtained value must be a bool.", struct{}{})

	trueBool := true
	falseBool := false

	testCheck(c, False, false, "", trueBool)
	testCheck(c, False, true, "", falseBool)
}
