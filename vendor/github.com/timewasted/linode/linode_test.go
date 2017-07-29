package linode

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinode_RequestSuccess(t *testing.T) {
	mockResponses := MockResponseMap{
		"test.echo": MockResponse{
			Response: map[string]string{"ping": "pong"},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	var response map[string]string
	l := New("testing")
	l.SetEndpoint(mockSrv.URL)
	_, err := l.Request("test.echo", Parameters{
		"ping": "pong",
	}, &response)
	assert.NoError(t, err)
	assert.Equal(t, mockResponses["test.echo"].Response, response)
}

func TestLinode_RequestSuccessWithoutResponse(t *testing.T) {
	mockResponses := MockResponseMap{
		"test.echo": MockResponse{},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	l := New("testing")
	l.SetEndpoint(mockSrv.URL)
	_, err := l.Request("test.echo", Parameters{
		"ping": "pong",
	}, nil)
	assert.NoError(t, err)
}

func TestLinode_RequestHttpError(t *testing.T) {
	mockResponses := MockResponseMap{
		"test.echo": MockResponse{StatusCode: http.StatusInternalServerError},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	l := New("testing")
	l.SetEndpoint(mockSrv.URL)
	_, err := l.Request("test.echo", Parameters{
		"ping": "pong",
	}, nil)
	assert.EqualError(t, err, "Expected status code 200, received 500")
}

func TestLinode_RequestApiError(t *testing.T) {
	mockResponses := MockResponseMap{
		"test.echo": MockResponse{
			Response: nil,
			Errors: []ResponseError{
				ResponseError{
					Code:    1234,
					Message: "Failed to ping pong",
				},
			},
		},
	}
	mockSrv := newMockServer(t, mockResponses)
	defer mockSrv.Close()

	var response map[string]string
	l := New("testing")
	l.SetEndpoint(mockSrv.URL)
	_, err := l.Request("test.echo", Parameters{
		"ping": "pong",
	}, &response)
	assert.EqualError(t, err, "Failed to ping pong")
}
