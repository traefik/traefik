package main

import ()

type NoInterface struct {
}

func Load() interface{} {
	return &NoInterface{}
}
