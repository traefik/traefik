---
layout: "intro"
page_title: "Consul vs. Custom Solutions"
sidebar_current: "vs-other-custom"
description: |-
  As a codebase grows, a monolithic app often evolves into a Service Oriented Architecture (SOA). A universal pain point for SOA is service discovery and configuration. In many cases, this leads to organizations building home grown solutions. It is an undisputed fact that distributed systems are hard; building one is error-prone and time-consuming. Most systems cut corners by introducing single points of failure such as a single Redis or RDBMS to maintain cluster state. These solutions may work in the short term, but they are rarely fault tolerant or scalable. Besides these limitations, they require time and resources to build and maintain.
---

# Consul vs. Custom Solutions

As a codebase grows, a monolithic app often evolves into a Service Oriented
Architecture (SOA). A universal pain point for SOA is service discovery and
configuration. In many cases, this leads to organizations building home grown
solutions. It is an undisputed fact that distributed systems are hard; building
one is error-prone and time-consuming. Most systems cut corners by introducing
single points of failure such as a single Redis or RDBMS to maintain cluster
state. These solutions may work in the short term, but they are rarely fault
tolerant or scalable. Besides these limitations, they require time and resources
to build and maintain.

Consul provides the core set of features needed by a SOA out of the box. By
using Consul, organizations can leverage open source work to reduce the time
and effort spent re-inventing the wheel and can focus instead on their business
applications.

Consul is built on well-cited research and is designed with the constraints of
distributed systems in mind. At every step, Consul takes efforts to provide a
robust and scalable solution for organizations of any size.
