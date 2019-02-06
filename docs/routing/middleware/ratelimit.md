# TODO -- RateLimit

Protection from Too Many Calls
{: .subtitle }

![RateLimit](../../img/middleware/ratelimit.png)

The RateLimit middleware ensures that services will receive a _fair_ number of requests, and allows you define what is fair.

## Configuration Example

??? example "Limit to 100 requests every 10 seconds (with a possible burst of 200)"

    ```toml
    [middlewares]
        [middlewares.fair-ratelimit.ratelimit]
            extractorfunc = "client.ip"
    
              [middlewares.fair-ratelimit.ratelimit.rateset1]
                period = "10s"
                average = 100
                burst = 200
    ```

??? example "Combine multiple limits"

    ```toml
    [middlewares]
        [middlewares.fair-ratelimit.ratelimit]
            extractorfunc = "client.ip"
    
              [middlewares.fair-ratelimit.ratelimit.rateset1]
                period = "10s"
                average = 100
                burst = 200

              [middlewares.fair-ratelimit.ratelimit.rateset2]
                period = "3s"
                average = 5
                burst = 10
    ```
    
    Here, an average of 5 requests every 3 seconds is allowed and an average of 100 requests every 10 seconds. These can "burst" up to 10 and 200 in each period, respectively. 

## Configuration Options

### extractorfunc
 
The `extractorfunc` option defines the strategy used to categorize requests.

The possible values are:

 * `request.host` categorizes requests based on the request host.
 * `client.ip` categorizes requests based on the client ip.
 * `request.header.ANY_HEADER` categorizes requests based on the provided `ANY_HEADER` value.

### ratelimit (multiple values)

You can combine multiple ratelimit. 
The ratelimit will trigger with the first reached limit.

Each ratelimit has 3 options, `period`, `average`, and `burst`.

The rate limit will allow an average of `average` requests every `period`, with a maximum of `burst` request on that period.

!!! note "Period Format"

    Period is to be given in a format understood by [time.ParseDuration](https://golang.org/pkg/time/#ParseDuration).
    