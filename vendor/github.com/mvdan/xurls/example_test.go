// Copyright (c) 2015, Daniel Martí <mvdan@mvdan.cc>
// See LICENSE for licensing information

package xurls_test

import (
	"fmt"

	"github.com/mvdan/xurls"
)

func Example() {
	fmt.Println(xurls.Relaxed.FindString("Do gophers live in http://golang.org?"))
	fmt.Println(xurls.Relaxed.FindAllString("foo.com is http://foo.com/.", -1))
	// Output:
	// http://golang.org
	// [foo.com http://foo.com/]
}
