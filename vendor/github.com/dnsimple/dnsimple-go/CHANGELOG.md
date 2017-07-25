# CHANGELOG

#### master

- NEW: Added support for the DNSSEC Beta (GH-58)

- CHANGED: Changed response types to not be exported (GH-54)
- CHANGED: Updated registrar URLs (GH-59)

#### Release 0.14.0

- NEW: Added support for Collaborators API (GH-48)
- NEW: Added support for ZoneRecord regions (GH-47)
- NEW: Added support for Domain Pushes API (GH-42)
- NEW: Added support for domains premium prices API (GH-53)

- CHANGED: Renamed `DomainTransferRequest.AuthInfo` to `AuthCode` (GH-46)
- CHANGED: Updated registration, transfer, renewal response payload (dnsimple/dnsimple-developer#111, GH-52).
- CHANGED: Normalize unique string identifiers to SID (dnsimple/dnsimple-developer#113)
- CHANGED: Update whois privacy setting for domain (dnsimple/dnsimple-developer#120)


#### Release 0.13.0

- NEW: Added support for Accounts API (GH-29)
- NEW: Added support for Services API (GH-30, GH-35)
- NEW: Added support for Certificates API (GH-31)
- NEW: Added support for Vanity name servers API (GH-34)
- NEW: Added support for delegation API (GH-32)
- NEW: Added support for Templates API (GH-36, GH-39)
- NEW: Added support for Template Records API (GH-37)
- NEW: Added support for Zone files API (GH-38)


#### Release 0.12.0

- CHANGED: Setting a custom user-agent no longer overrides the origina user-agent (GH-26)
- CHANGED: Renamed Contact#email_address to Contact#email (GH-27)


#### Release 0.11.0

- NEW: Added support for parsing ZoneRecord webhooks.
- NEW: Added support for listing options (GH-25).
- NEW: Added support for Template API (GH-21).


#### Release 0.10.0

Initial release.
