# Building

## Dealing with Glide issues
When merging in updates from the forked Traefik repo the Glide lock file can have diverged
far enough to make starting from scratch a better option. If necessary accept the glide.yaml
and glide.lock files from the upstream repo and readd dependencies required for the audit-tap
to glide.yaml and rerun script/glide up.

```
- package: github.com/Shopify/sarama
  version: v1.12.0
- package: github.com/assembla/cony
  version: v0.3.2
- package: github.com/streadway/amqp
- package: github.com/beeker1121/goque
  version: v2.0.1
- package: github.com/syndtr/goleveldb
  subpackages:
  - leveldb
- package: gopkg.in/beevik/etree.v0
  version: v0
```