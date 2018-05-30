package hostresolver

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestRules_CNAMEFlatten(t *testing.T) {
	hostResolver := &HostResolver{CacheDuration: 10 * time.Minute, ResolvConfig: "/etc/resolv.conf", ResolvDepth: 5}
	host := [4]string{"www.github.com", "github.com", "www.icann.org", "icann.org"}
	reqHost, flatHost := hostResolver.CNAMEFlatten(host[0])
	assert.Equal(t, host[0], reqHost)
	assert.NotEqual(t, host[0], flatHost)
	reqHost, flatHost = hostResolver.CNAMEFlatten(host[1])
	assert.Equal(t, host[1], reqHost)
	assert.Equal(t, host[1], flatHost)
	reqHost, flatHost = hostResolver.CNAMEFlatten(host[2])
	assert.Equal(t, host[2], reqHost)
	assert.NotEqual(t, host[2], flatHost)
	reqHost, flatHost = hostResolver.CNAMEFlatten(host[3])
	assert.Equal(t, host[3], reqHost)
	assert.Equal(t, host[3], flatHost)
}
