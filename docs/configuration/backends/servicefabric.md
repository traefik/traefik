# Service Fabric Backend

Træfik can be configured to use Service Fabric as a backend configuration.

See [this repository for an example deployment package and further documentation.](https://aka.ms/traefikonsf)

## Service Fabric

```################################################################
# Service Fabric provider
################################################################

# Enable Service Fabric configuration backend
[servicefabric]

# Service Fabric Management Endpoint
clustermanagementurl = "https://localhost:19080"

# Service Fabric Management Endpoint API Version
apiversion = "3.0"

# Client certifcate file path
clientcertfilepath = "certs/clientkey.cer"

# Client certifcate key file path
clientcertkeyfilepath = "certs/clientkey.key"

# Skip TLS verify
InsecureSkipVerify=true
```

## Labels

The provider uses labels to configure how services are exposed through Traefik. These can be set using Extensions and the Property Manager API

#### Extensions

Set labels with extensions through the services `ServiceManifest.xml` file. Here is an example of an extension setting Traefik labels:

```
    <StatelessServiceType ServiceTypeName="WebServiceType">
      <Extensions>
          <Extension Name="Traefik">
            <Labels xmlns="http://schemas.microsoft.com/2015/03/fabact-no-schema">
              <Label Key="traefik.frontend.rule.example2">PathPrefixStrip: /a/path/to/strip</Label>
              <Label Key="traefik.expose">true</Label>
              <Label Key="traefik.frontend.passHostHeader">true</Label>
            </Labels>
          </Extension>
      </Extensions>
    </StatelessServiceType>
```

#### Property Manager 

Set Labels with the property manager API to overwrite and add labels, while your service in running. Here is an example of adding a frontend rule using the property manager API. 

```
curl -X PUT \
  'http://localhost:19080/Names/GettingStartedApplication2/WebService/$/GetProperty?api-version=6.0&IncludeValues=true' \
  -d '{
  "PropertyName": "traefik.frontend.rule.default",
  "Value": {
    "Kind": "String",
    "Data": "PathPrefixStrip: /a/path/to/strip"
  },
  "CustomTypeId": "LabelType"
}'
```

**Note**: This functionality will be released in a future version of the [sfctl](https://docs.microsoft.com/en-us/azure/service-fabric/service-fabric-application-lifecycle-sfctl) tool.

## Available Labels

Labels, set through extensions or the property manager, can be used on services to override default behaviour.

| Label                                                     | Description                                                                                                                                                                                                         |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `traefik.backend.maxconn.amount=10`                       | Set a maximum number of connections to the backend. Must be used in conjunction with the below label to take effect.                                                                                                |
| `traefik.backend.maxconn.extractorfunc=client.ip`         | Set the function to be used against the request to determine what to limit maximum connections to the backend by. Must be used in conjunction with the above label to take effect.                                  |
| `traefik.backend.loadbalancer.method=drr`                 | Override the default `wrr` load balancer algorithm                                                                                                                                                                  |
| `traefik.backend.loadbalancer.stickiness=true`            | Enable backend sticky sessions                                                                                                                                                                                      |
| `traefik.backend.loadbalancer.stickiness.cookieName=NAME` | Manually set the cookie name for sticky sessions                                                                                                                                                                    |
| `traefik.backend.circuitbreaker.expression=EXPR`          | Create a [circuit breaker](/basics/#backends) to be used against the backend                                                                                                                                        |
| `traefik.backend.weight=10`                               | Assign this weight to the container                                                                                                                                                                                 |
| `traefik.expose=true`                                     | Expose this service using traefik                                                                                                                                                                                   |
| `traefik.frontend.rule=EXPR`                              | Override the default frontend rule. Defaults to SF address.                                                                                                                                                         |
| `traefik.frontend.passHostHeader=true`                    | Forward client `Host` header to the backend.                                                                                                                                                                        |
| `traefik.frontend.priority=10`                            | Override default frontend priority                                                                                                                                                                                  |
| `traefik.frontend.entryPoints=http,https`                 | Assign this frontend to entry points `http` and `https`. Overrides `defaultEntryPoints`                                                                                                                             |
| `traefik.frontend.auth.basic=EXPR`                        | Sets basic authentication for that frontend in CSV format: `User:Hash,User:Hash`                                                                                                                                    |
| `traefik.frontend.whitelistSourceRange:RANGE`             | List of IP-Ranges which are allowed to access. An unset or empty list allows all Source-IPs to access. If one of the Net-Specifications are invalid, the whole list is invalid and allows all Source-IPs to access. |
| `traefik.backend.group.name`                              | Groups all services with the same name into a single backend in Traefik                                                                                                                                             |
| `traefik.backend.group.weight`                            | Sets the weighting of the current services nodes in the backend group                                                                                                                                               |


