# TODO -- RateLimit

Protection from Too Many Calls
{: .subtitle }

## Old Content

Rate limiting can be configured per frontend.  
Multiple sets of rates can be added to each frontend, but the time periods must be unique.

```toml
[frontends]
    [frontends.frontend1]
      # ...
      [frontends.frontend1.ratelimit]
        extractorfunc = "client.ip"
          [frontends.frontend1.ratelimit.rateset.rateset1]
            period = "10s"
            average = 100
            burst = 200
          [frontends.frontend1.ratelimit.rateset.rateset2]
            period = "3s"
            average = 5
            burst = 10
```

In the above example, frontend1 is configured to limit requests by the client's ip address.  
An average of 5 requests every 3 seconds is allowed and an average of 100 requests every 10 seconds.  
These can "burst" up to 10 and 200 in each period respectively. 

Valid values for `extractorfunc` are:
  * `client.ip`
  * `request.host`
  * `request.header.<header name>`