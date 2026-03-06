package tcphealthcheckmode

type TCPHealthCheckMode string

const (
	ModeTcp  TCPHealthCheckMode = "TCP"
	ModeHttp TCPHealthCheckMode = "HTTP"
)
