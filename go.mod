module github.com/traefik/traefik/v3

go 1.24.0

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/Masterminds/sprig/v3 v3.2.3
	github.com/abbot/go-http-auth v0.0.0-00010101000000-000000000000 // No tag on the repo.
	github.com/andybalholm/brotli v1.1.1
	github.com/armon/go-socks5 v0.0.0-20160902184237-e75332964ef5
	github.com/aws/aws-sdk-go-v2 v1.41.0
	github.com/aws/aws-sdk-go-v2/config v1.32.6
	github.com/aws/aws-sdk-go-v2/credentials v1.19.6
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.203.1
	github.com/aws/aws-sdk-go-v2/service/ecs v1.53.15
	github.com/aws/aws-sdk-go-v2/service/ssm v1.56.13
	github.com/aws/smithy-go v1.24.0
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/containous/alice v0.0.0-20181107144136-d83ebdd94cbd // No tag on the repo.
	github.com/coreos/go-systemd/v22 v22.5.0
	github.com/docker/cli v28.3.3+incompatible
	github.com/docker/docker v28.3.3+incompatible
	github.com/docker/go-connections v0.5.0
	github.com/fatih/structs v1.1.0
	github.com/fsnotify/fsnotify v1.9.0
	github.com/go-acme/lego/v4 v4.31.0
	github.com/go-kit/kit v0.13.0
	github.com/go-kit/log v0.2.1
	github.com/golang/protobuf v1.5.4
	github.com/google/go-github/v28 v28.1.1
	github.com/gorilla/mux v1.8.1
	github.com/gorilla/websocket v1.5.4-0.20250319132907-e064f32e3674
	github.com/hashicorp/consul/api v1.26.1
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-retryablehttp v0.7.8
	github.com/hashicorp/go-version v1.8.0
	github.com/hashicorp/nomad/api v0.0.0-20231213195942-64e3dca9274b // No tag on the repo.
	github.com/http-wasm/http-wasm-host-go v0.7.0
	github.com/influxdata/influxdb-client-go/v2 v2.7.0
	github.com/influxdata/influxdb1-client v0.0.0-20200827194710-b269163b24ab // No tag on the repo.
	github.com/klauspost/compress v1.18.0
	github.com/kvtools/consul v1.0.2
	github.com/kvtools/etcdv3 v1.0.3
	github.com/kvtools/redis v1.2.0
	github.com/kvtools/valkeyrie v1.0.0
	github.com/kvtools/zookeeper v1.0.2
	github.com/mailgun/ttlmap v0.0.0-20170619185759-c1c17f74874f // No tag on the repo.
	github.com/miekg/dns v1.1.69
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.5.1-0.20231216201459-8508981c8b6c // No tag on the repo.
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pires/go-proxyproto v0.8.1
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // No tag on the repo.
	github.com/prometheus/client_golang v1.23.0
	github.com/prometheus/client_model v0.6.2
	github.com/quic-go/quic-go v0.59.0
	github.com/redis/go-redis/v9 v9.8.0
	github.com/rs/zerolog v1.33.0
	github.com/sirupsen/logrus v1.9.3
	github.com/spiffe/go-spiffe/v2 v2.6.0
	github.com/stealthrocket/wasi-go v0.8.0
	github.com/stealthrocket/wazergo v0.19.1
	github.com/stretchr/testify v1.11.1
	github.com/stvp/go-udp-testing v0.0.0-20201019212854-469649b16807 // No tag on the repo.
	github.com/tailscale/tscert v0.0.0-20230806124524-28a91b69a046 // No tag on the repo.
	github.com/testcontainers/testcontainers-go v0.32.0
	github.com/testcontainers/testcontainers-go/modules/k3s v0.32.0
	github.com/tetratelabs/wazero v1.8.0
	github.com/tidwall/gjson v1.17.0
	github.com/traefik/grpc-web v0.16.0
	github.com/traefik/paerser v0.2.2
	github.com/traefik/yaegi v0.16.1
	github.com/unrolled/render v1.0.2
	github.com/unrolled/secure v1.0.9
	github.com/valyala/fasthttp v1.58.0
	github.com/vulcand/oxy/v2 v2.0.3
	github.com/vulcand/predicate v1.2.0
	github.com/yuin/gopher-lua v1.1.1
	go.opentelemetry.io/collector/pdata v1.41.0
	go.opentelemetry.io/contrib/bridges/otellogrus v0.13.0
	go.opentelemetry.io/contrib/propagators/autoprop v0.63.0
	go.opentelemetry.io/otel v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc v0.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp v0.14.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.38.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.38.0
	go.opentelemetry.io/otel/log v0.14.0
	go.opentelemetry.io/otel/metric v1.38.0
	go.opentelemetry.io/otel/sdk v1.38.0
	go.opentelemetry.io/otel/sdk/log v0.14.0
	go.opentelemetry.io/otel/sdk/metric v1.38.0
	go.opentelemetry.io/otel/trace v1.38.0
	golang.org/x/crypto v0.47.0
	golang.org/x/mod v0.31.0
	golang.org/x/net v0.49.0
	golang.org/x/sync v0.19.0
	golang.org/x/sys v0.40.0
	golang.org/x/text v0.33.0
	golang.org/x/time v0.14.0
	golang.org/x/tools v0.40.0
	google.golang.org/grpc v1.78.0
	gopkg.in/natefinch/lumberjack.v2 v2.2.1
	gopkg.in/yaml.v3 v3.0.1
	k8s.io/api v0.34.3
	k8s.io/apiextensions-apiserver v0.34.3
	k8s.io/apimachinery v0.34.3
	k8s.io/client-go v0.34.3
	k8s.io/utils v0.0.0-20250820121507-0af2bda4dd1d // No tag on the repo.
	knative.dev/networking v0.0.0-20241022012959-60e29ff520dc
	knative.dev/pkg v0.0.0-20241021183759-9b9d535af5ad
	mvdan.cc/xurls/v2 v2.5.0
	sigs.k8s.io/controller-runtime v0.22.1
	sigs.k8s.io/gateway-api v1.4.0
	sigs.k8s.io/structured-merge-diff/v6 v6.3.1
	sigs.k8s.io/yaml v1.6.0
)

