/*
Copyright
*/

//go:generate go get github.com/jteeuwen/go-bindata/...
//go:generate go get github.com/elazarl/go-bindata-assetfs/...
//go:generate rm -vf gen.go
//go:generate go-bindata -o gen.go static/... templates/... providerTemplates/...

package main