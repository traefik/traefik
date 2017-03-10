package docker

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"golang.org/x/net/context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/reference"
	"github.com/docker/docker/registry"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

func removeImage(ctx context.Context, client client.APIClient, image string) error {
	_, err := client.ImageRemove(ctx, image, types.ImageRemoveOptions{})
	return err
}

func pullImage(ctx context.Context, client client.APIClient, service *Service, image string) error {
	fmt.Fprintf(os.Stderr, "Pulling %s (%s)...\n", service.name, image)
	distributionRef, err := reference.ParseNamed(image)
	if err != nil {
		return err
	}

	repoInfo, err := registry.ParseRepositoryInfo(distributionRef)
	if err != nil {
		return err
	}

	authConfig := service.authLookup.Lookup(repoInfo)

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
