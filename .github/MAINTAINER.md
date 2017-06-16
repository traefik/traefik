# Maintainers

## Labels

If we open/look an issue/PR, we must add a `kind/*` and an `area/*`.

### Contributor

* `contributor/need-more-information`: we need more information from the contributor in order to analyze a problem.
* `contributor/waiting-for-corrections`: we need the contributor to take actions in order to move forward with a PR. **(only for PR)**
* `contributor/needs-resolve-conflicts`: use it only when there is some conflicts (and an automatic rebase is not possible). **(only for PR)** _[bot, humans]_

### Kind

* `kind/enhancement`: a new or improved feature.
* `kind/question`: It's a question. **(only for issue)**
* `kind/proposal`: proposal PR/issues need a public debate.
  * _Proposal issues_ are design proposal that need to be refined with multiple contributors.
  * _Proposal PRs_ are technical prototypes that need to be refined with multiple contributors.

* `kind/bug/possible`: if we need to analyze to understand if it's a bug or not. **(only for issues)** _[bot only]_
* `kind/bug/confirmed`: we are sure, it's a bug. **(only for issues)**
* `kind/bug/fix`: it's a bug fix. **(only for PR)**

### Resolution

* `resolution/duplicate`: it's a duplicate issue/PR.
* `resolution/declined`: Rule #1 of open-source: no is temporary, yes is forever.
* `WIP`: Work In Progress. **(only for PR)**

### Platform

* `platform/windows`: Windows related.

### Area

* `area/provider`: related to all providers.
* `area/provider/boltdb`: Boltd DB related.
* `area/provider/consul`: Consul related.
* `area/provider/docker`: Docker and Swarm related.
* `area/provider/ecs`: ECS related.
* `area/provider/etcd`: Etcd related.
* `area/provider/eureka`: Eureka related.
* `area/provider/k8s`: Kubernetes related.
* `area/provider/marathon`: Marathon related.
* `area/provider/mesos`: Mesos related.
* `area/provider/rancher`: Rancher related.
* `area/provider/zk`: Zoo Keeper related.
* `area/middleware`: Middleware related.
* `area/acme`: ACME related.
* `area/authentication`: Authentication related.
* `area/api`: Traefik API related.
* `area/logs`: Traefik logs related.
* `area/sticky-session`: Sticky session related.
* `area/websocket`: WebSocket related.
* `area/webui`: Web UI related.
* `area/infrastructure`: related to CI or Traefik building scripts.
* `area/documentation`: regards improving/adding documentation.
* `area/cluster`: Traefik clustering related.

### Priority

* `priority/P0`: needs hot fix. **(only for issue)**
* `priority/P1`: need to be fixed in next release. **(only for issue)**
* `priority/P2`: need to be fixed in the future. **(only for issue)**
* `priority/P3`: maybe. **(only for issue)**

### PR size

* `size/S`: small PR. **(only for PR)** _[bot only]_
* `size/M`: medium PR. **(only for PR)** _[bot only]_
* `size/L`: Large PR. **(only for PR)** _[bot only]_

### Status - Workflow

The `status/*` labels represent the desired state in the workflow.

* `status/0-needs-triage`: all new issue or PR have this status. _[bot only]_
* `status/1-needs-design-review`: need a design review. **(only for PR)**
* `status/2-needs-review`: need a code/documentation review. **(only for PR)**
* `status/3-needs-merge`: ready to merge. **(only for PR)**

Note:
* The status `needs-design-review` is only used in complex/heavy PRs.
* From `1` to `2`: 2 design LGTM needed in comment.
* From `2` to `3`: PR need 2 approvals.
