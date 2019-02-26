| `traefik.domain`                                                        | Sets the default base domain for the frontend rules. For more information, check the [Container Labels section's of the user guide "Let's Encrypt & Docker"](/user-guide/docker-and-lets-encrypt/#container-labels)              |
| `traefik.port=80`                                                       | Registers this port. Useful when the container exposes multiples ports.                                                                                                                                                          |
| `traefik.protocol=https`                                                | Overrides the default `http` protocol                                                                                                                                                                                            |
| `traefik.weight=10`                                                     | Assigns this weight to the container                                                                                                                                                                                             

[2] `traefik.frontend.auth.basic.users=EXPR`:  
To create `user:password` pair, it's possible to use this command:  
`echo $(htpasswd -nb user password) | sed -e s/\\$/\\$\\$/g`.  
The result will be `user:$$apr1$$9Cv/OMGj$$ZomWQzuQbL.3TRCS81A1g/`, note additional symbol `$` makes escaping.

[3] `traefik.backend.loadbalancer.swarm`:  
If you enable this option, Traefik will use the virtual IP provided by docker swarm instead of the containers IPs.
Which means that Traefik will not perform any kind of load balancing and will delegate this task to swarm.  
It also means that Traefik will manipulate only one backend, not one backend per container.

!!! warning
    When running inside a container, Traefik will need network access through:

    `docker network connect <network> <traefik-container>`