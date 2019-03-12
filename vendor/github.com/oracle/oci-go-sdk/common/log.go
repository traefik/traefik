// Copyright (c) 2016, 2018, Oracle and/or its affiliates. All rights reserved.

package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

//sdkLogger an interface for logging in the SDK
type sdkLogger interface {
	//LogLevel returns the log level of sdkLogger
	LogLevel() int

	//Log logs v with the provided format if the current log level is loglevel
	Log(logLevel int, format string, v ...interface{}) error
}

//noLogging no logging messages
const noLogging = 0

//infoLogging minimal logging messages
const infoLogging = 1

//debugLogging some logging messages
const debugLogging = 2

//verboseLogging all logging messages
const verboseLogging = 3

//defaultSDKLogger the default implementation of the sdkLogger
type defaultSDKLogger struct {
	currentLoggingLevel int
	verboseLogger       *log.Logger
	debugLogger         *log.Logger
	infoLogger          *log.Logger
	nullLogger          *log.Logger
}

//defaultLogger is the defaultLogger in the SDK
var defaultLogger sdkLogger
var loggerLock sync.Mutex

//initializes the SDK defaultLogger as a defaultLogger
func init() {
	l, _ := newSDKLogger()
	setSDKLogger(l)
}

//setSDKLogger sets the logger used by the sdk
func setSDKLogger(logger sdkLogger) {
	loggerLock.Lock()
	defaultLogger = logger
	loggerLock.Unlock()
}

// newSDKLogger creates a defaultSDKLogger
// Debug logging is turned on/off by the presence of the environment variable "OCI_GO_SDK_DEBUG"
// The value of the "OCI_GO_SDK_DEBUG" environment variable controls the logging level.
// "null" outputs no log messages
// "i" or "info" outputs minimal log messages
// "d" or "debug" outputs some logs messages
// "v" or "verbose" outputs all logs messages, including body of requests
func newSDKLogger() (defaultSDKLogger, error) {
	logger := defaultSDKLogger{}

	logger.currentLoggingLevel = noLogging
	logger.verboseLogger = log.New(os.Stderr, "VERBOSE ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logger.debugLogger = log.New(os.Stderr, "DEBUG ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logger.infoLogger = log.New(os.Stderr, "INFO ", log.Ldate|log.Lmicroseconds|log.Lshortfile)
	logger.nullLogger = log.New(ioutil.Discard, "", log.Ldate|log.Lmicroseconds|log.Lshortfile)

	configured, isLogEnabled := os.LookupEnv("OCI_GO_SDK_DEBUG")

	// If env variable not present turn logging of
	if !isLogEnabled {
		logger.currentLoggingLevel = noLogging
	} else {

		switch strings.ToLower(configured) {
		case "null":
			logger.currentLoggingLevel = noLogging
			break
		case "i", "info":
			logger.currentLoggingLevel = infoLogging
			break
		case "d", "debug":
			logger.currentLoggingLevel = debugLogging
			break
		//1 here for backwards compatibility
		case "v", "verbose", "1":
			logger.currentLoggingLevel = verboseLogging
			break
		default:
			logger.currentLoggingLevel = infoLogging
		}
		logger.infoLogger.Println("logger level set to: ", logger.currentLoggingLevel)
	}

	return logger, nil
}

func (l defaultSDKLogger) getLoggerForLevel(logLevel int) *log.Logger {
	if logLevel > l.currentLoggingLevel {
		return l.nullLogger
	}

	switch logLevel {
	case noLogging:
		return l.nullLogger
	case infoLogging:
		return l.infoLogger
	case debugLogging:
		return l.debugLogger
	case verboseLogging:
		return l.verboseLogger
	default:
		return l.nullLogger
	}
}

//LogLevel returns the current debug level
func (l defaultSDKLogger) LogLevel() int {
	return l.currentLoggingLevel
}

func (l defaultSDKLogger) Log(logLevel int, format string, v ...interface{}) error {
	logger := l.getLoggerForLevel(logLevel)
	logger.Output(4, fmt.Sprintf(format, v...))
	return nil
}

//Logln logs v appending a new line at the end
//Deprecated
func Logln(v ...interface{}) {
	defaultLogger.Log(infoLogging, "%v\n", v...)
}

// Logf logs v with the provided format
func Logf(format string, v ...interface{}) {
	defaultLogger.Log(infoLogging, format, v...)
}

// Debugf logs v with the provided format if debug mode is set
func Debugf(format string, v ...interface{}) {
	defaultLogger.Log(debugLogging, format, v...)
}

// Debug  logs v if debug mode is set
func Debug(v ...interface{}) {
	m := fmt.Sprint(v...)
	defaultLogger.Log(debugLogging, "%s", m)
}

// Debugln logs v appending a new line if debug mode is set
func Debugln(v ...interface{}) {
	m := fmt.Sprint(v...)
	defaultLogger.Log(debugLogging, "%s\n", m)
}

// IfDebug executes closure if debug is enabled
func IfDebug(fn func()) {
	if defaultLogger.LogLevel() >= debugLogging {
		fn()
	}
}
