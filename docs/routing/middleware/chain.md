# Chain

When One Isn't Enougth
{: .subtitle }

![Chain](../../img/middleware/chain.png)

The Chain middleware enables you to define reusable combinations of other pieces of middleware. 
It makes reusing the same groups easier.

## Configuration Example

??? example "A Chain for WhiteList, BasicAuth, and HTTPS"
    
    ```toml
    # ...    
    [Routers]
        [Routers.router1]
            service = "service1"
            middlewares = ["secured"]
            rule = "Host: mydomain"
    
    [Middlewares]
        [Middlewares.secured.Chain]
            middlewares = ["https-only", "known-ips", "auth-users"]
            
        [Middlewares.auth-users.BasicAuth]
            users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"]
        [Middlewares.https-only.SchemeRedirect]
            scheme = "https"
        [Middlewares.known-ips.ipWhiteList]
            sourceRange = ["192.168.1.7", "x.x.x.x", "x.x.x.x"]
    
    [Services]
      [Services.service1]
        [Services.service1.LoadBalancer]
          [[Services.service1.LoadBalancer.Servers]]
            URL = "http://127.0.0.1:80"
            Weight = 1
    ```