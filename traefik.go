package main

import (
	fmtlog "log"
	"os"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	if err := traefikCmd.Execute(); err != nil {
		fmtlog.Println(err)
		os.Exit(-1)
	}
	os.Exit(0)
}
