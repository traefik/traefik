Starting build size on darwin/arm64: 127M

Without embedded webui: 122M

Checking for the largest crap:
go tool nm -sort size -size dist/traefik | head -n 20
 121bd40     770928 T github.com/aws/aws-sdk-go/aws/endpoints.init
 4d8aae0     256906 r runtime.findfunctab
 4dc9800     252448 r runtime.typelink
  6f7090     184864 T github.com/traefik/yaegi/stdlib.init.38
 1b3fbd0     135120 T github.com/yandex-cloud/go-genproto/yandex/cloud/mdb/postgresql/v1/config.init
 16bae90     133776 T github.com/sacloud/iaas-api-go/fake.init
 7a963a0     101376 D runtime.mheap_
 236f5b0      88752 T k8s.io/api/core/v1.init
  7f8320      81632 T github.com/traefik/yaegi/stdlib.init.144
 4e07220      81584 r runtime.itablink
  6e5e50      69808 T github.com/traefik/yaegi/stdlib.init.37
 7a862a0      65784 D runtime.trace
 2b27a30      58816 T k8s.io/client-go/informers.(*sharedInformerFactory).ForResource
  d1cb50      58304 T github.com/civo/civogo.decodeError
  7a6520      51808 T github.com/traefik/yaegi/stdlib.init.102
  761da0      42256 T github.com/traefik/yaegi/stdlib.init.71
  3e1390      42192 T github.com/traefik/yaegi/interp.(*Interpreter).cfg.func2
  724d20      35408 T github.com/traefik/yaegi/stdlib.init.40
  7be840      32944 T github.com/traefik/yaegi/stdlib.init.117
 7a7e520      32128 D runtime.semtable
go tool nm: signal: broken pipe

So we need to remove aws-sdk-go - fun......., can't do anything about yaegi that is their interpreter

Anyway I started nuking things in pkg/providers...

Removing the ECS provider moved AWS to an indirect dependency at least... 

Looks like it will be difficult to remove...

Remove all the unused providers results: 88M

One of the last big things seems to be ACME but that is more tightly integrated.

