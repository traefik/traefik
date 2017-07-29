package stack_test

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/go-stack/stack"
)

const importPath = "github.com/go-stack/stack"

type testType struct{}

func (tt testType) testMethod() (c stack.Call, pc uintptr, file string, line int, ok bool) {
	c = stack.Caller(0)
	pc, file, line, ok = runtime.Caller(0)
	line--
	return
}

func TestCallFormat(t *testing.T) {
	t.Parallel()

	c := stack.Caller(0)
	pc, file, line, ok := runtime.Caller(0)
	line--
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	relFile := path.Join(importPath, filepath.Base(file))

	c2, pc2, file2, line2, ok2 := testType{}.testMethod()
	if !ok2 {
		t.Fatal("runtime.Caller(0) failed")
	}
	relFile2 := path.Join(importPath, filepath.Base(file2))

	data := []struct {
		c    stack.Call
		desc string
		fmt  string
		out  string
	}{
		{stack.Call{}, "error", "%s", "%!s(NOFUNC)"},

		{c, "func", "%s", path.Base(file)},
		{c, "func", "%+s", relFile},
		{c, "func", "%#s", file},
		{c, "func", "%d", fmt.Sprint(line)},
		{c, "func", "%n", "TestCallFormat"},
		{c, "func", "%+n", runtime.FuncForPC(pc - 1).Name()},
		{c, "func", "%v", fmt.Sprint(path.Base(file), ":", line)},
		{c, "func", "%+v", fmt.Sprint(relFile, ":", line)},
		{c, "func", "%#v", fmt.Sprint(file, ":", line)},

		{c2, "meth", "%s", path.Base(file2)},
		{c2, "meth", "%+s", relFile2},
		{c2, "meth", "%#s", file2},
		{c2, "meth", "%d", fmt.Sprint(line2)},
		{c2, "meth", "%n", "testType.testMethod"},
		{c2, "meth", "%+n", runtime.FuncForPC(pc2).Name()},
		{c2, "meth", "%v", fmt.Sprint(path.Base(file2), ":", line2)},
		{c2, "meth", "%+v", fmt.Sprint(relFile2, ":", line2)},
		{c2, "meth", "%#v", fmt.Sprint(file2, ":", line2)},
	}

	for _, d := range data {
		got := fmt.Sprintf(d.fmt, d.c)
		if got != d.out {
			t.Errorf("fmt.Sprintf(%q, Call(%s)) = %s, want %s", d.fmt, d.desc, got, d.out)
		}
	}
}

func TestCallString(t *testing.T) {
	t.Parallel()

	c := stack.Caller(0)
	_, file, line, ok := runtime.Caller(0)
	line--
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	c2, _, file2, line2, ok2 := testType{}.testMethod()
	if !ok2 {
		t.Fatal("runtime.Caller(0) failed")
	}

	data := []struct {
		c    stack.Call
		desc string
		out  string
	}{
		{stack.Call{}, "error", "%!v(NOFUNC)"},
		{c, "func", fmt.Sprint(path.Base(file), ":", line)},
		{c2, "meth", fmt.Sprint(path.Base(file2), ":", line2)},
	}

	for _, d := range data {
		got := d.c.String()
		if got != d.out {
			t.Errorf("got %s, want %s", got, d.out)
		}
	}
}

func TestCallMarshalText(t *testing.T) {
	t.Parallel()

	c := stack.Caller(0)
	_, file, line, ok := runtime.Caller(0)
	line--
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}

	c2, _, file2, line2, ok2 := testType{}.testMethod()
	if !ok2 {
		t.Fatal("runtime.Caller(0) failed")
	}

	data := []struct {
		c    stack.Call
		desc string
		out  []byte
		err  error
	}{
		{stack.Call{}, "error", nil, stack.ErrNoFunc},
		{c, "func", []byte(fmt.Sprint(path.Base(file), ":", line)), nil},
		{c2, "meth", []byte(fmt.Sprint(path.Base(file2), ":", line2)), nil},
	}

	for _, d := range data {
		text, err := d.c.MarshalText()
		if got, want := err, d.err; got != want {
			t.Errorf("%s: got err %v, want err %v", d.desc, got, want)
		}
		if got, want := text, d.out; !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %s, want %s", d.desc, got, want)
		}
	}
}

func TestCallStackString(t *testing.T) {
	cs, line0 := getTrace(t)
	_, file, line1, ok := runtime.Caller(0)
	line1--
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	file = path.Base(file)
	if got, want := cs.String(), fmt.Sprintf("[%s:%d %s:%d]", file, line0, file, line1); got != want {
		t.Errorf("\n got %v\nwant %v", got, want)
	}
}

