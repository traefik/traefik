package shakers

import (
	"github.com/go-check/check"
)

func init() {
	check.Suite(&StringCheckerS{})
}

type StringCheckerS struct{}

func (s *StringCheckerS) TestContains(c *check.C) {
	testInfo(c, Contains, "Contains", []string{"obtained", "substring"})

	testCheck(c, Contains, true, "", "abcd", "bc")
	testCheck(c, Contains, false, "", "abcd", "efg")
	testCheck(c, Contains, false, "", "", "bc")
	testCheck(c, Contains, true, "", "abcd", "")
	testCheck(c, Contains, true, "", "", "")

	testCheck(c, Contains, false, "obtained value is not a string and has no .String().", 12, "1")
	testCheck(c, Contains, false, "substring value must be a string.", "", 1)
}

func (s *StringCheckerS) TestContainsAny(c *check.C) {
	testInfo(c, ContainsAny, "ContainsAny", []string{"obtained", "chars"})

	testCheck(c, ContainsAny, true, "", "abcd", "b")
	testCheck(c, ContainsAny, true, "", "abcd", "b & c")
	testCheck(c, ContainsAny, false, "", "abcd", "e")
	testCheck(c, ContainsAny, false, "", "", "bc")
	testCheck(c, ContainsAny, false, "", "abcd", "")
	testCheck(c, ContainsAny, false, "", "", "")

	testCheck(c, ContainsAny, false, "obtained value is not a string and has no .String().", 12, "1")
	testCheck(c, ContainsAny, false, "chars value must be a string.", "", 1)
}

func (s *StringCheckerS) TestHasPrefix(c *check.C) {
	testInfo(c, HasPrefix, "HasPrefix", []string{"obtained", "prefix"})

	testCheck(c, HasPrefix, true, "", "abcd", "ab")
	testCheck(c, HasPrefix, false, "", "abcd", "efg")
	testCheck(c, HasPrefix, false, "", "", "bc")
	testCheck(c, HasPrefix, true, "", "abcd", "")
	testCheck(c, HasPrefix, true, "", "", "")

	testCheck(c, HasPrefix, false, "obtained value is not a string and has no .String().", 12, "1")
	testCheck(c, HasPrefix, false, "prefix value must be a string.", "", 1)
}

func (s *StringCheckerS) TestHasSuffix(c *check.C) {
	testInfo(c, HasSuffix, "HasSuffix", []string{"obtained", "suffix"})

	testCheck(c, HasSuffix, true, "", "abcd", "cd")
	testCheck(c, HasSuffix, false, "", "abcd", "efg")
	testCheck(c, HasSuffix, false, "", "", "bc")
	testCheck(c, HasSuffix, true, "", "abcd", "")
	testCheck(c, HasSuffix, true, "", "", "")

	testCheck(c, HasSuffix, false, "obtained value is not a string and has no .String().", 12, "1")
	testCheck(c, HasSuffix, false, "suffix value must be a string.", "", 1)
}

func (s *StringCheckerS) TestEqualFold(c *check.C) {
	testInfo(c, EqualFold, "EqualFold", []string{"obtained", "expected"})

	testCheck(c, EqualFold, true, "", "abcd", "ABCD")
	testCheck(c, EqualFold, true, "", "abcd", "AbCd")
	testCheck(c, EqualFold, true, "", "", "")
	testCheck(c, EqualFold, true, "", "🐸", "🐸")
	testCheck(c, EqualFold, false, "", "🐭", "🐹")
	testCheck(c, EqualFold, false, "", "abcd", "acde")
	testCheck(c, EqualFold, false, "", "", "bc")
	testCheck(c, EqualFold, false, "", "abcd", "")

	testCheck(c, EqualFold, false, "obtained value is not a string and has no .String().", 12, "1")
	testCheck(c, EqualFold, false, "expected value must be a string.", "", 1)
}

