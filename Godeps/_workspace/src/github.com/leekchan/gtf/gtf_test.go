package gtf

import (
	"bytes"
	"html/template"
	"testing"
)

func AssertEqual(t *testing.T, buffer *bytes.Buffer, testString string) {
	if buffer.String() != testString {
		t.Errorf("Expected %s, got %s", testString, buffer.String())
	}
	buffer.Reset()
}

func ParseTest(buffer *bytes.Buffer, body string, data interface{}) {
	tpl := New("test")
	tpl.Parse(body)
	tpl.Execute(buffer, data)
}

func CustomParseTest(funcMap map[string]interface{}, buffer *bytes.Buffer, body string, data interface{}) {
	tpl := template.New("test").Funcs(funcMap)
	tpl.Parse(body)
	tpl.Execute(buffer, data)
}

func TestGtfFuncMap(t *testing.T) {
	var buffer bytes.Buffer

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | replace \" \" }}", "")
	AssertEqual(t, &buffer, "TheGoProgrammingLanguage")

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | default \"default value\" }}", "")
	AssertEqual(t, &buffer, "The Go Programming Language")

	ParseTest(&buffer, "{{ \"\" | default \"default value\" }}", "")
	AssertEqual(t, &buffer, "default value")

	ParseTest(&buffer, "{{ . | default \"default value\" }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "[go python ruby]")

	ParseTest(&buffer, "{{ . | default 10 }}", []int{})
	AssertEqual(t, &buffer, "10")

	ParseTest(&buffer, "{{ . | default \"empty\" }}", false)
	AssertEqual(t, &buffer, "empty")

	ParseTest(&buffer, "{{ . | default \"empty\" }}", 1)
	AssertEqual(t, &buffer, "1")

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | length }}", "")
	AssertEqual(t, &buffer, "27")

	ParseTest(&buffer, "{{ \"안녕하세요\" | length }}", "")
	AssertEqual(t, &buffer, "5")

	ParseTest(&buffer, "{{ . | length }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "3")

	ParseTest(&buffer, "{{ . | length }}", false)
	AssertEqual(t, &buffer, "0")

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | lower }}", "")
	AssertEqual(t, &buffer, "the go programming language")

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | upper }}", "")
	AssertEqual(t, &buffer, "THE GO PROGRAMMING LANGUAGE")

	ParseTest(&buffer, "{{ \"안녕하세요. 반갑습니다.\" | truncatechars 12 }}", "")
	AssertEqual(t, &buffer, "안녕하세요. 반갑...")

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | truncatechars 12 }}", "")
	AssertEqual(t, &buffer, "The Go Pr...")

	ParseTest(&buffer, "{{ \"안녕하세요. The Go Programming Language\" | truncatechars 30 }}", "")
	AssertEqual(t, &buffer, "안녕하세요. The Go Programming L...")

	ParseTest(&buffer, "{{ \"The\" | truncatechars 30 }}", "")
	AssertEqual(t, &buffer, "The")

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | truncatechars 3 }}", "")
	AssertEqual(t, &buffer, "The")

	ParseTest(&buffer, "{{ \"The Go\" | truncatechars 6 }}", "")
	AssertEqual(t, &buffer, "The Go")

	ParseTest(&buffer, "{{ \"The Go\" | truncatechars 30 }}", "")
	AssertEqual(t, &buffer, "The Go")

	ParseTest(&buffer, "{{ \"The Go\" | truncatechars 0 }}", "")
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ \"The Go\" | truncatechars -1 }}", "")
	AssertEqual(t, &buffer, "The Go")

	ParseTest(&buffer, "{{ \"http://www.example.org/foo?a=b&c=d\" | urlencode }}", "")
	AssertEqual(t, &buffer, "http%3A%2F%2Fwww.example.org%2Ffoo%3Fa%3Db%26c%3Dd")

	ParseTest(&buffer, "{{ \"The Go Programming Language\" | wordcount }}", "")
	AssertEqual(t, &buffer, "4")

	ParseTest(&buffer, "{{ \"      The      Go       Programming      Language        \" | wordcount }}", "")
	AssertEqual(t, &buffer, "4")

	ParseTest(&buffer, "{{ 21 | divisibleby 3 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 21 | divisibleby 4 }}", "")
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ 3.0 | divisibleby 3 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 3.0 | divisibleby 1.5 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ . | divisibleby 1.5 }}", uint(300))
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 12 | divisibleby . }}", uint(3))
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ 21 | divisibleby 4 }}", "")
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ false | divisibleby 3 }}", "")
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ 3 | divisibleby false }}", "")
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ \"Go\" | lengthis 2 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ \"안녕하세요.\" | lengthis 6 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ \"안녕하세요. Go!\" | lengthis 10 }}", "")
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ . | lengthis 3 }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "true")

	ParseTest(&buffer, "{{ . | lengthis 3 }}", false)
	AssertEqual(t, &buffer, "false")

	ParseTest(&buffer, "{{ \"       The Go Programming Language     \" | trim }}", "")
	AssertEqual(t, &buffer, "The Go Programming Language")

	ParseTest(&buffer, "{{ \"the go programming language\" | capfirst }}", "")
	AssertEqual(t, &buffer, "The go programming language")

	ParseTest(&buffer, "You have 0 message{{ 0 | pluralize \"s\" }}", "")
	AssertEqual(t, &buffer, "You have 0 messages")

	ParseTest(&buffer, "You have 1 message{{ 1 | pluralize \"s\" }}", "")
	AssertEqual(t, &buffer, "You have 1 message")

	ParseTest(&buffer, "0 cand{{ 0 | pluralize \"y,ies\" }}", "")
	AssertEqual(t, &buffer, "0 candies")

	ParseTest(&buffer, "1 cand{{ 1 | pluralize \"y,ies\" }}", "")
	AssertEqual(t, &buffer, "1 candy")

	ParseTest(&buffer, "2 cand{{ 2 | pluralize \"y,ies\" }}", "")
	AssertEqual(t, &buffer, "2 candies")

	ParseTest(&buffer, "{{ 2 | pluralize \"y,ies,s\" }}", "")
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "2 cand{{ . | pluralize \"y,ies\" }}", uint(2))
	AssertEqual(t, &buffer, "2 candies")

	ParseTest(&buffer, "1 cand{{ . | pluralize \"y,ies\" }}", uint(1))
	AssertEqual(t, &buffer, "1 candy")

	ParseTest(&buffer, "{{ . | pluralize \"y,ies\" }}", "test")
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ true | yesno \"yes~\" \"no~\" }}", "")
	AssertEqual(t, &buffer, "yes~")

	ParseTest(&buffer, "{{ false | yesno \"yes~\" \"no~\" }}", "")
	AssertEqual(t, &buffer, "no~")

	ParseTest(&buffer, "{{ \"Go\" | rjust 10 }}", "")
	AssertEqual(t, &buffer, "        Go")

	ParseTest(&buffer, "{{ \"안녕하세요\" | rjust 10 }}", "")
	AssertEqual(t, &buffer, "     안녕하세요")

	ParseTest(&buffer, "{{ \"Go\" | ljust 10 }}", "")
	AssertEqual(t, &buffer, "Go        ")

	ParseTest(&buffer, "{{ \"안녕하세요\" | ljust 10 }}", "")
	AssertEqual(t, &buffer, "안녕하세요     ")

	ParseTest(&buffer, "{{ \"Go\" | center 10 }}", "")
	AssertEqual(t, &buffer, "    Go    ")

	ParseTest(&buffer, "{{ \"안녕하세요\" | center 10 }}", "")
	AssertEqual(t, &buffer, "  안녕하세요   ")

	ParseTest(&buffer, "{{ 123456789 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "117.7 MB")

	ParseTest(&buffer, "{{ 234 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "234 bytes")

	ParseTest(&buffer, "{{ 12345 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "12.1 KB")

	ParseTest(&buffer, "{{ 554832114 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "529.1 MB")

	ParseTest(&buffer, "{{ 1048576 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "1 MB")

	ParseTest(&buffer, "{{ 14868735121 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "13.8 GB")

	ParseTest(&buffer, "{{ 14868735121365 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "13.5 TB")

	ParseTest(&buffer, "{{ 1486873512136523 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "1.3 PB")

	ParseTest(&buffer, "{{ 12345.35335 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "12.1 KB")

	ParseTest(&buffer, "{{ 4294967293 | filesizeformat }}", "")
	AssertEqual(t, &buffer, "4 GB")

	ParseTest(&buffer, "{{ \"Go\" | filesizeformat }}", "")
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ . | filesizeformat }}", uint(500))
	AssertEqual(t, &buffer, "500 bytes")

	ParseTest(&buffer, "{{ . | apnumber }}", uint(500))
	AssertEqual(t, &buffer, "500")

	ParseTest(&buffer, "{{ . | apnumber }}", uint(7))
	AssertEqual(t, &buffer, "seven")

	ParseTest(&buffer, "{{ . | apnumber }}", int8(3))
	AssertEqual(t, &buffer, "three")

	ParseTest(&buffer, "{{ 2 | apnumber }}", "")
	AssertEqual(t, &buffer, "two")

	ParseTest(&buffer, "{{ 1000 | apnumber }}", "")
	AssertEqual(t, &buffer, "1000")

	ParseTest(&buffer, "{{ 1000 | intcomma }}", "")
	AssertEqual(t, &buffer, "1,000")

	ParseTest(&buffer, "{{ -1000 | intcomma }}", "")
	AssertEqual(t, &buffer, "-1,000")

	ParseTest(&buffer, "{{ 1578652313 | intcomma }}", "")
	AssertEqual(t, &buffer, "1,578,652,313")

	ParseTest(&buffer, "{{ . | intcomma }}", uint64(12315358198))
	AssertEqual(t, &buffer, "12,315,358,198")

	ParseTest(&buffer, "{{ . | intcomma }}", 25.352)
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ 1 | ordinal }}", "")
	AssertEqual(t, &buffer, "1st")

	ParseTest(&buffer, "{{ 2 | ordinal }}", "")
	AssertEqual(t, &buffer, "2nd")

	ParseTest(&buffer, "{{ 3 | ordinal }}", "")
	AssertEqual(t, &buffer, "3rd")

	ParseTest(&buffer, "{{ 11 | ordinal }}", "")
	AssertEqual(t, &buffer, "11th")

	ParseTest(&buffer, "{{ 12 | ordinal }}", "")
	AssertEqual(t, &buffer, "12th")

	ParseTest(&buffer, "{{ 13 | ordinal }}", "")
	AssertEqual(t, &buffer, "13th")

	ParseTest(&buffer, "{{ 14 | ordinal }}", "")
	AssertEqual(t, &buffer, "14th")

	ParseTest(&buffer, "{{ -1 | ordinal }}", "")
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ . | ordinal }}", uint(14))
	AssertEqual(t, &buffer, "14th")

	ParseTest(&buffer, "{{ . | ordinal }}", false)
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ . | first }}", "The go programming language")
	AssertEqual(t, &buffer, "T")

	ParseTest(&buffer, "{{ . | first }}", "안녕하세요")
	AssertEqual(t, &buffer, "안")

	ParseTest(&buffer, "{{ . | first }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "go")

	ParseTest(&buffer, "{{ . | first }}", [3]string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "go")

	ParseTest(&buffer, "{{ . | first }}", false)
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ . | last }}", "The go programming language")
	AssertEqual(t, &buffer, "e")

	ParseTest(&buffer, "{{ . | last }}", "안녕하세요")
	AssertEqual(t, &buffer, "요")

	ParseTest(&buffer, "{{ . | last }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "ruby")

	ParseTest(&buffer, "{{ . | last }}", false)
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ . | join \" \" }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "go python ruby")

	ParseTest(&buffer, "{{ . | slice 0 2 }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "[go python]")

	ParseTest(&buffer, "{{ . | slice 0 6 }}", "The go programming language")
	AssertEqual(t, &buffer, "The go")

	ParseTest(&buffer, "{{ . | slice 0 2 }}", "안녕하세요")
	AssertEqual(t, &buffer, "안녕")

	ParseTest(&buffer, "{{ . | slice -10 10 }}", "안녕하세요")
	AssertEqual(t, &buffer, "안녕하세요")

	ParseTest(&buffer, "{{ . | slice 0 2 }}", false)
	AssertEqual(t, &buffer, "")

	ParseTest(&buffer, "{{ . | random }}", "T")
	AssertEqual(t, &buffer, "T")

	ParseTest(&buffer, "{{ . | random }}", "안")
	AssertEqual(t, &buffer, "안")

	ParseTest(&buffer, "{{ . | random }}", []string{"go"})
	AssertEqual(t, &buffer, "go")

	ParseTest(&buffer, "{{ . | random }}", [1]string{"go"})
	AssertEqual(t, &buffer, "go")

	ParseTest(&buffer, "{{ . | random }}", false)
	AssertEqual(t, &buffer, "")
}

func TestInject(t *testing.T) {
	var buffer bytes.Buffer

	var originalFuncMap = template.FuncMap{
		// originalFuncMap is made for test purpose.
		// It tests that Inject function does not overwrite the original functions
		// which have same names.
		"length": func(value interface{}) int {
			return -1
		},
		"lower": func(s string) string {
			return "foo"
		},
	}

	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | length }}", "")
	AssertEqual(t, &buffer, "-1")
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | lower }}", "")
	AssertEqual(t, &buffer, "foo")

	Inject(originalFuncMap) // Inject!

	// Check if Inject function does not overwrite the original functions which have same names.
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | length }}", "")
	AssertEqual(t, &buffer, "-1")
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | lower }}", "")
	AssertEqual(t, &buffer, "foo")

	// Now, I can use gtf functions because they are injected into originalFuncMap.
	CustomParseTest(originalFuncMap, &buffer, "{{ . | first }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "go")
	CustomParseTest(originalFuncMap, &buffer, "{{ . | slice 0 6 }}", "The go programming language")
	AssertEqual(t, &buffer, "The go")
	CustomParseTest(originalFuncMap, &buffer, "{{ 13 | ordinal }}", "")
	AssertEqual(t, &buffer, "13th")
}

