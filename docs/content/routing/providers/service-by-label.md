In general when configuring a Traefik provider,
a service assigned to one (or several) router(s) must be defined as well for the routing to be functional.

There are, however, exceptions when using label-based configurations:

1. When there is only one service and the router does not specify any service,
then that service is automatically assigned to the router.
2. When a label defines a router and there is no service defined,
then a service is automatically created and assigned to the router.

!!! info ""
    As one would expect, in either of these cases, if in addition a service is specified for the router,
    then that service is the one assigned, regardless of whether it actually is defined or whatever else other services are defined.
