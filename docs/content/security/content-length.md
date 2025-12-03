---
title: "Content-Length"
description: "Enforce strict Content‑Length validation in Baqup by streaming or full buffering to prevent truncated or over‑long requests and responses. Read the technical documentation."
---

Baqup acts as a streaming proxy. By default, it checks each chunk of data against the `Content-Length` header as it passes it on to the backend or client. This live check blocks truncated or over‑long streams without holding the entire message.

If you need Baqup to read and verify the full body before any data moves on, add the [buffering middleware](../middlewares/http/buffering.md):

```yaml
http:
  middlewares:
    buffer-and-validate:
      buffering: {}
```

With buffering enabled, Baqup will:

- Read the entire request or response into memory.
- Compare the actual byte count to the `Content-Length` header.
- Reject the message if the counts do not match.

!!!warning 
    Buffering adds overhead. Every request and response is held in full before forwarding, which can increase memory use and latency. Use it when strict content validation is critical to your security posture.
