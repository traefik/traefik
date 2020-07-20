module github.com/containous/traefik/v2

go 1.14

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/ExpediaDotCom/haystack-client-go v0.0.0-20190315171017-e7edbdf53a61
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/NYTimes/gziphandler v1.1.1
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/abbot/go-http-auth v0.0.0-00010101000000-000000000000
	github.com/abronan/valkeyrie v0.0.0-20200127174252-ef4277a138cd
	github.com/aws/aws-sdk-go v1.31.12
	github.com/c0va23/go-proxyprotocol v0.9.1
	github.com/cenkalti/backoff/v4 v4.0.0
	github.com/containous/alice v0.0.0-20181107144136-d83ebdd94cbd
	github.com/containous/yaegi v0.8.13
	github.com/coreos/go-systemd v0.0.0-20191104093116-d3cd4ed1dbcf
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/cli v0.0.0-20200221155518-740919cc7fc0
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0
	github.com/docker/go-metrics v0.0.0-20181218153428-b84716841b82 // indirect
	github.com/docker/libcompose v0.0.0-20190805081528-eac9fe1b8b03 // indirect
	github.com/docker/libtrust v0.0.0-20160708172513-aabc10ec26b7 // indirect
	github.com/donovanhide/eventsource v0.0.0-20170630084216-b8f31a59085e // indirect
	github.com/eapache/channels v1.1.0
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/fatih/structs v1.1.0
	github.com/gambol99/go-marathon v0.0.0-20180614232016-99a156b96fb2
	github.com/go-acme/lego/v3 v3.8.0
	github.com/go-check/check v0.0.0-00010101000000-000000000000
	github.com/go-kit/kit v0.9.0
	github.com/golang/protobuf v1.4.2
	github.com/google/go-github/v28 v28.1.1
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/consul/api v1.3.0
	github.com/hashicorp/go-version v1.2.0
	github.com/influxdata/influxdb1-client v0.0.0-20190809212627-fc22c7df067e
	github.com/instana/go-sensor v1.5.1
	github.com/libkermit/compose v0.0.0-20171122111507-c04e39c026ad
	github.com/libkermit/docker v0.0.0-20171122101128-e6674d32b807
	github.com/libkermit/docker-check v0.0.0-20171122104347-1113af38e591
	github.com/mailgun/ttlmap v0.0.0-20170619185759-c1c17f74874f
	github.com/miekg/dns v1.1.27
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.3.2
	github.com/opencontainers/runc v1.0.0-rc10 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.4.5
	github.com/openzipkin/zipkin-go v0.2.2
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v1.5.0
	github.com/prometheus/client_model v0.2.0
	github.com/rancher/go-rancher-metadata v0.0.0-20200311180630-7f4c936a06ac
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/stvp/go-udp-testing v0.0.0-20191102171040-06b61409b154
	github.com/tinylib/msgp v1.0.2 // indirect
	github.com/uber/jaeger-client-go v2.22.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible
	github.com/unrolled/render v1.0.2
	github.com/unrolled/secure v1.0.7
	github.com/vdemeester/shakers v0.1.0
	github.com/vulcand/oxy v1.1.0
	github.com/vulcand/predicate v1.1.0
	go.elastic.co/apm v1.7.0
	go.elastic.co/apm/module/apmot v1.7.0
	golang.org/x/mod v0.3.0
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1
	google.golang.org/grpc v1.30.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.19.0
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/jcmturner/goidentity.v3 v3.0.0 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	k8s.io/code-generator v0.18.2
	knative.dev/networking v0.0.0-20200716201933-30a27fbaff8a
	mvdan.cc/xurls/v2 v2.1.0
)

// Docker v19.03.6
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203

replace (
	github.com/golang/protobuf => github.com/golang/protobuf v1.3.4
	google.golang.org/genproto => google.golang.org/genproto v0.0.0-20200513103714-09dca8ec2884
	google.golang.org/grpc => google.golang.org/grpc v1.27.1
	k8s.io/client-go => k8s.io/client-go v0.18.2
)

// Containous forks
replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
	github.com/gorilla/mux => github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
	github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595
	github.com/mailgun/multibuf => github.com/containous/multibuf v0.0.0-20190809014333-8b6c9a7e6bba
)
