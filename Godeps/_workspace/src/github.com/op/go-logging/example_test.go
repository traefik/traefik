package logging

import "os"

func Example() {
	// This call is for testing purposes and will set the time to unix epoch.
	InitForTesting(DEBUG)

	var log = MustGetLogger("example")

	// For demo purposes, create two backend for os.Stdout.
	//
	// os.Stderr should most likely be used in the real world but then the
	// "Output:" check in this example would not work.
	backend1 := NewLogBackend(os.Stdout, "", 0)
	backend2 := NewLogBackend(os.Stdout, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	var format = MustStringFormatter(
		"%{time:15:04:05.000} %{shortfunc} %{level:.1s} %{message}",
	)
	backend2Formatter := NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend2
	backend2Leveled := AddModuleLevel(backend2Formatter)
	backend2Leveled.SetLevel(ERROR, "")

	// Set the backends to be used and the default level.
	SetBackend(backend1, backend2Leveled)

	log.Debug("debug %s", "arg")
	log.Error("error")

	// Output:
	// debug arg
	// error
	// 00:00:00.000 Example E error
}
