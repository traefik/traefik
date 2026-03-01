//go:build tools

package tools

// The following dependencies are required by the Knative conformance tests.
// They allow to download the test_images when calling "go mod vendor".
import (
	_ "knative.dev/networking/test/test_images/grpc-ping"
	_ "knative.dev/networking/test/test_images/httpproxy"
	_ "knative.dev/networking/test/test_images/retry"
	_ "knative.dev/networking/test/test_images/runtime"
	_ "knative.dev/networking/test/test_images/timeout"
	_ "knative.dev/networking/test/test_images/wsserver"
)
