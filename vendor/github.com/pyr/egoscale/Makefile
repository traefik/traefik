VERSION=0.3.0-snapshot
PREFIX?=/usr/local
GOPATH=$(PWD)/build:$(PWD)
PROGRAM=exo
GO=env GOPATH=$(GOPATH) go
RM?=rm -f
LN=ln -s
MAIN=exo.go
SRCS=		src/egoscale/types.go		 \
			src/egoscale/error.go		 \
			src/egoscale/topology.go \
			src/egoscale/groups.go   \
			src/egoscale/vm.go       \
			src/egoscale/dns.go       \
			src/egoscale/request.go  \
			src/egoscale/async.go  \
			src/egoscale/keypair.go  \
			src/egoscale/ip.go  \
			src/egoscale/init.go

all: $(PROGRAM)

$(PROGRAM): $(MAIN) $(SRCS)
				$(GO) build egoscale
				$(GO) build -o $(PROGRAM) $(MAIN)

clean:
				$(RM) $(PROGRAM)
				$(GO) clean
