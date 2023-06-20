# Releases

## Versions

Below is a non-exhaustive list of versions and their maintenance status:

| Version | Release Date | Active Support     | Security Support | 
|---------|--------------|--------------------|------------------|
| 2.10    | Apr 24, 2023 | Yes                | Yes              |
| 2.9     | Oct 03, 2022 | Ended Apr 24, 2023 | No               |
| 2.8     | Jun 29, 2022 | Ended Oct 03, 2022 | No               |
| 2.7     | May 24, 2022 | Ended Jun 29, 2022 | No               |
| 2.6     | Jan 24, 2022 | Ended May 24, 2022 | No               |
| 2.5     | Aug 17, 2021 | Ended Jan 24, 2022 | No               |
| 2.4     | Jan 19, 2021 | Ended Aug 17, 2021 | No               |
| 2.3     | Sep 23, 2020 | Ended Jan 19, 2021 | No               |
| 2.2     | Mar 25, 2020 | Ended Sep 23, 2020 | No               |
| 2.1     | Dec 11, 2019 | Ended Mar 25, 2020 | No               |
| 2.0     | Sep 16, 2019 | Ended Dec 11, 2019 | No               |
| 1.7     | Sep 24, 2018 | Ended Dec 31, 2021 | Contact Support  |

??? example "Active Support / Security Support"

    **Active support**: receives any bug fixes.
    **Security support**: receives only critical bug and security fixes.

This page is maintained and updated periodically to reflect our roadmap and any decisions affecting the end of support for Traefik Proxy.

Please refer to our migration guides for specific instructions on upgrading between versions, an example is the [v1 to v2 migration guide](../migration/v1-to-v2.md).

!!! important "All target dates for end of support or feature removal announcements may be subject to change."

## Versioning Scheme

The Traefik Proxy project follows the [semantic versioning](https://semver.org/) scheme and maintains a separate branch for each minor version. The main branch always represents the next upcoming minor or major version.

And these are our guiding rules for version support:

- **Only the latest `minor`** will be on active support at any given time
- **The last `minor` after releasing a new `major`** will be supported for 1 year following the `major` release
- **Previous rules are subject to change** and in such cases an announcement will be made publicly, [here](https://traefik.io/blog/traefik-2-1-in-the-wild/) is an example extending v1.x branch support.
