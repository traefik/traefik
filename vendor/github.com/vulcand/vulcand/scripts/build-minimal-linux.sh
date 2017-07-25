docker run --rm \
 -v /var/run/docker.sock:/var/run/docker.sock \
 -v $(which docker):$(which docker) \
 -v $(pwd):$(pwd) \
 -e HOST_PROJECT_PATH=$(pwd) \
 -e HOST_GOPATH=${GOPATH} \
 -e DOCKER_IMAGE_NAME=${DOCKER_IMAGE_NAME-mailgun/vulcand} \
 golang:1.4-onbuild bash $(pwd)/scripts/static-compile-docker.sh
