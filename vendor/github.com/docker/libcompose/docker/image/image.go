package image

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	"github.com/docker/libcompose/docker/auth"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Exists return whether or not the service image already exists
func Exists(ctx context.Context, clt client.ImageAPIClient, image string) (bool, error) {
	_, err := InspectImage(ctx, clt, image)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// InspectImage inspect the specified image (can be a name, an id or a digest)
// with the specified client.
func InspectImage(ctx context.Context, client client.ImageAPIClient, image string) (types.ImageInspect, error) {
	imageInspect, _, err := client.ImageInspectWithRaw(ctx, image)
	return imageInspect, err
}

// RemoveImage removes the specified image (can be a name, an id or a digest)
// from the daemon store with the specified client.
func RemoveImage(ctx context.Context, client client.ImageAPIClient, image string) error {
	_, err := client.ImageRemove(ctx, image, types.ImageRemoveOptions{})
	return err
}

// PullImage pulls the specified image (can be a name, an id or a digest)
// to the daemon store with the specified client.
func PullImage(ctx context.Context, client client.ImageAPIClient, serviceName string, authLookup auth.Lookup, image string) error {
	fmt.Fprintf(os.Stderr, "Pulling %s (%s)...\n", serviceName, image)
	distributionRef, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		return err
	}

	repoInfo, err := registry.ParseRepositoryInfo(distributionRef)
	if err != nil {
		return err
	}

	authConfig := authLookup.Lookup(repoInfo)

	// Use ConfigFile.SaveToWriter to not re-define encodeAuthToBase64
	encodedAuth, err := encodeAuthToBase64(authConfig)
	if err != nil {
		return err
	}

	options := types.ImagePullOptions{
		RegistryAuth: encodedAuth,
	}
	responseBody, err := client.ImagePull(ctx, distributionRef.String(), options)
	if err != nil {
		logrus.Errorf("Failed to pull image %s: %v", image, err)
		return err
	}
	defer responseBody.Close()

	var writeBuff io.Writer = os.Stderr

	outFd, isTerminalOut := term.GetFdInfo(os.Stderr)

	err = jsonmessage.DisplayJSONMessagesStream(responseBody, writeBuff, outFd, isTerminalOut, nil)
	if err != nil {
		if jerr, ok := err.(*jsonmessage.JSONError); ok {
			// If no error code is set, default to 1
			if jerr.Code == 0 {
				jerr.Code = 1
			}
			fmt.Fprintf(os.Stderr, "%s", writeBuff)
			return fmt.Errorf("Status: %s, Code: %d", jerr.Message, jerr.Code)
		}
	}
	return err
}

// encodeAuthToBase64 serializes the auth configuration as JSON base64 payload
func encodeAuthToBase64(authConfig types.AuthConfig) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}
