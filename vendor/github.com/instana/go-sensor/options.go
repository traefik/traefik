package instana

// Options allows the user to configure the to-be-initialized
// sensor
type Options struct {
	Service                     string
	AgentHost                   string
	AgentPort                   int
	MaxBufferedSpans            int
	ForceTransmissionStartingAt int
	LogLevel                    int
}
