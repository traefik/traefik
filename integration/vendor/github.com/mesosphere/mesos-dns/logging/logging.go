package logging

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/golang/glog"
)

var (
	// VerboseFlag enables verbose logging if set to true.
	VerboseFlag bool
	// VeryVerboseFlag enables very verbose logging if set to true.
	VeryVerboseFlag bool
	// Verbose is this package's verbose Logger.
	Verbose *log.Logger
	// VeryVerbose is this package's very verbose Logger.
	VeryVerbose *log.Logger
	// Error is this package's error Logger.
	Error *log.Logger
)

// Counter defines an interface for a monotonically incrementing value.
type Counter interface {
	Inc()
}

// LogCounter implements the Counter interface with a uint64 register.
// It's safe for concurrent use.
type LogCounter struct {
	value uint64
}

// Inc increments the counter by one.
func (lc *LogCounter) Inc() {
	atomic.AddUint64(&lc.value, 1)
}

// String returns a string represention of the counter.
func (lc *LogCounter) String() string {
	return strconv.FormatUint(atomic.LoadUint64(&lc.value), 10)
}

// LogOut holds metrics captured in an instrumented runtime.
type LogOut struct {
	MesosRequests     Counter
	MesosSuccess      Counter
	MesosNXDomain     Counter
	MesosFailed       Counter
	NonMesosRequests  Counter
	NonMesosSuccess   Counter
	NonMesosNXDomain  Counter
	NonMesosFailed    Counter
	NonMesosForwarded Counter
}

// CurLog is the default package level LogOut.
var CurLog = LogOut{
	MesosRequests:     &LogCounter{},
	MesosSuccess:      &LogCounter{},
	MesosNXDomain:     &LogCounter{},
	MesosFailed:       &LogCounter{},
	NonMesosRequests:  &LogCounter{},
	NonMesosSuccess:   &LogCounter{},
	NonMesosNXDomain:  &LogCounter{},
	NonMesosFailed:    &LogCounter{},
	NonMesosForwarded: &LogCounter{},
}

// PrintCurLog prints out the current LogOut and then resets
func PrintCurLog() {
	VeryVerbose.Printf("%+v\n", CurLog)
}

// SetupLogs provides the following logs
// Verbose = optional verbosity
// VeryVerbose = optional verbosity
// Error = stderr
func SetupLogs() {
	// initialize logging flags
	if glog.V(2) {
		VeryVerboseFlag = true
	} else if glog.V(1) {
		VerboseFlag = true
	}

	logopts := log.Ldate | log.Ltime | log.Lshortfile

	if VerboseFlag {
		Verbose = log.New(os.Stdout, "VERBOSE: ", logopts)
		VeryVerbose = log.New(ioutil.Discard, "VERY VERBOSE: ", logopts)
	} else if VeryVerboseFlag {
		Verbose = log.New(os.Stdout, "VERY VERBOSE: ", logopts)
		VeryVerbose = Verbose
	} else {
		Verbose = log.New(ioutil.Discard, "VERBOSE: ", logopts)
		VeryVerbose = log.New(ioutil.Discard, "VERY VERBOSE: ", logopts)
	}

	Error = log.New(os.Stderr, "ERROR: ", logopts)
}
