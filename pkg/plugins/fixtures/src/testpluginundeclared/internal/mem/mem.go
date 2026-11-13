package mem

import "unsafe"

func WordSize() uintptr {
	return unsafe.Sizeof(uintptr(0))
}
