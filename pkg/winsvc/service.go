package winsvc

var ChanExit = make(chan int, 1) // Buffered channel to prevent blocking
