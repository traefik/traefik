package appinsights

import "fmt"

type DiagnosticsMessageWriter interface {
	Write(string)
	appendListener(*diagnosticsMessageListener)
}

type diagnosticsMessageWriter struct {
	listeners []chan string
}

type DiagnosticsMessageProcessor func(string)

type DiagnosticsMessageListener interface {
	ProcessMessages(DiagnosticsMessageProcessor)
}

type diagnosticsMessageListener struct {
	channel chan string
}

var diagnosticsWriter *diagnosticsMessageWriter = &diagnosticsMessageWriter{
	listeners: make([]chan string, 0),
}

func NewDiagnosticsMessageListener() DiagnosticsMessageListener {
	listener := &diagnosticsMessageListener{
		channel: make(chan string),
	}

	diagnosticsWriter.appendListener(listener)

	return listener
}

func (writer *diagnosticsMessageWriter) appendListener(listener *diagnosticsMessageListener) {
	writer.listeners = append(writer.listeners, listener.channel)
}

func (writer *diagnosticsMessageWriter) Write(message string) {
	for _, c := range writer.listeners {
		c <- message
	}
}

func (writer *diagnosticsMessageWriter) Printf(message string, args ...interface{}) {
	// Don't bother with Sprintf if nobody is listening
	if writer.hasListeners() {
		writer.Write(fmt.Sprintf(message, args...))
	}
}

func (writer *diagnosticsMessageWriter) hasListeners() bool {
	return len(writer.listeners) > 0
}

func (listener *diagnosticsMessageListener) ProcessMessages(process DiagnosticsMessageProcessor) {
	for {
		message := <-listener.channel
		process(message)
	}
}
