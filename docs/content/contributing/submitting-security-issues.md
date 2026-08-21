---
title: "Traefik Security Documentation"
description: "Security is a key part of Traefik Proxy. Read the technical documentation to learn about security advisories, CVE, and how to report a vulnerability."
---

# Security

## Security Advisories

We strongly advise you to join our mailing list to be aware of the latest announcements from our security team.
You can subscribe by sending an email to security+subscribe@traefik.io or on [the online viewer](https://groups.google.com/a/traefik.io/forum/#!forum/security).

## CVE

Reported vulnerabilities can be found on
[cve.mitre.org](https://cve.mitre.org/cgi-bin/cvekey.cgi?keyword=traefik).

CVEs are only created for vulnerabilities affecting **Generally Available (GA) versions** of Traefik.
Vulnerabilities discovered in non-GA versions (release candidates, betas, early access, or development branches)
will be fixed without creating a CVE.

## Threat Model

Traefik is an edge router. Its security boundary sits between **untrusted network clients** and the
services it routes to. A report is a vulnerability when an **unprivileged, untrusted client crosses that
boundary**. Reports that start from the other side of it describe bugs, and we fix bugs.

This threat model applies to reports submitted on or after **1 September 2026**. It states positions we
already apply; publishing them is meant to save you the work of rediscovering them.

### What Traefik Trusts

- **The configuration, and every provider that supplies it**: static and dynamic configuration,
  Kubernetes objects, container labels, files, and KV stores. Anyone able to write configuration that
  Traefik reads already controls routing, and can redirect traffic, terminate TLS, or remove a
  middleware, by design.
- **Operators and cluster administrators.** A report requiring operator, cluster-admin, or equivalent
  privilege does not cross the boundary.
- **Trust the operator has declared.** Where the configuration says a client, a namespace or a
  forwarding header is trusted, Traefik trusts it. Such decisions are taken **once, at the entrypoint,
  before any middleware runs**, and are not re-litigated per middleware.
- **The internal configuration keyspace** is not a tenant isolation boundary on a shared instance.
  Hardening it is defence in depth.

### What Crosses the Boundary

With no configuration-write access and no operator privilege required:

- **Routing or matching bypass**: an untrusted request reaching a router, service or backend the
  configuration does not grant it.
- **Authentication or authorization that fails open** or is bypassable under a configuration a
  reasonable reading of the documentation would call correct.
- **Loss of confidentiality or integrity of proxied traffic**, including credential exposure and TLS or
  mTLS enforcement that does not hold.
- **Remote crash or unbounded resource consumption** reachable from unauthenticated requests.
- **Escalation across a boundary the configuration explicitly established.**

### Where the Line Falls in Practice

A boundary stated this briefly does not resolve a concrete report on its own. The surfaces where we
have already settled a position, each with the neighbouring variant we **do** treat as a vulnerability
and the CVEs that prove it, are on the [Security Decisions](./security-decisions.md) page. Check your
finding there before submitting.

## Handled as a Bug, Without an Advisory

Some reports describe real defects that we fix, often at high priority, but that do not receive an
advisory or a CVE, because they do not cross the boundary above. We say so explicitly rather than
leaving it implicit, and we will point at a specific entry in
[Security Decisions](./security-decisions.md) when we close a report on these grounds.

## Report a Vulnerability

We want to keep Traefik safe for everyone.
If you've discovered a security vulnerability in Traefik,
we appreciate your help in disclosing it to us in a responsible manner,
by creating a [security advisory](https://github.com/traefik/traefik/security/advisories).

## Code of Conduct for Vulnerability Submissions

We are committed to handling every legitimate report responsibly,
and we expect submitters to engage with our security team in a respectful and collaborative manner.

The following behaviors are **not acceptable** and will not be tolerated:

- **Threats** to publicly disclose the vulnerability if it is not fixed within a timeframe you set unilaterally.
- **Ultimatums** or pressure tactics intended to force a faster response than our normal triage and remediation process allows.
- **Demands** for payment, bug bounties, or any form of compensation in exchange for not disclosing the issue
  (Traefik does not operate a paid bug bounty program).
- **Aggressive, abusive, or disrespectful communication** with our security team.

Submitters who engage in any of the above may face the following consequences:

- The submitter **will not be credited** in the security advisory or any subsequent communication.
- The submitter's GitHub profile may be **reported to GitHub** for violation of platform terms of service.
- We may **decline to engage further** on the report, while still addressing the underlying issue if it is legitimate.

We take security seriously and act on legitimate reports as quickly as our resources allow.
Patience and constructive dialogue help us protect users effectively.

## Submission Quality Guidelines

We have been receiving an increasing number of low-quality vulnerability reports that are not actual security issues.
Many of these reports originate from AI/LLM tools and are submitted without any human validation or testing.
This wastes the time of our security team and delays the handling of legitimate vulnerabilities.

Before submitting a security advisory, you **must**:

- **Carefully test and validate** the vulnerability yourself before submitting.
  You must be able to demonstrate a working proof of concept with clear reproduction steps.
- **Understand the impact** of the vulnerability and explain how it can be exploited in a realistic scenario.
- **Verify that the issue is not a false positive**.
  Ensure the behavior you are reporting is actually a security concern and not expected behavior.
  Check it against the [Threat Model](#threat-model) above before submitting.
- **Submit one finding per report.**
  Bundling several unrelated issues into a single advisory makes each one slower to triage, and means a
  single closure decision has to cover claims that deserve different answers.
- **Disclose your use of AI tooling.**
  Using an AI assistant to find or write up an issue is acceptable, and saying so is required. State
  which tool you used and what you verified yourself. An undisclosed AI-generated report that turns out
  to be unvalidated is treated as a breach of these guidelines, not merely a low-quality report.

### Policy on AI-Generated Reports

Security reports that are **directly generated by AI/LLM tools without proper human validation** will be **closed immediately**.

Indicators of unvalidated AI-generated reports include (but are not limited to):

- No working proof of concept or reproduction steps.
- Generic or theoretical vulnerability descriptions with no evidence of actual testing.
- Misunderstanding of Traefik's architecture or threat model.
- Hallucinated code paths, configuration options, or behaviors that do not exist.

**Contributors who repeatedly submit low-quality or unvalidated reports may have their accounts blocked.**

We appreciate the work of security researchers who take the time to rigorously validate their findings.
Quality over quantity helps keep Traefik safe for everyone.
