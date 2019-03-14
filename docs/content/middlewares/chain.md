# Chain

When One Isn't Enougth
{: .subtitle }

![Chain](../assets/img/middleware/chain.png)

The Chain middleware enables you to define reusable combinations of other pieces of middleware. 
It makes reusing the same groups easier.

## Configuration Example

??? example "A Chain for WhiteList, BasicAuth, and HTTPS"
    
    ```toml
    # ...    
    [http.routers]
        [http.routers.router1]
            service = "service1"
            middlewares = ["secured"]
            rule = "Host: mydomain"
    
    [http.middlewares]
        [http.middlewares.secured.Chain]
            middlewares = ["https-only", "known-ips", "auth-users"]
            
        [http.middlewares.auth-users.BasicAuth]
            users = ["test:$apr1$H6uskkkW$IgXLP6ewTrSuBkTrqE8wj/"]
        [http.middlewares.https-only.SchemeRedirect]
            scheme = "https"
        [http.middlewares.known-ips.ipWhiteList]
            sourceRange = ["192.168.1.7", "x.x.x.x", "x.x.x.x"]
    
    [http.services]
      [http.services.service1]
        [http.services.service1.LoadBalancer]
          [[http.services.service1.LoadBalancer.Servers]]
            URL = "http://127.0.0.1:80"
            Weight = 1
    ```
