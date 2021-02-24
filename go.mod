module github.com/traefik/traefik

go 1.16

require (
	github.com/ArthurHlt/go-eureka-client v0.0.0-20170403140305-9d0a49cbd39a
	github.com/ArthurHlt/gominlog v0.0.0-20170402142412-72eebf980f46 // indirect
	github.com/BurntSushi/toml v0.3.1
	github.com/BurntSushi/ty v0.0.0-20140213233908-6add9cd6ad42
	github.com/Masterminds/sprig v2.19.0+incompatible
	github.com/Microsoft/go-winio v0.4.2 // indirect
	github.com/NYTimes/gziphandler v1.0.1-0.20180125165240-289a3b81f5ae
	github.com/PuerkitoBio/purell v1.0.0 // indirect
	github.com/PuerkitoBio/urlesc v0.0.0-20160726150825-5bd2802263f2 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/abbot/go-http-auth v0.0.0-00010101000000-000000000000
	github.com/abronan/valkeyrie v0.0.0-20171113095143-063d875e3c5f
	github.com/armon/go-metrics v0.3.0 // indirect
	github.com/armon/go-proxyproto v0.0.0-20170620220930-48572f11356f
	github.com/aws/aws-sdk-go v1.23.0
	github.com/cenk/backoff v2.1.1+incompatible
	github.com/codahale/hdrhistogram v0.9.0 // indirect
	github.com/containous/flaeg v1.4.1
	github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
	github.com/containous/staert v3.1.2+incompatible
	github.com/containous/traefik-extra-service-fabric v1.7.1-0.20210227093100-8dcd57b609a8
	github.com/coreos/bbolt v1.3.1-coreos.5 // indirect
	github.com/coreos/etcd v3.3.5+incompatible // indirect
	github.com/coreos/go-semver v0.2.0 // indirect
	github.com/coreos/go-systemd v0.0.0-20161114122254-48702e0da86b
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/docker/docker v1.4.2-0.20171023200535-7848b8beb9d3
	github.com/docker/go-connections v0.3.0
	github.com/docker/leadership v0.0.0-00010101000000-000000000000
	github.com/docker/libcompose v0.4.1-0.20190808084053-143e0f3f1ab9 // indirect
	github.com/donovanhide/eventsource v0.0.0-20170630084216-b8f31a59085e // indirect
	github.com/eapache/channels v1.1.0
	github.com/eknkc/amber v0.0.0-20171010120322-cdade1c07385 // indirect
	github.com/elazarl/go-bindata-assetfs v1.0.0
	github.com/emicklei/go-restful v1.1.4-0.20160814184150-89ef8af493ab // indirect
	github.com/fatih/color v1.5.1-0.20170523202404-62e9147c64a1 // indirect
	github.com/gambol99/go-marathon v0.7.2-0.20180614232016-99a156b96fb2
	github.com/go-acme/lego/v3 v3.0.1
	github.com/go-check/check v0.0.0-00010101000000-000000000000
	github.com/go-kit/kit v0.8.0
	github.com/go-openapi/jsonpointer v0.0.0-20160704185906-46af16f9f7b1 // indirect
	github.com/go-openapi/jsonreference v0.0.0-20160704190145-13c6e3589ad9 // indirect
	github.com/go-openapi/spec v0.0.0-20160808142527-6aced65f8501 // indirect
	github.com/go-openapi/swag v0.0.0-20160704191624-1d0bd113de87 // indirect
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e // indirect
	github.com/golang/protobuf v1.3.2
	github.com/google/go-github v9.0.0+incompatible
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.1.0 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/gravitational/trace v1.1.3 // indirect
	github.com/gregjones/httpcache v0.0.0-20171119193500-2bcd89a1743f // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/hashicorp/consul v1.0.6
	github.com/hashicorp/go-msgpack v1.1.5 // indirect
	github.com/hashicorp/go-rootcerts v0.0.0-20160503143440-6bb64b370b90 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v0.0.0-20170202080759-03c5bf6be031
	github.com/hashicorp/memberlist v0.1.5 // indirect
	github.com/hashicorp/serf v0.8.2-0.20170308193951-19f2c401e122 // indirect
	github.com/influxdata/influxdb v1.3.7
	github.com/jjcollinge/servicefabric v0.0.2-0.20180125130438-8eebe170fa1b
	github.com/juju/ratelimit v1.0.1 // indirect
	github.com/libkermit/compose v0.0.0-20171122111507-c04e39c026ad
	github.com/libkermit/docker v0.0.0-20171122101128-e6674d32b807
	github.com/libkermit/docker-check v0.0.0-20171122104347-1113af38e591
	github.com/mailru/easyjson v0.0.0-20160728113105-d5b7844b561a // indirect
	github.com/mattn/go-colorable v0.0.8-0.20170210172801-5411d3eea597 // indirect
	github.com/mesos/mesos-go v0.0.3-0.20150930144802-068d5470506e
	github.com/mesosphere/mesos-dns v0.0.0-00010101000000-000000000000
	github.com/miekg/dns v1.1.26
	github.com/mitchellh/copystructure v0.0.0-20170525013902-d23ffcb85de3
	github.com/mitchellh/go-testing-interface v1.14.1 // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mitchellh/reflectwalk v0.0.0-20170726202117-63d60e9d0dbc // indirect
	github.com/mvdan/xurls v1.1.1-0.20170309204242-db96455566f0
	github.com/ogier/pflag v0.0.2-0.20160129220114-45c278ab3607
	github.com/opencontainers/image-spec v1.0.0-rc5.0.20170515205857-f03dbe35d449 // indirect
	github.com/opencontainers/runc v1.0.0-rc3.0.20170425215914-b6b70e534517 // indirect
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.0.2
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_golang v1.1.0
	github.com/prometheus/client_model v0.0.0-20190129233127-fd36f4220a90
	github.com/rancher/go-rancher v0.1.1-0.20171004213057-52e2f4895340
	github.com/rancher/go-rancher-metadata v0.0.0-00010101000000-000000000000
	github.com/ryanuber/go-glob v1.0.0
	github.com/samuel/go-zookeeper v0.0.0-20161028232340-1d7be4effb13 // indirect
	github.com/shopspring/decimal v1.1.1-0.20191009025716-f1972eb1d1f5
	github.com/sirupsen/logrus v1.4.2
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/spf13/pflag v0.0.0-20160427162146-cb88ea77998c // indirect
	github.com/stretchr/testify v1.5.1
	github.com/stvp/go-udp-testing v0.0.0-20171104055251-c4434f09ec13
	github.com/thoas/stats v0.0.0-20190104110215-4975baf6a358
	github.com/tinylib/msgp v1.0.2 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/tv42/zbase32 v0.0.0-20150911225513-03389da7e0bf // indirect
	github.com/uber/jaeger-client-go v2.15.0+incompatible
	github.com/uber/jaeger-lib v1.5.0
	github.com/ugorji/go v1.1.1 // indirect
	github.com/unrolled/render v0.0.0-20170109143244-50716a0a8537
	github.com/unrolled/secure v1.0.5
	github.com/urfave/negroni v0.2.1-0.20170426175938-490e6a555d47
	github.com/vdemeester/shakers v0.1.0
	github.com/vulcand/oxy v1.2.0
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	golang.org/x/net v0.0.0-20190923162816-aa69164e4478
	golang.org/x/sys v0.0.0-20191115151921-52ab43148777 // indirect
	google.golang.org/grpc v1.22.1
	gopkg.in/DataDog/dd-trace-go.v1 v1.13.0
	gopkg.in/fsnotify.v1 v1.4.7
	gopkg.in/inf.v0 v0.9.0 // indirect
	gopkg.in/yaml.v2 v2.2.5
	k8s.io/api v0.0.0-20171214033149-af4bc157c3a2
	k8s.io/apimachinery v0.0.0-20171207040834-180eddb345a5
	k8s.io/client-go v6.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20180201014056-275e2ce91dec // indirect
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
