# Building

## Dealing with vendoring changes
As of Traefik v1.5 [dep](https://github.com/golang/dep "Go dep repository") is used as the dependency management tool replacing Glide. Vendoring updates can be made by running `dep update` (_ensure dep is installed first_)

```
[[constraint]]
  name = "github.com/Shopify/sarama"
  version = "v1.12.0"

[[constraint]]
  name = "github.com/assembla/cony"
  version = "v0.3.2"

[[constraint]]
  name = "github.com/streadway/amqp"
  branch = "master"

[[constraint]]
  name = "github.com/beeker1121/goque"
  version = "v2.0.1"

[[constraint]]
  name = "github.com/beevik/etree"
  version = "v1.0.0"
```