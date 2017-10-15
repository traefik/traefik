//go:generate rm -vf autogen/gen.go
//go:generate mkdir -p static
//go:generate go-bindata -pkg autogen -o autogen/gen.go ./static/... ./templates/...

package main

func main() {}
