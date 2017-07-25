package integration

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	check "gopkg.in/check.v1"
)

type testCondition func() bool

type testRequirement struct {
	Condition   testCondition
	SkipMessage string
}

// List test requirements
var (
	IsWindows = testRequirement{
		func() bool { return runtime.GOOS == "windows" },
		"Test requires a Windows daemon",
	}
	IsLinux = testRequirement{
		func() bool { return runtime.GOOS == "linux" },
		"Test requires a Linux daemon",
	}
	Network = testRequirement{
		func() bool {
			// Set a timeout on the GET at 15s
			var timeout = 15 * time.Second
			var url = "https://hub.docker.com"

			client := http.Client{
				Timeout: timeout,
			}

			resp, err := client.Get(url)
			if err != nil && strings.Contains(err.Error(), "use of closed network connection") {
				panic(fmt.Sprintf("Timeout for GET request on %s", url))
			}
			if resp != nil {
				resp.Body.Close()
			}
			return err == nil
		},
		"Test requires network availability, environment variable set to none to run in a non-network enabled mode.",
	}
)

func not(r testRequirement) testRequirement {
	return testRequirement{
		func() bool {
			return !r.Condition()
		},
		fmt.Sprintf("Not(%s)", r.SkipMessage),
	}
}

func DaemonVersionIs(version string) testRequirement {
	return testRequirement{
		func() bool {
			return strings.Contains(os.Getenv("DOCKER_DAEMON_VERSION"), version)
		},
		"Test requires the daemon version to be " + version,
	}
}

// testRequires checks if the environment satisfies the requirements
// for the test to run or skips the tests.
func testRequires(c *check.C, requirements ...testRequirement) {
	for _, r := range requirements {
		if !r.Condition() {
			c.Skip(r.SkipMessage)
		}
	}
}
