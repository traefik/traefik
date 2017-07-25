package marathon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadinessCheck(t *testing.T) {
	rc := ReadinessCheck{}
	rc.SetName("readiness").
		SetProtocol("HTTP").
		SetPath("/ready").
		SetPortName("http").
		SetInterval(3 * time.Second).
		SetTimeout(5 * time.Second).
		SetHTTPStatusCodesForReady([]int{200, 201}).
		SetPreserveLastResponse(true)

	if assert.NotNil(t, rc.Name) {
		assert.Equal(t, "readiness", *rc.Name)
	}
	assert.Equal(t, rc.Protocol, "HTTP")
	assert.Equal(t, rc.Path, "/ready")
	assert.Equal(t, rc.PortName, "http")
	assert.Equal(t, rc.IntervalSeconds, 3)
	assert.Equal(t, rc.TimeoutSeconds, 5)
	if assert.NotNil(t, rc.HTTPStatusCodesForReady) {
		assert.Equal(t, *rc.HTTPStatusCodesForReady, []int{200, 201})
	}
	if assert.NotNil(t, rc.PreserveLastResponse) {
		assert.True(t, *rc.PreserveLastResponse)
	}
}
