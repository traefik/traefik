FROM golang:1.16

ARG USER=$USER
ARG UID=$UID
ARG GID=$GID
RUN useradd -m ${USER} --uid=${UID} && echo "${USER}:" chpasswd
USER ${UID}:${GID}

ARG KUBE_VERSION

RUN go get k8s.io/code-generator@$KUBE_VERSION; exit 0
RUN go get k8s.io/apimachinery@$KUBE_VERSION; exit 0

RUN mkdir -p $GOPATH/src/k8s.io/{code-generator,apimachinery}
RUN cp -R $GOPATH/pkg/mod/k8s.io/code-generator@$KUBE_VERSION $GOPATH/src/k8s.io/code-generator
RUN cp -R $GOPATH/pkg/mod/k8s.io/apimachinery@$KUBE_VERSION $GOPATH/src/k8s.io/apimachinery
RUN chmod +x $GOPATH/src/k8s.io/code-generator/generate-groups.sh

# Install deepcopy-gen executable
RUN go install k8s.io/code-generator/cmd/deepcopy-gen@$KUBE_VERSION; exit 0

WORKDIR $GOPATH/src/k8s.io/code-generator
