In general when configuring a Traefik provider,
a service assigned to one (or several) router(s) must be defined as well for the routing to be functional.

There are, however, exceptions when using label-based configurations:

1. If a label defines a router (e.g. through a router Rule)
and a label defines a service (e.g. implicitly through a loadbalancer server port value),
but the router does not specify any service,
then that service is automatically assigned to the router.

2. If a label defines a router (e.g. through a router Rule) but no service is defined,
then a service is automatically created and assigned to the router.

!!! info ""
    As one would expect, in either of these cases, if in addition a service is specified for the router,
    then that service is the one assigned, regardless of whether it actually is defined or whatever else other services are defined.
