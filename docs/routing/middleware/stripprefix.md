# StripPrefix

Removing Prefixes From the Path Before Forwarding the Request
{: .subtitle }

# OldContent
 
Use a `*Strip` matcher if your backend listens on the root path (`/`) but should be routeable on a specific prefix.
For instance, `PathPrefixStrip: /products` would match `/products` but also `/products/shoes` and `/products/shirts`.  
Since the path is stripped prior to forwarding, your backend is expected to listen on `/`.  
If your backend is serving assets (e.g., images or Javascript files), chances are it must return properly constructed relative URLs.  
Continuing on the example, the backend should return `/products/shoes/image.png` (and not `/images.png` which Traefik would likely not be able to associate with the same backend).  
The `X-Forwarded-Prefix` header (available since Traefik 1.3) can be queried to build such URLs dynamically.
