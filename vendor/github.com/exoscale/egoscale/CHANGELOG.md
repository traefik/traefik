Changelog
=========

0.14.3
------

- fix: `AffinityGroup` lists virtual machines with `UUID` rather than string

0.14.2
------

- fix: `ListVirtualMachines` by `IDs` to accept `UUID` rather than string

0.14.1
------

- fix: `GetRunstatusPage` to always contain the subresources
- fix: `ListRunstatus*` to fetch all the subresources
- feature: `PaginateRunstatus*` used by list

0.14.0
------

- change: all DNS calls require a context
- fix: `CreateAffinityGroup` allows empty `name`

0.13.3
------

- fix: runstatus unmarshalling errors
- feature: `UUID` implements DeepCopy, DeepCopyInto
- change: export `BooleanResponse`

0.13.2
------

- feat: initial Runstatus API support
- feat: `admin` namespace containing `ListVirtualMachines` for admin usage

0.13.1
------

- feat: `Iso` support `ListIsos`, `AttachIso`, and `DetachIso`

0.13.0
------

- change: `Paginate` to accept `Listable`
- change: `ListCommand` is also `Listable`
- change: `client.Get` doesn't modify the given resource, returns a new one
- change: `Command` and `AsyncCommand` are fully public, thus extensible
- remove: `Gettable`

0.12.5
------

- fix: `AuthorizeSecurityGroupEgress` could return `authorizeSecurityGroupIngress` as name

0.12.4
------

- feat: `Snapshot` is `Listable`

0.12.3
------

- change: replace dep by Go modules
- change: remove domainid,domain,regionid,listall,isrecursive,... fields
- remove: `MigrateVirtualMachine`, `CreateUser`, `EnableAccount`, and other admin calls

0.12.2
------

- fix: `ListNics` has no virtualmachineid limitations anymore
- fix: `PCIDevice` ids are not UUIDs

0.12.1
------

- fix: `UpdateVMNicIP` is async

0.12.0
------

- feat: new VM state `Moving`
- feat: `UpdateNetwork` with `startip`, `endip`, `netmask`
- feat: `NetworkOffering` is `Listable`
- feat: when it fails parsing the body, it shows it
- fix: `Snapshot.State` is a string, rather than an scalar
- change: signature are now using the v3 version with expires by default

0.11.6
------

- fix: `Network.ListRequest` accepts a `Name` argument
- change: `SecurityGroup` and the rules aren't `Taggable` anymore

0.11.5
------

- feat: addition of `UpdateVMNicIP`
- fix: `UpdateVMAffinityGroup` expected response

0.11.4
------

*no changes in the core library*

0.11.3
------

*no changes in the core library*

0.11.2
------

- fix: empty list responses

0.11.1
------

- fix: `client.Sign` handles correctly the brackets (kudos to @stffabi)
- change: `client.Payload` returns a `url.Values`

0.11.0
------

- feat: `listOSCategories` and `OSCategory` type
- feat: `listApis` supports recursive response structures
- feat: `GetRecordsWithFilters` to list records with name or record_type filters
- fix: better `DNSErrorResponse`
- fix: `ListResourceLimits` type
- change: use UUID everywhere

0.10.5
------

- feat: `Client.Logger` to plug in any `*log.Logger`
- feat: `Client.TraceOn`/`ClientTraceOff` to toggle the HTTP tracing

0.10.4
------

- feat: `CIDR` to replace string string
- fix: prevent panic on nil

0.10.3
------

- feat: `Account` is Listable
- feat: `MACAddress` to replace string type
- fix: Go 1.7 support

0.10.2
------

- fix: ActivateIP6 response

0.10.1
------

- feat: expose `SyncRequest` and `SyncRequestWithContext`
- feat: addition of reverse DNS calls
- feat: addition of `SecurityGroup.UserSecurityGroup`

0.10.0
------

- global: cloudstack documentation links are moved into cs
- global: removal of all the `...Response` types
- feat: `Network` is `Listable`
- feat: addition of `deleteUser`
- feat: addition of `listHosts`
- feat: addition of `updateHost`
- feat: exo cmd (kudos to @pierre-emmanuelJ)
- change: refactor `Gettable` to use `ListRequest`

0.9.31
------

- fix: `IPAddress`.`ListRequest` with boolean fields
- fix: `Network`.`ListRequest` with boolean fields
- fix: `ServiceOffering`.`ListRequest` with boolean fields

0.9.30
------

- fix: `VirtualMachine` `PCIDevice` representation was incomplete

0.9.29
------

- change: `DNSErrorResponse` is a proper `error`

0.9.28
------

- feat: addition of `GetDomains`
- fix: `UpdateDomain` may contain more empty fields than `CreateDomain`

0.9.27
------

- fix: expects body to be `application/json`

0.9.26
------

- change: async timeout strategy wait two seconds and not fib(n) seconds

0.9.25
------

- fix: `GetVirtualUserData` response with `Decode` method handling base64 and gzip

0.9.24
------

- feat: `Template` is `Gettable`
- feat: `ServiceOffering` is `Gettable`
- feat: addition of `GetAPILimit`
- feat: addition of `CreateTemplate`, `PrepareTemplate`, `CopyTemplate`, `UpdateTemplate`, `RegisterTemplate`
- feat: addition of `MigrateVirtualMachine`
- feat: cmd cli
- change: remove useless fields related to Project and VPC

0.9.23
------

- feat: `booleanResponse` supports true booleans: https://github.com/apache/cloudstack/pull/2428

