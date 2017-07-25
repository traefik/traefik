package lib

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_AccountInfo_GetAccountInfo_Error(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusNotAcceptable, `{error}`)
	defer server.Close()

	info, err := client.GetAccountInfo()
	assert.NotNil(t, info)
	if assert.NotNil(t, err) {
		assert.Equal(t, `{error}`, err.Error())
	}
}

func Test_AccountInfo_GetAccountInfo_NoInfo(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `[]`)
	defer server.Close()

	info, err := client.GetAccountInfo()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, info) {
		assert.Equal(t, 0.00, info.Balance)
		assert.Equal(t, 0.00, info.PendingCharges)
		assert.Equal(t, "", info.LastPaymentDate)
		assert.Equal(t, 0.00, info.LastPaymentAmount)
	}
}

func Test_AccountInfo_GetAccountInfo_OK(t *testing.T) {
	server, client := getTestServerAndClient(http.StatusOK, `{
		"balance":-15.97,"pending_charges":"2.34",
		"last_payment_date":"2015-01-29 05:06:27","last_payment_amount":"-5.00"}`)
	defer server.Close()

	info, err := client.GetAccountInfo()
	if err != nil {
		t.Error(err)
	}
	if assert.NotNil(t, info) {
		assert.Equal(t, -15.97, info.Balance)
		assert.Equal(t, 2.34, info.PendingCharges)
		assert.Equal(t, "2015-01-29 05:06:27", info.LastPaymentDate)
		assert.Equal(t, -5.00, info.LastPaymentAmount)
	}
}
