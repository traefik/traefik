package main

import (
	"fmt"
	"sync"

	"github.com/codegangsta/negroni"
	"github.com/tylerb/graceful"
)

func main() {

	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		n := negroni.New()
		fmt.Println("Launching server on :3000")
		graceful.Run(":3000", 0, n)
		fmt.Println("Terminated server on :3000")
		wg.Done()
	}()
	go func() {
		n := negroni.New()
		fmt.Println("Launching server on :3001")
		graceful.Run(":3001", 0, n)
		fmt.Println("Terminated server on :3001")
		wg.Done()
	}()
	go func() {
		n := negroni.New()
		fmt.Println("Launching server on :3002")
		graceful.Run(":3002", 0, n)
		fmt.Println("Terminated server on :3002")
		wg.Done()
	}()
	fmt.Println("Press ctrl+c. All servers should terminate.")
	wg.Wait()

}
