package cookie

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetName(t *testing.T) {
	testCases := []struct {
		desc               string
		cookieName         string
		backendName        string
		expectedCookieName string
	}{
		{
			desc:               "with backend name, without cookie name",
			cookieName:         "",
			backendName:        "/my/BACKEND-v1.0~rc1",
			expectedCookieName: "_5f7bc",
		},
		{
			desc:               "without backend name, with cookie name",
			cookieName:         "/my/BACKEND-v1.0~rc1",
			backendName:        "",
			expectedCookieName: "_my_BACKEND-v1.0~rc1",
		},
		{
			desc:               "with backend name, with cookie name",
			cookieName:         "containous",
			backendName:        "treafik",
			expectedCookieName: "containous",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cookieName := GetName(test.cookieName, test.backendName)

			assert.Equal(t, test.expectedCookieName, cookieName)
		})
	}
}

func Test_sanitizeName(t *testing.T) {
	testCases := []struct {
		desc         string
		srcName      string
		expectedName string
	}{
		{
			desc:         "with /",
			srcName:      "/my/BACKEND-v1.0~rc1",
			expectedName: "_my_BACKEND-v1.0~rc1",
		},
		{
			desc:         "some chars",
			srcName:      "!#$%&'()*+-./:<=>?@[]^_`{|}~",
			expectedName: "!#$%&'__*+-._________^_`_|_~",
		},
	}

	for _, test := range testCases {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			cookieName := sanitizeName(test.srcName)

			assert.Equal(t, test.expectedName, cookieName, "Cookie name")
		})
	}
}

func TestGenerateName(t *testing.T) {
	cookieName := GenerateName("containous")

	assert.Len(t, "_8a7bc", 6)
	assert.Equal(t, "_8a7bc", cookieName)
}
