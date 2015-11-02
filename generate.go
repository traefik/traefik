/*
Copyright
*/

//go:generate rm -vf autogen/gen.go
//go:generate go-bindata -pkg autogen -o autogen/gen.go ./static/... ./templates/...

//go:generate mkdir -p vendor/github.com/docker/docker/autogen/dockerversion
//go:generate cp script/dockerversion vendor/github.com/docker/docker/autogen/dockerversion/dockerversion.go

package main
