package plugin

import (
	"os/exec"
	"runtime"
	"testing"

	"github.com/containous/traefik/plugin/proto"
	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
)

func TestBasicPlugin(t *testing.T) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  RemoteHandshake,
		Plugins:          RemotePluginMap,
		Cmd:              exec.Command("sh", "-c", "./test/plugin-test-go-grpc_"+runtime.GOOS+"-"+runtime.GOARCH),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC, plugin.ProtocolGRPC},
	})

	assert.NotNil(t, client)

	rpcClient, err := client.Client()

	assert.Nil(t, err)

	assert.NotNil(t, rpcClient)

	raw, err := rpcClient.Dispense("middleware")

	assert.Nil(t, err)

	assert.NotNil(t, raw)

	remote := raw.(RemotePluginMiddleware)

	assert.NotNil(t, remote)

	resp, err := remote.ServeHTTP(&proto.Request{
		RequestUuid: "blah-blah-blah",
		Request: &proto.HttpRequest{
			Body: []byte("test"),
		},
	})

	assert.Nil(t, err)

	assert.NotNil(t, resp)

	assert.Equal(t, int32(200), resp.Response.StatusCode)
	assert.Equal(t, []byte("test"), resp.Response.Body)

	client.Kill()
}