func (s *StringCheckerS) TestCount(c *check.C) {
	testInfo(c, Count, "Count", []string{"obtained", "sep", "expected"})

	testCheck(c, Count, true, "", "abcd", "a", 1)
	testCheck(c, Count, true, "", "abcdAbcd", "a", 1)
	testCheck(c, Count, true, "", "abcd", "e", 0)
	testCheck(c, Count, true, "", "ABCD", "a", 0)
	testCheck(c, Count, true, "", "aaaaa", "a", 5)
	testCheck(c, Count, true, "", "🐭🐹", "🐹", 1)
	testCheck(c, Count, false, "", "aaaaa", "a", 1)
	testCheck(c, Count, false, "", "abcd", "a", 0)
	testCheck(c, Count, false, "", "🐭🐹", "a", 1)

	testCheck(c, Count, false, "obtained value is not a string and has no .String().", 12, "1", 1)
	testCheck(c, Count, false, "sep value must be a string.", "", 1, 1)
	testCheck(c, Count, false, "", "", "", "")
}

func (s *StringCheckerS) TestIndex(c *check.C) {
	testInfo(c, Index, "Index", []string{"obtained", "sep", "expected"})

	testCheck(c, Index, true, "", "abcd", "a", 0)
	testCheck(c, Index, true, "", "abcdAbcd", "a", 0)
	testCheck(c, Index, true, "", "abcdAbcd", "A", 4)
	testCheck(c, Index, true, "", "abcd", "e", -1)
	testCheck(c, Index, true, "", "ABCD", "a", -1)
	testCheck(c, Index, true, "", "aaaaa", "a", 0)
	testCheck(c, Index, false, "", "dcba", "a", 0)
	testCheck(c, Index, false, "", "abcd", "d", 0)

	testCheck(c, Index, false, "obtained value is not a string and has no .String().", 12, "1", 1)
	testCheck(c, Index, false, "sep value must be a string.", "", 1, 1)
	testCheck(c, Index, false, "", "", "", "")
}

func (s *StringCheckerS) TestIndexAny(c *check.C) {
	testInfo(c, IndexAny, "IndexAny", []string{"obtained", "chars", "expected"})

	testCheck(c, IndexAny, true, "", "abcd", "b", 1)
	testCheck(c, IndexAny, true, "", "abcdAbcd", "b & c", 1)
	testCheck(c, IndexAny, true, "", "abcdAbcd", "bc", 1)
	testCheck(c, IndexAny, true, "", "abcdAbcde", "A & e", 4)
	testCheck(c, IndexAny, true, "", "abcdAbcde", "Ae", 4)
	testCheck(c, IndexAny, true, "", "abcd", "e", -1)
	testCheck(c, IndexAny, true, "", "ABCD", "a", -1)
	testCheck(c, IndexAny, false, "", "abcd", "d", 0)
	testCheck(c, IndexAny, false, "", "dcba", "a & b", 0)
	testCheck(c, IndexAny, false, "", "dcba", "ab", 0)

	testCheck(c, IndexAny, false, "obtained value is not a string and has no .String().", 12, "1", 1)
	testCheck(c, IndexAny, false, "chars value must be a string.", "", 1, 1)
	testCheck(c, IndexAny, false, "", "", "", "")
}

func (s *StringCheckerS) TestIsLower(c *check.C) {
	testInfo(c, IsLower, "IsLower", []string{"obtained"})

	testCheck(c, IsLower, true, "", "abcd")
	testCheck(c, IsLower, true, "", "1234")
	testCheck(c, IsLower, true, "", "abcd abcde")
	testCheck(c, IsLower, true, "", "хлеб")
	testCheck(c, IsLower, false, "", "ABCD")
	testCheck(c, IsLower, false, "", "Abcd")
	testCheck(c, IsLower, false, "", "ABCD ABCD")

	testCheck(c, IsLower, false, "obtained value is not a string and has no .String().", 12)
}

func (s *StringCheckerS) TestIsUpper(c *check.C) {
	testInfo(c, IsUpper, "IsUpper", []string{"obtained"})

	testCheck(c, IsUpper, true, "", "1234")
	testCheck(c, IsUpper, true, "", "ABCD")
	testCheck(c, IsUpper, true, "", "ABCD ABCD")
	testCheck(c, IsUpper, true, "", "ХЛЕБ")
	testCheck(c, IsUpper, false, "", "abcd")
	testCheck(c, IsUpper, false, "", "Abcd")
	testCheck(c, IsUpper, false, "", "abcd abcde")

	testCheck(c, IsUpper, false, "obtained value is not a string and has no .String().", 12)
}
