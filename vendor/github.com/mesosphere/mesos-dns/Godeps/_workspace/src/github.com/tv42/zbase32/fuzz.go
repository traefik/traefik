// +build gofuzz

package zbase32

func Fuzz(data []byte) int {
	if _, err := DecodeString(string(data)); err != nil {
		return 0
	}
	return 1
}
