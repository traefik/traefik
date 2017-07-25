Custom Endpoint Example
===

This example provides examples on how you can provide custom endpoints, and logic to how endpoints are resolved by the SDK.

The example creates multiple clients with different endpoint configuraiton. From a custom endpoint resolver that wraps the defeault resolver so that any S3 service client created uses the custom endpoint, to how you can provide your own logic to a single service's endpoint resolving.