func TestCallStackMarshalText(t *testing.T) {
	cs, line0 := getTrace(t)
	_, file, line1, ok := runtime.Caller(0)
	line1--
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	file = path.Base(file)
	text, _ := cs.MarshalText()
	if got, want := text, []byte(fmt.Sprintf("[%s:%d %s:%d]", file, line0, file, line1)); !reflect.DeepEqual(got, want) {
		t.Errorf("\n got %v\nwant %v", got, want)
	}
}
func getTrace(t *testing.T) (stack.CallStack, int) {
	cs := stack.Trace().TrimRuntime()
	_, _, line, ok := runtime.Caller(0)
	line--
	if !ok {
		t.Fatal("runtime.Caller(0) failed")
	}
	return cs, line
}

func TestTrimAbove(t *testing.T) {
	trace := trimAbove()
	if got, want := len(trace), 2; got != want {
		t.Errorf("got len(trace) == %v, want %v, trace: %n", got, want, trace)
	}
	if got, want := fmt.Sprintf("%n", trace[1]), "TestTrimAbove"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func trimAbove() stack.CallStack {
	call := stack.Caller(1)
	trace := stack.Trace()
	return trace.TrimAbove(call)
}

func TestTrimBelow(t *testing.T) {
	trace := trimBelow()
	if got, want := fmt.Sprintf("%n", trace[0]), "TestTrimBelow"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func trimBelow() stack.CallStack {
	call := stack.Caller(1)
	trace := stack.Trace()
	return trace.TrimBelow(call)
}

func TestTrimRuntime(t *testing.T) {
	trace := stack.Trace().TrimRuntime()
	if got, want := len(trace), 1; got != want {
		t.Errorf("got len(trace) == %v, want %v, goroot: %q, trace: %#v", got, want, runtime.GOROOT(), trace)
	}
}

func BenchmarkCallVFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprint(ioutil.Discard, c)
	}
}

func BenchmarkCallPlusVFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%+v", c)
	}
}

func BenchmarkCallSharpVFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%#v", c)
	}
}

func BenchmarkCallSFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%s", c)
	}
}

func BenchmarkCallPlusSFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%+s", c)
	}
}

func BenchmarkCallSharpSFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%#s", c)
	}
}

func BenchmarkCallDFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%d", c)
	}
}

func BenchmarkCallNFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%n", c)
	}
}

func BenchmarkCallPlusNFmt(b *testing.B) {
	c := stack.Caller(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fmt.Fprintf(ioutil.Discard, "%+n", c)
	}
}

func BenchmarkCaller(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stack.Caller(0)
	}
}

func BenchmarkTrace(b *testing.B) {
	for i := 0; i < b.N; i++ {
		stack.Trace()
	}
}

func deepStack(depth int, b *testing.B) stack.CallStack {
	if depth > 0 {
		return deepStack(depth-1, b)
	}
	b.StartTimer()
	s := stack.Trace()
	return s
}

func BenchmarkTrace10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		deepStack(10, b)
	}
}

func BenchmarkTrace50(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		deepStack(50, b)
	}
}

func BenchmarkTrace100(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		deepStack(100, b)
	}
}

////////////////
// Benchmark functions followed by formatting
////////////////

func BenchmarkCallerAndVFmt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Fprint(ioutil.Discard, stack.Caller(0))
	}
}

func BenchmarkTraceAndVFmt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Fprint(ioutil.Discard, stack.Trace())
	}
}

func BenchmarkTrace10AndVFmt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		fmt.Fprint(ioutil.Discard, deepStack(10, b))
	}
}

////////////////
// Baseline against package runtime.
////////////////

func BenchmarkRuntimeCaller(b *testing.B) {
	for i := 0; i < b.N; i++ {
		runtime.Caller(0)
	}
}

func BenchmarkRuntimeCallerAndFmt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, file, line, _ := runtime.Caller(0)
		const sep = "/"
		if i := strings.LastIndex(file, sep); i != -1 {
			file = file[i+len(sep):]
		}
		fmt.Fprint(ioutil.Discard, file, ":", line)
	}
}

func BenchmarkFuncForPC(b *testing.B) {
	pc, _, _, _ := runtime.Caller(0)
	pc--
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.FuncForPC(pc)
	}
}

func BenchmarkFuncFileLine(b *testing.B) {
	pc, _, _, _ := runtime.Caller(0)
	pc--
	fn := runtime.FuncForPC(pc)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fn.FileLine(pc)
	}
}