0.9.22
------

- feat: `ListUsers`, `CreateUser`, `UpdateUser`
- feat: `ListResourceDetails`
- feat: `SecurityGroup` helper `RuleByID`
- feat: `Sign` signs the payload
- feat: `UpdateNetworkOffering`
- feat: `GetVirtualMachineUserData`
- feat: `EnableAccount` and `DisableAccount` (admin stuff)
- feat: `AsyncRequest` and `AsyncRequestWithContext` to examine the polling
- fix: `AuthorizeSecurityGroupIngress` support for ICMPv6
- change: move `APIName()` into the `Client`, nice godoc
- change: `Payload` doesn't sign the request anymore
- change: `Client` exposes more of its underlying data
- change: requests are sent as GET unless it body size is too big

0.9.21
------

- feat: `Network` is `Listable`
- feat: `Zone` is `Gettable`
- feat: `Client.Payload` to help preview the HTTP parameters
- feat: generate command utility
- fix: `CreateSnapshot` was missing the `Name` attribute
- fix: `ListSnapshots` was missing the `IDs` attribute
- fix: `ListZones` was missing the `NetworkType` attribute
- fix: `ListAsyncJobs` was missing the `ListAll` attribute
- change: ICMP Type/Code are uint8 and TCP/UDP port are uint16

0.9.20
------

- feat: `Template` is `Listable`
- feat: `IPAddress` is `Listable`
- change: `List` and `Paginate` return pointers
- fix: `Template` was missing `tags`

0.9.19
------

- feat: `SSHKeyPair` is `Listable`

0.9.18
------

- feat: `VirtualMachine` is `Listable`
- feat: new `Client.Paginate` and `Client.PaginateWithContext`
- change: the inner logic of `Listable`
- remove: not working `Client.AsyncList`

0.9.17
------

- fix: `AuthorizeSecurityGroup(In|E)gress` startport may be zero

0.9.16
------

- feat: new `Listable` interface
- feat: `Nic` is `Listable`
- feat: `Volume` is `Listable`
- feat: `Zone` is `Listable`
- feat: `AffinityGroup` is `Listable`
- remove: deprecated methods `ListNics`, `AddIPToNic`, and `RemoveIPFromNic`
- remove: deprecated method `GetRootVolumeForVirtualMachine`

0.9.15
------

- feat: `IPAddress` is `Gettable` and `Deletable`
- fix: serialization of *bool

0.9.14
------

- fix: `GetVMPassword` response
- remove: deprecated `GetTopology`, `GetImages`, and al

0.9.13
------

- feat: IP4 and IP6 flags to DeployVirtualMachine
- feat: add ActivateIP6
- fix: error message was gobbled on 40x

0.9.12
------

- feat: add `BooleanRequestWithContext`
- feat: add `client.Get`, `client.GetWithContext` to fetch a resource
- feat: add `cleint.Delete`, `client.DeleteWithContext` to delete a resource
- feat: `SSHKeyPair` is `Gettable` and `Deletable`
- feat: `VirtualMachine` is `Gettable` and `Deletable`
- feat: `AffinityGroup` is `Gettable` and `Deletable`
- feat: `SecurityGroup` is `Gettable` and `Deletable`
- remove: deprecated methods `CreateAffinityGroup`, `DeleteAffinityGroup`
- remove: deprecated methods `CreateKeypair`, `DeleteKeypair`, `RegisterKeypair`
- remove: deprecated method `GetSecurityGroupID`

0.9.11
------

- feat: CloudStack API name is now public `APIName()`
- feat: enforce the mutual exclusivity of some fields
- feat: add `context.Context` to `RequestWithContext`
- change: `AsyncRequest` and `BooleanAsyncRequest` are gone, use `Request` and `BooleanRequest` instead.
- change: `AsyncInfo` is no more

0.9.10
------

- fix: typo made ListAll required in ListPublicIPAddresses
- fix: all bool are now *bool, respecting CS default value
- feat: (*VM).DefaultNic() to obtain the main Nic

0.9.9
-----

- fix: affinity groups virtualmachineIds attribute
- fix: uuidList is not a list of strings

0.9.8
-----

- feat: add RootDiskSize to RestoreVirtualMachine
- fix: monotonic polling using Context

0.9.7
-----

- feat: add Taggable interface to expose ResourceType
- feat: add (Create|Update|Delete|List)InstanceGroup(s)
- feat: add RegisterUserKeys
- feat: add ListResourceLimits
- feat: add ListAccounts

0.9.6
-----

- fix: update UpdateVirtualMachine userdata
- fix: Network's name/displaytext might be empty

0.9.5
-----

- fix: serialization of slice

0.9.4
-----

- fix: constants

0.9.3
-----

- change: userdata expects a string
- change: no pointer in sub-struct's

0.9.2
-----

- bug: createNetwork is a sync call
- bug: typo in listVirtualMachines' domainid
- bug: serialization of map[string], e.g. UpdateVirtualMachine
- change: IPAddress's use net.IP type
- feat: helpers VM.NicsByType, VM.NicByNetworkID, VM.NicByID
- feat: addition of CloudStack ApiErrorCode constants

0.9.1
-----

- bug: sync calls returns succes as a string rather than a bool
- change: unexport BooleanResponse types
- feat: original CloudStack error response can be obtained

0.9.0
-----

Big refactoring, addition of the documentation, compliance to golint.

0.1.0
-----

Initial library
