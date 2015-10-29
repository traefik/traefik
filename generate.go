/*
Copyright
*/

//go:generate go get github.com/jteeuwen/go-bindata/...
//go:generate rm -vf gen.go
//go:generate go-bindata -o gen.go static/... templates/...

//go:generate mkdir -p vendor/github.com/docker/docker/autogen/dockerversion
//go:generate cp script/dockerversion vendor/github.com/docker/docker/autogen/dockerversion/dockerversion.go

package main
