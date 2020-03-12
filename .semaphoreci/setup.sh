# For personnal CI
# mv /home/runner/workspace/src/github.com/<username>/ /home/runner/workspace/src/github.com/containous/
# cd /home/runner/workspace/src/github.com/containous/traefik/
for s in apache2 cassandra elasticsearch memcached mysql mongod postgresql sphinxsearch rethinkdb rabbitmq-server redis-server; do sudo service $s stop; done
sudo swapoff -a
sudo dd if=/dev/zero of=/swapfile bs=1M count=3072
sudo mkswap /swapfile
sudo swapon /swapfile
sudo rm -rf /home/runner/.rbenv
sudo rm -rf /usr/local/golang/{1.4.3,1.5.4,1.6.4,1.7.6,1.8.6,1.9.7,1.10.3,1.11}
#export DOCKER_VERSION=18.06.3
source .semaphoreci/vars
if [ -z "${PULL_REQUEST_NUMBER}" ]; then SHOULD_TEST="-*-"; else TEMP_STORAGE=$(curl --silent https://patch-diff.githubusercontent.com/raw/containous/traefik/pull/${PULL_REQUEST_NUMBER}.diff | patch --dry-run -p1 -R || true); fi
echo ${SHOULD_TEST}
if [ -n "$TEMP_STORAGE" ]; then SHOULD_TEST=$(echo "$TEMP_STORAGE" | grep -Ev '(.md|.yaml|.yml)' || :); fi
echo ${TEMP_STORAGE}
echo ${SHOULD_TEST}
#if [ -n "$SHOULD_TEST" ]; then sudo -E apt-get -yq update; fi
#if [ -n "$SHOULD_TEST" ]; then sudo -E apt-get -yq --no-install-suggests --no-install-recommends --force-yes install docker-ce=${DOCKER_VERSION}*; fi
if [ -n "$SHOULD_TEST" ]; then docker version; fi
export GO_VERSION=1.13
if [ -f "./go.mod" ]; then GO_VERSION="$(grep '^go .*' go.mod | awk '{print $2}')"; export GO_VERSION; fi
#if [ "${GO_VERSION}" == '1.14' ]; then export GO_VERSION=1.14rc2; fi
echo "Selected Go version: ${GO_VERSION}"

if [ -f "./.semaphoreci/golang.sh" ]; then ./.semaphoreci/golang.sh; fi
if [ -f "./.semaphoreci/golang.sh" ]; then export GOROOT="/usr/local/golang/${GO_VERSION}/go"; fi
if [ -f "./.semaphoreci/golang.sh" ]; then export GOTOOLDIR="/usr/local/golang/${GO_VERSION}/go/pkg/tool/linux_amd64"; fi
go version

if [ -f "./go.mod" ]; then export GO111MODULE=on; fi
if [ -f "./go.mod" ]; then export GOPROXY=https://proxy.golang.org; fi
if [ -f "./go.mod" ]; then go mod download; fi

df
