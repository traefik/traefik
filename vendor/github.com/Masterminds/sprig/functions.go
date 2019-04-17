package sprig

import (
	"errors"
	"html/template"
	"os"
	"path"
	"strconv"
	"strings"
	ttemplate "text/template"
	"time"

	util "github.com/Masterminds/goutils"
	"github.com/huandu/xstrings"
)

// Produce the function map.
//
// Use this to pass the functions into the template engine:
//
// 	tpl := template.New("foo").Funcs(sprig.FuncMap()))
//
func FuncMap() template.FuncMap {
	return HtmlFuncMap()
}

// HermeticTextFuncMap returns a 'text/template'.FuncMap with only repeatable functions.
func HermeticTxtFuncMap() ttemplate.FuncMap {
	r := TxtFuncMap()
	for _, name := range nonhermeticFunctions {
		delete(r, name)
	}
	return r
}

// HermeticHtmlFuncMap returns an 'html/template'.Funcmap with only repeatable functions.
func HermeticHtmlFuncMap() template.FuncMap {
	r := HtmlFuncMap()
	for _, name := range nonhermeticFunctions {
		delete(r, name)
	}
	return r
}

// TextFuncMap returns a 'text/template'.FuncMap
func TxtFuncMap() ttemplate.FuncMap {
	return ttemplate.FuncMap(GenericFuncMap())
}

// HtmlFuncMap returns an 'html/template'.Funcmap
func HtmlFuncMap() template.FuncMap {
	return template.FuncMap(GenericFuncMap())
}

// GenericFuncMap returns a copy of the basic function map as a map[string]interface{}.
func GenericFuncMap() map[string]interface{} {
	gfm := make(map[string]interface{}, len(genericMap))
	for k, v := range genericMap {
		gfm[k] = v
	}
	return gfm
}

// These functions are not guaranteed to evaluate to the same result for given input, because they
// refer to the environemnt or global state.
var nonhermeticFunctions = []string{
	// Date functions
	"date",
	"date_in_zone",
	"date_modify",
	"now",
	"htmlDate",
	"htmlDateInZone",
	"dateInZone",
	"dateModify",

	// Strings
	"randAlphaNum",
	"randAlpha",
	"randAscii",
	"randNumeric",
	"uuidv4",

	// OS
	"env",
	"expandenv",
}

