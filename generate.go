//go:generate mkdir -p autogen
//go:generate rm -vf autogen/genstatic/gen.go
//go:generate mkdir -p static
//go:generate go-bindata -pkg genstatic -nocompress -o autogen/genstatic/gen.go ./static/...
//go:generate go run ./internal/

package main

func main() {}
