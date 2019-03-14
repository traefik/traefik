//go:generate rm -vf autogen/genstatic/gen.go
//go:generate mkdir -p static
//go:generate go-bindata -pkg genstatic -nocompress -o autogen/genstatic/gen.go ./static/...

package main

func main() {}