var genericMap = map[string]interface{}{
	"hello": func() string { return "Hello!" },

	// Date functions
	"date":           date,
	"date_in_zone":   dateInZone,
	"date_modify":    dateModify,
	"now":            func() time.Time { return time.Now() },
	"htmlDate":       htmlDate,
	"htmlDateInZone": htmlDateInZone,
	"dateInZone":     dateInZone,
	"dateModify":     dateModify,
	"ago":            dateAgo,
	"toDate":         toDate,

	// Strings
	"abbrev":     abbrev,
	"abbrevboth": abbrevboth,
	"trunc":      trunc,
	"trim":       strings.TrimSpace,
	"upper":      strings.ToUpper,
	"lower":      strings.ToLower,
	"title":      strings.Title,
	"untitle":    untitle,
	"substr":     substring,
	// Switch order so that "foo" | repeat 5
	"repeat": func(count int, str string) string { return strings.Repeat(str, count) },
	// Deprecated: Use trimAll.
	"trimall": func(a, b string) string { return strings.Trim(b, a) },
	// Switch order so that "$foo" | trimall "$"
	"trimAll":      func(a, b string) string { return strings.Trim(b, a) },
	"trimSuffix":   func(a, b string) string { return strings.TrimSuffix(b, a) },
	"trimPrefix":   func(a, b string) string { return strings.TrimPrefix(b, a) },
	"nospace":      util.DeleteWhiteSpace,
	"initials":     initials,
	"randAlphaNum": randAlphaNumeric,
	"randAlpha":    randAlpha,
	"randAscii":    randAscii,
	"randNumeric":  randNumeric,
	"swapcase":     util.SwapCase,
	"shuffle":      xstrings.Shuffle,
	"snakecase":    xstrings.ToSnakeCase,
	"camelcase":    xstrings.ToCamelCase,
	"kebabcase":    xstrings.ToKebabCase,
	"wrap":         func(l int, s string) string { return util.Wrap(s, l) },
	"wrapWith":     func(l int, sep, str string) string { return util.WrapCustom(str, l, sep, true) },
	// Switch order so that "foobar" | contains "foo"
	"contains":  func(substr string, str string) bool { return strings.Contains(str, substr) },
	"hasPrefix": func(substr string, str string) bool { return strings.HasPrefix(str, substr) },
	"hasSuffix": func(substr string, str string) bool { return strings.HasSuffix(str, substr) },
	"quote":     quote,
	"squote":    squote,
	"cat":       cat,
	"indent":    indent,
	"nindent":   nindent,
	"replace":   replace,
	"plural":    plural,
	"sha1sum":   sha1sum,
	"sha256sum": sha256sum,
	"adler32sum": adler32sum,
	"toString":  strval,

	// Wrap Atoi to stop errors.
	"atoi":    func(a string) int { i, _ := strconv.Atoi(a); return i },
	"int64":   toInt64,
	"int":     toInt,
	"float64": toFloat64,

	//"gt": func(a, b int) bool {return a > b},
	//"gte": func(a, b int) bool {return a >= b},
	//"lt": func(a, b int) bool {return a < b},
	//"lte": func(a, b int) bool {return a <= b},

	// split "/" foo/bar returns map[int]string{0: foo, 1: bar}
	"split":     split,
	"splitList": func(sep, orig string) []string { return strings.Split(orig, sep) },
	// splitn "/" foo/bar/fuu returns map[int]string{0: foo, 1: bar/fuu}
	"splitn":    splitn,
	"toStrings": strslice,

	"until":     until,
	"untilStep": untilStep,

	// VERY basic arithmetic.
	"add1": func(i interface{}) int64 { return toInt64(i) + 1 },
	"add": func(i ...interface{}) int64 {
		var a int64 = 0
		for _, b := range i {
			a += toInt64(b)
		}
		return a
	},
	"sub": func(a, b interface{}) int64 { return toInt64(a) - toInt64(b) },
	"div": func(a, b interface{}) int64 { return toInt64(a) / toInt64(b) },
	"mod": func(a, b interface{}) int64 { return toInt64(a) % toInt64(b) },
	"mul": func(a interface{}, v ...interface{}) int64 {
		val := toInt64(a)
		for _, b := range v {
			val = val * toInt64(b)
		}
		return val
	},
	"biggest": max,
	"max":     max,
	"min":     min,
	"ceil":    ceil,
	"floor":   floor,
	"round":   round,

	// string slices. Note that we reverse the order b/c that's better
	// for template processing.
	"join":      join,
	"sortAlpha": sortAlpha,

	// Defaults
	"default":      dfault,
	"empty":        empty,
	"coalesce":     coalesce,
	"compact":      compact,
	"toJson":       toJson,
	"toPrettyJson": toPrettyJson,
	"ternary":      ternary,

	// Reflection
	"typeOf":     typeOf,
	"typeIs":     typeIs,
	"typeIsLike": typeIsLike,
	"kindOf":     kindOf,
	"kindIs":     kindIs,

	// OS:
	"env":       func(s string) string { return os.Getenv(s) },
	"expandenv": func(s string) string { return os.ExpandEnv(s) },

	// File Paths:
	"base":  path.Base,
	"dir":   path.Dir,
	"clean": path.Clean,
	"ext":   path.Ext,
	"isAbs": path.IsAbs,

	// Encoding:
	"b64enc": base64encode,
	"b64dec": base64decode,
	"b32enc": base32encode,
	"b32dec": base32decode,

	// Data Structures:
	"tuple":  list, // FIXME: with the addition of append/prepend these are no longer immutable.
	"list":   list,
	"dict":   dict,
	"set":    set,
	"unset":  unset,
	"hasKey": hasKey,
	"pluck":  pluck,
	"keys":   keys,
	"pick":   pick,
	"omit":   omit,
	"merge":  merge,
	"mergeOverwrite": mergeOverwrite,
	"values": values,

	"append": push, "push": push,
	"prepend": prepend,
	"first":   first,
	"rest":    rest,
	"last":    last,
	"initial": initial,
	"reverse": reverse,
	"uniq":    uniq,
	"without": without,
	"has":     has,
	"slice":   slice,

	// Crypto:
	"genPrivateKey":     generatePrivateKey,
	"derivePassword":    derivePassword,
	"buildCustomCert":   buildCustomCertificate,
	"genCA":             generateCertificateAuthority,
	"genSelfSignedCert": generateSelfSignedCertificate,
	"genSignedCert":     generateSignedCertificate,

	// UUIDs:
	"uuidv4": uuidv4,

	// SemVer:
	"semver":        semver,
	"semverCompare": semverCompare,

	// Flow Control:
	"fail": func(msg string) (string, error) { return "", errors.New(msg) },

	// Regex
	"regexMatch":             regexMatch,
	"regexFindAll":           regexFindAll,
	"regexFind":              regexFind,
	"regexReplaceAll":        regexReplaceAll,
	"regexReplaceAllLiteral": regexReplaceAllLiteral,
	"regexSplit":             regexSplit,
}
