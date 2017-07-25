package zbase32_test

import "fmt"
import "github.com/tv42/zbase32"

func Example() {
	s := zbase32.EncodeToString([]byte{240, 191, 199})
	fmt.Println(s)
	// Output:
	// 6n9hq
}
