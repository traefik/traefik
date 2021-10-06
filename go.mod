module github.com/traefik/traefik

go 1.16

require (
	github.com/ArthurHlt/go-eureka-client v0.0.0-20170403140305-9d0a49cbd39a
	github.com/ArthurHlt/gominlog v0.0.0-20170402142412-72eebf980f46 // indirect
	github.com/Azure/azure-sdk-for-go v40.3.0+incompatible // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/BurntSushi/ty v0.0.0-20140213233908-6add9cd6ad42
	github.com/Masterminds/sprig v2.19.0+incompatible
	github.com/Microsoft/go-winio v0.4.3 // indirect
	github.com/NYTimes/gziphandler v1.0.1
	github.com/Shopify/sarama v1.30.0 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/abbot/go-http-auth v0.0.0-00010101000000-000000000000
	github.com/abronan/valkeyrie v0.2.0
	github.com/apache/thrift v0.12.0 // indirect
	github.com/armon/go-metrics v0.3.8 // indirect
	github.com/armon/go-proxyproto v0.0.0-20170620220930-48572f11356f
	github.com/aws/aws-sdk-go v1.39.0
	github.com/cenk/backoff v2.1.1+incompatible
	github.com/codahale/hdrhistogram v0.9.0 // indirect
	github.com/containous/flaeg v1.4.1
	github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
	github.com/containous/staert v3.1.2+incompatible
	github.com/containous/traefik-extra-service-fabric v1.7.1-0.20210227093100-8dcd57b609a8
	github.com/coreos/go-systemd v0.0.0-20190321100706-95778dfbb74e
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker v1.4.2-0.20171023200535-7848b8beb9d3
	github.com/docker/go-connections v0.3.0
	github.com/docker/leadership v0.0.0-00010101000000-000000000000
	github.com/docker/libcompose v0.4.1-0.20190808084053-143e0f3f1ab9 // indirect
	github.com/donovanhide/eventsource v0.0.0-20170630084216-b8f31a59085e // indirect
	github.com/eapache/channels v1.1.0
	github.com/eknkc/amber v0.0.0-20171010120322-cdade1c07385 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/gambol99/go-marathon v0.7.2-0.20180614232016-99a156b96fb2
	github.com/go-acme/lego/v4 v4.5.3
	github.com/go-check/check v0.0.0-00010101000000-000000000000
	github.com/go-kit/kit v0.9.0
	github.com/golang/protobuf v1.5.2
	github.com/google/go-github v9.0.0+incompatible
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/websocket v1.4.2
	github.com/gravitational/trace v1.1.3 // indirect
	github.com/hashicorp/consul/api v1.9.1
	github.com/hashicorp/go-hclog v0.14.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.0 // indirect
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.1
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/memberlist v0.2.4 // indirect
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab
	github.com/jjcollinge/servicefabric v0.0.2-0.20180125130438-8eebe170fa1b
	github.com/libkermit/compose v0.0.0-20171122111507-c04e39c026ad
	github.com/libkermit/docker v0.0.0-20171122101128-e6674d32b807
	github.com/libkermit/docker-check v0.0.0-20171122104347-1113af38e591
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mesos/mesos-go v0.0.3-0.20150930144802-068d5470506e
	github.com/mesosphere/mesos-dns v0.0.0-00010101000000-000000000000
	github.com/miekg/dns v1.1.43
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.4.1
	github.com/mitchellh/reflectwalk v1.0.1 // indirect
	github.com/mvdan/xurls v1.1.1-0.20170309204242-db96455566f0
	github.com/ogier/pflag v0.0.2-0.20160129220114-45c278ab3607
	github.com/opencontainers/image-spec v1.0.0-rc5.0.20170515205857-f03dbe35d449 // indirect
	github.com/opencontainers/runc v1.0.0-rc3.0.20170425215914-b6b70e534517 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.0.2
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/rancher/go-rancher v0.1.1-0.20171004213057-52e2f4895340
	github.com/rancher/go-rancher-metadata v0.0.0-00010101000000-000000000000
	github.com/ryanuber/go-glob v1.0.0
	github.com/shopspring/decimal v1.1.1-0.20191009025716-f1972eb1d1f5
	github.com/sirupsen/logrus v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/stvp/go-udp-testing v0.0.0-20171104055251-c4434f09ec13
	github.com/thoas/stats v0.0.0-20190104110215-4975baf6a358
	github.com/tinylib/msgp v1.0.2 // indirect
	github.com/tv42/zbase32 v0.0.0-20150911225513-03389da7e0bf // indirect
	github.com/uber/jaeger-client-go v2.15.0+incompatible
	github.com/uber/jaeger-lib v1.5.0
	github.com/unrolled/render v0.0.0-20170109143244-50716a0a8537
	github.com/unrolled/secure v1.0.5
	github.com/urfave/negroni v0.2.1-0.20170426175938-490e6a555d47
	github.com/vdemeester/shakers v0.1.0
	github.com/vulcand/oxy v1.2.0
	go.etcd.io/bbolt v1.3.5 // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf
	google.golang.org/grpc v1.38.0
	gopkg.in/DataDog/dd-trace-go.v1 v1.13.0
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/utils v0.0.0-20210709001253-0e1f9d693477 // indirect
)

replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20180112153951-65b0cdae8d7f
	github.com/docker/docker => github.com/docker/engine v0.0.0-20190725163905-fa8dd90ceb7b
	github.com/docker/leadership => github.com/containous/leadership v0.1.1-0.20180123135645-a2e096d9fe0a
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
	github.com/mesosphere/mesos-dns => github.com/containous/mesos-dns v0.5.3-rc1.0.20160623212649-b47dc4c19f21
	github.com/rancher/go-rancher-metadata => github.com/containous/go-rancher-metadata v0.0.0-20180116133453-e937e8308985
	gopkg.in/fsnotify.v1 => github.com/fsnotify/fsnotify v1.4.2
)
