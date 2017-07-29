# Annotations
Several components of the specification, like [Image Manifests](manifest.md) and [Descriptors](descriptor.md), feature an optional annotations property, whose format is common and defined in this section.

This property contains arbitrary metadata.

## Rules

* Annotations MUST be a key-value map where both the key and value MUST be strings.
* While the value MUST be present, it MAY be an empty string.
* Keys MUST be unique within this map, and best practice is to namespace the keys.
* Keys SHOULD be named using a reverse domain notation - e.g. `com.example.myKey`.
* The prefix `org.opencontainers` is reserved for keys defined in Open Container Initiative (OCI) specifications and MUST NOT be used by other specifications and extensions.
* Keys using the `org.opencontainers.image` namespace are reserved for use in the OCI Image Specification and MUST NOT be used by other specifications and extensions, including other OCI specifications.
* If there are no annotations then this property MUST either be absent or be an empty map.
* Consumers MUST NOT generate an error if they encounter an unknown annotation key.

## Pre-Defined Annotation Keys

This specification defines the following annotation keys, intended for but not limited to [image index](image-index.md) and image [manifest](manifest.md) authors:
* **org.opencontainers.image.created** date on which the image was built (string, date-time as defined by [RFC 3339](https://tools.ietf.org/html/rfc3339#section-5.6)).
* **org.opencontainers.image.authors** contact details of the people or organization responsible for the image (freeform string)
* **org.opencontainers.image.homepage** URL to find more information on the image (string, a URL with scheme HTTP or HTTPS)
* **org.opencontainers.image.documentation** URL to get documentation on the image (string, a URL with scheme HTTP or HTTPS)
* **org.opencontainers.image.source** URL to get source code for the binary files in the image (string, a URL with scheme HTTP or HTTPS)
* **org.opencontainers.image.version** [Semantic versioning-compatible](http://semver.org/) version of the packaged software.
* **org.opencontainers.image.revision** Source control revision identifier for packaged software.
* **org.opencontainers.image.vendor** Name of the distributing entity, organization or individual.
* **org.opencontainers.image.licenses** Comma-separated list of licenses under which contained software is distributed, in [SPDX Short identifier](https://spdx.org/licenses/) form.
* **org.opencontainers.image.ref.name** Name of the reference for a target (string). SHOULD only be considered valid when on descriptors on `index.json` within [image layout](image-layout.md).