func TestForceInject(t *testing.T) {
	var buffer bytes.Buffer

	var originalFuncMap = template.FuncMap{
		"length": func(value interface{}) int {
			return -1
		},
		"lower": func(s string) string {
			return "foo"
		},
	}

	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | length }}", "")
	AssertEqual(t, &buffer, "-1")
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | lower }}", "")
	AssertEqual(t, &buffer, "foo")

	ForceInject(originalFuncMap) // ForceInject!

	// Check if ForceInject function overwrites the original functions which have same names.
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | length }}", "")
	AssertEqual(t, &buffer, "27")
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | lower }}", "")
	AssertEqual(t, &buffer, "the go programming language")

	// Now, I can use gtf functions because they are injected into originalFuncMap.
	CustomParseTest(originalFuncMap, &buffer, "{{ . | first }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "go")
	CustomParseTest(originalFuncMap, &buffer, "{{ . | slice 0 6 }}", "The go programming language")
	AssertEqual(t, &buffer, "The go")
	CustomParseTest(originalFuncMap, &buffer, "{{ 13 | ordinal }}", "")
	AssertEqual(t, &buffer, "13th")
}

func TestInjectWithPrefix(t *testing.T) {
	var buffer bytes.Buffer

	var originalFuncMap = template.FuncMap{
		"length": func(value interface{}) int {
			return -1
		},
		"lower": func(s string) string {
			return "foo"
		},
	}

	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | length }}", "")
	AssertEqual(t, &buffer, "-1")
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | lower }}", "")
	AssertEqual(t, &buffer, "foo")

	InjectWithPrefix(originalFuncMap, "gtf_") // InjectWithPrefix! (prefix : gtf_)

	// Check if Inject function does not overwrite the original functions which have same names.
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | length }}", "")
	AssertEqual(t, &buffer, "-1")
	CustomParseTest(originalFuncMap, &buffer, "{{ \"The Go Programming Language\" | lower }}", "")
	AssertEqual(t, &buffer, "foo")

	// Now, I can use gtf functions because they are injected into originalFuncMap.
	CustomParseTest(originalFuncMap, &buffer, "{{ . | gtf_first }}", []string{"go", "python", "ruby"})
	AssertEqual(t, &buffer, "go")
	CustomParseTest(originalFuncMap, &buffer, "{{ . | gtf_slice 0 6 }}", "The go programming language")
	AssertEqual(t, &buffer, "The go")
	CustomParseTest(originalFuncMap, &buffer, "{{ 13 | gtf_ordinal }}", "")
	AssertEqual(t, &buffer, "13th")
}