require (
	cloud.google.com/go/auth v0.18.0 // indirect
	cloud.google.com/go/auth/oauth2adapt v0.2.8 // indirect
	cloud.google.com/go/compute/metadata v0.9.0 // indirect
	dario.cat/mergo v1.0.1 // indirect
	github.com/AdamSLevy/jsonrpc2/v14 v14.1.0 // indirect
	github.com/Azure/azure-sdk-for-go v68.0.0+incompatible // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.20.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.13.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dns/armdns v1.2.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns v1.3.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resourcegraph/armresourcegraph v0.9.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible // indirect
	github.com/Azure/go-autorest/autorest v0.11.30 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.22 // indirect
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.13 // indirect
	github.com/Azure/go-autorest/autorest/azure/cli v0.4.6 // indirect
	github.com/Azure/go-autorest/autorest/date v0.3.0 // indirect
	github.com/Azure/go-autorest/autorest/to v0.4.1 // indirect
	github.com/Azure/go-autorest/logger v0.2.1 // indirect
	github.com/Azure/go-autorest/tracing v0.6.0 // indirect
	github.com/AzureAD/microsoft-authentication-library-for-go v1.6.0 // indirect
	github.com/HdrHistogram/hdrhistogram-go v1.1.2 // indirect
	github.com/Masterminds/goutils v1.1.1 // indirect
	github.com/Masterminds/semver/v3 v3.3.1 // indirect
	github.com/Microsoft/go-winio v0.6.2 // indirect
	github.com/Microsoft/hcsshim v0.13.0 // indirect
	github.com/VividCortex/gohistogram v1.0.0 // indirect
	github.com/akamai/AkamaiOPEN-edgegrid-golang/v11 v11.1.0 // indirect
	github.com/alibabacloud-go/alibabacloud-gateway-spi v0.0.5 // indirect
	github.com/alibabacloud-go/darabonba-openapi/v2 v2.1.13 // indirect
	github.com/alibabacloud-go/debug v1.0.1 // indirect
	github.com/alibabacloud-go/tea v1.4.0 // indirect
	github.com/alibabacloud-go/tea-utils/v2 v2.0.7 // indirect
	github.com/aliyun/credentials-go v1.4.7 // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.18.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.16 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.16 // indirect
	github.com/aws/aws-sdk-go-v2/service/lightsail v1.50.10 // indirect
	github.com/aws/aws-sdk-go-v2/service/route53 v1.62.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/signin v1.0.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.30.8 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.35.12 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.41.5 // indirect
	github.com/aziontech/azionapi-go-sdk v0.144.0 // indirect
	github.com/baidubce/bce-sdk-go v0.9.256 // indirect
	github.com/benbjohnson/clock v1.3.5 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/blendle/zapdriver v1.3.1 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/bytedance/sonic v1.10.0 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/clbanning/mxj/v2 v2.7.0 // indirect
	github.com/containerd/containerd v1.7.23 // indirect
	github.com/containerd/errdefs v1.0.0 // indirect
	github.com/containerd/errdefs/pkg v0.3.0 // indirect
	github.com/containerd/log v0.1.0 // indirect
	github.com/containerd/platforms v1.0.0-rc.1 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/cpuguy83/dockercfg v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/deepmap/oapi-codegen v1.9.1 // indirect
	github.com/desertbit/timer v0.0.0-20180107155436-c41aec40b27f // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dimchansky/utfbom v1.1.1 // indirect
	github.com/distribution/reference v0.6.0 // indirect
	github.com/dnsimple/dnsimple-go/v4 v4.0.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/emicklei/go-restful/v3 v3.13.0 // indirect
	github.com/evanphx/json-patch/v5 v5.9.11 // indirect
	github.com/exoscale/egoscale/v3 v3.1.33 // indirect
	github.com/fatih/color v1.18.0 // indirect
	github.com/felixge/httpsnoop v1.0.4 // indirect
	github.com/fxamacker/cbor/v2 v2.9.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.3 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/gin-gonic/gin v1.9.1 // indirect
	github.com/go-acme/alidns-20150109/v4 v4.7.0 // indirect
	github.com/go-acme/esa-20240910/v2 v2.44.0 // indirect
	github.com/go-acme/jdcloud-sdk-go v1.64.0 // indirect
	github.com/go-acme/tencentclouddnspod v1.1.25 // indirect
	github.com/go-acme/tencentedgdeone v1.1.48 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-jose/go-jose/v4 v4.1.3 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-logr/zapr v1.3.0 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-openapi/jsonpointer v0.21.2 // indirect
	github.com/go-openapi/jsonreference v0.21.0 // indirect
	github.com/go-openapi/swag v0.23.1 // indirect
	github.com/go-ozzo/ozzo-validation/v4 v4.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.23.0 // indirect
	github.com/go-resty/resty/v2 v2.17.1 // indirect
	github.com/go-viper/mapstructure/v2 v2.4.0 // indirect
	github.com/go-zookeeper/zk v1.0.3 // indirect
	github.com/goccy/go-yaml v1.18.0 // indirect
	github.com/gofrs/flock v0.13.0 // indirect
	github.com/gofrs/uuid v4.4.0+incompatible // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang-jwt/jwt/v4 v4.5.2 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gnostic-models v0.7.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.2.0 // indirect
	github.com/google/s2a-go v0.1.9 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/googleapis/enterprise-certificate-proxy v0.3.7 // indirect
	github.com/googleapis/gax-go/v2 v2.16.0 // indirect
	github.com/gophercloud/gophercloud v1.14.1 // indirect
	github.com/gophercloud/utils v0.0.0-20231010081019-80377eca5d56 // indirect
	github.com/gravitational/trace v1.1.16-0.20220114165159-14a9a7dd6aaf // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/hashicorp/cronexpr v1.1.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/hashicorp/golang-lru v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.1-vault-5 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/huandu/xstrings v1.5.0 // indirect
	github.com/huaweicloud/huaweicloud-sdk-go-v3 v0.1.182 // indirect
	github.com/iij/doapi v0.0.0-20190504054126-0bbf12d6d7df // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/influxdata/line-protocol v0.0.0-20200327222509-2487e7298839 // indirect
	github.com/infobloxopen/infoblox-go-client/v2 v2.10.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.13-0.20220915233716-71ac16282d12 // indirect
	github.com/k0kubun/go-ansi v0.0.0-20180517002512-3bf9e2903213 // indirect
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/kolo/xmlrpc v0.0.0-20220921171641-a4b6fa1dd06b // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/labbsr0x/bindman-dns-webhook v1.0.2 // indirect
	github.com/labbsr0x/goh v1.0.1 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/linode/linodego v1.64.0 // indirect
	github.com/liquidweb/liquidweb-cli v0.6.9 // indirect
	github.com/liquidweb/liquidweb-go v1.6.4 // indirect
	github.com/lufia/plan9stats v0.0.0-20250317134145-8bc96cf8fc35 // indirect
	github.com/magiconair/properties v1.8.7 // indirect
	github.com/mailgun/minheap v0.0.0-20170619185613-3dbe6c6bf55f // indirect
	github.com/mailgun/multibuf v0.1.2 // indirect
	github.com/mailgun/timetools v0.0.0-20141028012446-7e6055773c51 // indirect
	github.com/mailru/easyjson v0.9.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mimuret/golang-iij-dpf v0.9.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-ps v1.0.0 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/moby/docker-image-spec v1.3.1 // indirect
	github.com/moby/go-archive v0.1.0 // indirect
	github.com/moby/patternmatcher v0.6.0 // indirect
	github.com/moby/spdystream v0.5.0 // indirect
	github.com/moby/sys/atomicwriter v0.1.0 // indirect
	github.com/moby/sys/sequential v0.6.0 // indirect
	github.com/moby/sys/user v0.4.0 // indirect
	github.com/moby/sys/userns v0.1.0 // indirect
	github.com/moby/term v0.5.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.3-0.20250322232337-35a7c28c31ee // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/mxk/go-flowrate v0.0.0-20140419014527-cca7078d478f // indirect
	github.com/namedotcom/go/v4 v4.0.2 // indirect
	github.com/nrdcg/auroradns v1.1.0 // indirect
	github.com/nrdcg/bunny-go v0.1.0 // indirect
	github.com/nrdcg/desec v0.11.1 // indirect
	github.com/nrdcg/dnspod-go v0.4.0 // indirect
	github.com/nrdcg/freemyip v0.3.0 // indirect
	github.com/nrdcg/goacmedns v0.2.0 // indirect
	github.com/nrdcg/goinwx v0.12.0 // indirect
	github.com/nrdcg/mailinabox v0.3.0 // indirect
	github.com/nrdcg/namesilo v0.5.0 // indirect
	github.com/nrdcg/nodion v0.1.0 // indirect
	github.com/nrdcg/oci-go-sdk/common/v1065 v1065.105.2 // indirect
	github.com/nrdcg/oci-go-sdk/dns/v1065 v1065.105.2 // indirect
	github.com/nrdcg/porkbun v0.4.0 // indirect
	github.com/nrdcg/vegadns v0.3.0 // indirect
	github.com/nzdjb/go-metaname v1.0.0 // indirect
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.1.1 // indirect
	github.com/openzipkin/zipkin-go v0.4.3 // indirect
	github.com/ovh/go-ovh v1.9.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/peterhellberg/link v1.2.0 // indirect
	github.com/pkg/browser v0.0.0-20240102092130-5ac0b6a4141c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55 // indirect
	github.com/pquerna/otp v1.5.0 // indirect
	github.com/prometheus/common v0.65.0 // indirect
	github.com/prometheus/procfs v0.17.0 // indirect
	github.com/quic-go/qpack v0.6.0 // indirect
	github.com/regfish/regfish-dnsapi-go v0.1.1 // indirect
	github.com/rs/cors v1.7.0 // indirect
	github.com/rs/dnscache v0.0.0-20230804202142-fc85eb664529 // indirect
	github.com/sacloud/api-client-go v0.3.3 // indirect
	github.com/sacloud/go-http v0.1.9 // indirect
	github.com/sacloud/iaas-api-go v1.23.1 // indirect
	github.com/sacloud/packages-go v0.0.12 // indirect
	github.com/sagikazarmark/locafero v0.4.0 // indirect
	github.com/sagikazarmark/slog-shim v0.1.0 // indirect
	github.com/scaleway/scaleway-sdk-go v1.0.0-beta.36 // indirect
	github.com/selectel/domains-go v1.1.0 // indirect
	github.com/selectel/go-selvpcclient/v4 v4.1.0 // indirect
	github.com/shirou/gopsutil/v3 v3.24.4 // indirect
	github.com/shoenig/go-m1cpu v0.1.6 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/softlayer/softlayer-go v1.2.1 // indirect
	github.com/softlayer/xmlrpc v0.0.0-20200409220501-5f089df7cb7e // indirect
	github.com/sony/gobreaker v1.0.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/spf13/afero v1.11.0 // indirect
	github.com/spf13/cast v1.7.0 // indirect
	github.com/spf13/pflag v1.0.7 // indirect
	github.com/spf13/viper v1.18.2 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common v1.3.28 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tjfoc/gmsm v1.4.1 // indirect
	github.com/tklauser/go-sysconf v0.3.15 // indirect
	github.com/tklauser/numcpus v0.10.0 // indirect
	github.com/transip/gotransip/v6 v6.26.1 // indirect
	github.com/ultradns/ultradns-go-sdk v1.8.1-20250722213956-faef419 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/vinyldns/go-vinyldns v0.9.17 // indirect
	github.com/volcengine/volc-sdk-golang v1.0.233 // indirect
	github.com/vultr/govultr/v3 v3.26.1 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	github.com/yandex-cloud/go-genproto v0.43.0 // indirect
	github.com/yandex-cloud/go-sdk/services/dns v0.0.25 // indirect
	github.com/yandex-cloud/go-sdk/v2 v2.37.0 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	github.com/yusufpapurcu/wmi v1.2.4 // indirect
	go.etcd.io/etcd/api/v3 v3.6.4 // indirect
	go.etcd.io/etcd/client/pkg/v3 v3.6.4 // indirect
	go.etcd.io/etcd/client/v3 v3.6.4 // indirect
	go.mongodb.org/mongo-driver v1.13.1 // indirect
	go.opencensus.io v0.24.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/collector/featuregate v1.41.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.61.0 // indirect
	go.opentelemetry.io/contrib/propagators/aws v1.38.0 // indirect
	go.opentelemetry.io/contrib/propagators/b3 v1.38.0 // indirect
	go.opentelemetry.io/contrib/propagators/jaeger v1.38.0 // indirect
	go.opentelemetry.io/contrib/propagators/ot v1.38.0 // indirect
	go.opentelemetry.io/proto/otlp v1.7.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/ratelimit v0.3.1 // indirect
	go.uber.org/zap v1.27.0 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/arch v0.4.0 // indirect
	golang.org/x/exp v0.0.0-20241210194714-1829a127f884 // indirect
	golang.org/x/oauth2 v0.34.0 // indirect
	golang.org/x/term v0.39.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.4.0 // indirect
	google.golang.org/api v0.259.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20251202230838-ff82c1b0f217 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251222181119-0a764e51fe1b // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/evanphx/json-patch.v4 v4.13.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
	gopkg.in/ns1/ns1-go.v2 v2.16.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	k8s.io/klog/v2 v2.130.1 // indirect
	k8s.io/kube-openapi v0.0.0-20250814151709-d7b6acb124c3 // indirect
	nhooyr.io/websocket v1.8.7 // indirect
	sigs.k8s.io/json v0.0.0-20250730193827-2d320260d730 // indirect
	sigs.k8s.io/randfill v1.0.0 // indirect
)

// Containous forks
replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/gorilla/mux => github.com/containous/mux v0.0.0-20250523120546-41b6ec3aed59
	github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595
)

// ambiguous import: found package github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/http in multiple modules
// tencentcloud uses monorepo with multimodule but the go.mod files are incomplete.
exclude github.com/tencentcloud/tencentcloud-sdk-go v3.0.83+incompatible
