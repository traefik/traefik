package marathon

import (
	"errors"

	"github.com/gambol99/go-marathon"
	"github.com/stretchr/testify/mock"
	"github.com/traefik/traefik/v2/pkg/provider/marathon/mocks"
)

type fakeClient struct {
	mocks.Marathon
}

func newFakeClient(applicationsError bool, applications marathon.Applications) *fakeClient {
	// create an instance of our test object
	fakeClient := new(fakeClient)
	if applicationsError {
		fakeClient.On("Applications", mock.Anything).Return(nil, errors.New("fake Marathon server error"))
	} else {
		fakeClient.On("Applications", mock.Anything).Return(&applications, nil)
	}
	return fakeClient
}
