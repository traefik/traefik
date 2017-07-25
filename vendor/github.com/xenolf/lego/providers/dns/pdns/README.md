## PowerDNS provider

Tested and confirmed to work with PowerDNS authoratative server 3.4.8 and 4.0.1. Refer to [PowerDNS documentation](https://doc.powerdns.com/md/httpapi/README/) instructions on how to enable the built-in API interface.

PowerDNS Notes:
- PowerDNS API does not currently support SSL, therefore you should take care to ensure that traffic between lego and the PowerDNS API is over a trusted network, VPN etc.
- In order to have the SOA serial automatically increment each time the `_acme-challenge` record is added/modified via the API, set `SOA-API-EDIT` to `INCEPTION-INCREMENT` for the zone in the `domainmetadata` table
