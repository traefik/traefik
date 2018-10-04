# Maintainers

## The team

* Emile Vauge [@emilevauge](https://github.com/emilevauge)
* Vincent Demeester [@vdemeester](https://github.com/vdemeester)
* Ed Robinson [@errm](https://github.com/errm)
* Daniel Tomcej [@dtomcej](https://github.com/dtomcej)
* Manuel Zapf [@SantoDE](https://github.com/SantoDE)
* Timo Reimann [@timoreimann](https://github.com/timoreimann)
* Ludovic Fernandez [@ldez](https://github.com/ldez)
* Julien Salleyron [@juliens](https://github.com/juliens)
* Nicolas Mengin [@nmengin](https://github.com/nmengin)
* Marco Jantke [@marco-jantke](https://github.com/marco-jantke)
* Michaël Matur [@mmatur](https://github.com/mmatur)
* Gérald Croës [@geraldcroes](https://github.com/geraldcroes)
* Jean-Baptiste Doumenjou [@jbdoumenjou](https://github.com/jbdoumenjou)
* Damien Duportal [@dduportal](https://github.com/dduportal)

## Contributions Daily Meeting

* 3 Maintainers should attend to a Contributions Daily Meeting where we sort and label new issues ([is:issue label:status/0-needs-triage](https://github.com/containous/traefik/issues?utf8=%E2%9C%93&q=is%3Aissue+label%3Astatus%2F0-needs-triage+)), and review every Pull Requests
* Every pull request should be checked during the Contributions Daily Meeting
   * Even if it’s already assigned
   * Even PR labelled with `contributor/waiting-for-corrections` or `contributor/waiting-for-feedback`
* Issues labeled with `priority/P0` and `priority/P1` should be assigned.
* Modifying an issue or a pull request (labels, assignees, milestone) is only possible:
   * During the Contributions Daily Meeting
   * By an assigned maintainer
   * In case of emergency, if a change proposal is approved by 2 other maintainers (on Slack, Discord, etc)

## PR review process:

* The status `needs-design-review` is only used in complex/heavy/tricky PRs.
* From `status/1-needs-design-review` to `status/2-needs-review`: 1 design LGTM in comment, by a senior maintainer, if needed.
* From `status/2-needs-review` to `status/3-needs-merge`: 3 LGTM by any maintainer.
* If needed, a specific maintainer familiar with a particular domain can be requested for the review.
* If a PR has been implemented in pair programming, one peer's LGTM goes into the review for free
* Amending someone else's pull request is authorized only in emergency, if a rebase is needed, or if the initial contributor is silent

We use [PRM](https://github.com/ldez/prm) to manage locally pull requests.


## Bots

### [Myrmica Lobicornis](https://github.com/containous/lobicornis/)

**Update and Merge Pull Request**

The maintainer giving the final LGTM must add the `status/3-needs-merge` label to trigger the merge bot.

By default, a squash-rebase merge will be carried out.
If you want to preserve commits you must add `bot/merge-method-rebase` before `status/3-needs-merge`.

The status `status/4-merge-in-progress` is only for the bot.

If the bot is not able to perform the merge, the label `bot/need-human-merge` is added.  
In this case you must solve conflicts/CI/... and after you only need to remove `bot/need-human-merge`.

A maintainer can add `bot/no-merge` on a PR if he want (temporarily) prevent a merge by the bot.

`bot/light-review` can be used to decrease required LGTM from 3 to 1 when:

- vendor updates from previously reviewed PRs
- merges branches into master
- prepare release


### [Myrmica Bibikoffi](https://github.com/containous/bibikoffi/)

* closes stale issues [cron]
    * use some criterion as number of days between creation, last update, labels, ...


### [Myrmica Aloba](https://github.com/containous/aloba)

**Manage GitHub labels**

* Add labels on new PR [GitHub WebHook]
* Add milestone to a new PR based on a branch version (1.4, 1.3, ...) [GitHub WebHook]
* Add and remove `contributor/waiting-for-corrections` label when a review request changes [GitHub WebHook]
* Weekly report of PR status on Slack (CaptainPR) [cron]


## Labels

If we open/look an issue/PR, we must add a `kind/*`, an `area/*` and a `status/*`.

### Contributor

* `contributor/need-more-information`: we need more information from the contributor in order to analyze a problem.
* `contributor/waiting-for-feedback`: we need the contributor to give us feedback.
* `contributor/waiting-for-corrections`: we need the contributor to take actions in order to move forward with a PR. **(only for PR)** _[bot, humans]_
* `contributor/needs-resolve-conflicts`: use it only when there is some conflicts (and an automatic rebase is not possible). **(only for PR)** _[bot, humans]_

### Kind

* `kind/enhancement`: a new or improved feature.
* `kind/question`: It's a question. **(only for issue)**
* `kind/proposal`: proposal PR/issues need a public debate.
  * _Proposal issues_ are design proposal that need to be refined with multiple contributors.
  * _Proposal PRs_ are technical prototypes that need to be refined with multiple contributors.

* `kind/bug/possible`: if we need to analyze to understand if it's a bug or not. **(only for issues)**
* `kind/bug/confirmed`: we are sure, it's a bug. **(only for issues)**
* `kind/bug/fix`: it's a bug fix. **(only for PR)**

### Resolution

* `resolution/duplicate`: it's a duplicate issue/PR.
* `resolution/declined`: Rule #1 of open-source: no is temporary, yes is forever.
* `WIP`: Work In Progress. **(only for PR)**

### Platform

* `platform/windows`: Windows related.

### Area

* `area/acme`: ACME related.
* `area/api`: Traefik API related.
* `area/authentication`: Authentication related.
* `area/cluster`: Traefik clustering related.
* `area/documentation`: regards improving/adding documentation.
* `area/infrastructure`: related to CI or Traefik building scripts.
* `area/healthcheck`: Health-check related.
* `area/logs`: Traefik logs related.
* `area/middleware`: Middleware related.
* `area/middleware/metrics`: Metrics related. (Prometheus, StatsD, ...)
* `area/middleware/tracing`: Tracing related. (Jaeger, Zipkin, ...)
* `area/oxy`: Oxy related.
* `area/provider`: related to all providers.
* `area/provider/boltdb`: Boltd DB related.
* `area/provider/consul`: Consul related.
* `area/provider/docker`: Docker and Swarm related.
* `area/provider/ecs`: ECS related.
* `area/provider/etcd`: Etcd related.
* `area/provider/eureka`: Eureka related.
* `area/provider/file`: file provider related.
* `area/provider/k8s`: Kubernetes related.
* `area/provider/kv`: KV related.
* `area/provider/marathon`: Marathon related.
* `area/provider/mesos`: Mesos related.
* `area/provider/rancher`: Rancher related.
* `area/provider/servicefabric`: Azure service fabric related.
* `area/provider/zk`: Zoo Keeper related.
* `area/rules`: Rules related.
* `area/server`: Server related.
* `area/sticky-session`: Sticky session related.
* `area/tls`: TLS related.
* `area/websocket`: WebSocket related.
* `area/webui`: Web UI related.

### Issues Priority

* `priority/P0`: needs hot fix.
* `priority/P1`: need to be fixed in next release.
* `priority/P2`: need to be fixed in the future.
* `priority/P3`: maybe.

### PR size

_Automatically set by a bot_

* `size/S`: small PR.
* `size/M`: medium PR.
* `size/L`: Large PR.

### Status - Workflow

The `status/*` labels represent the desired state in the workflow.

* `status/0-needs-triage`: all new issue or PR have this status. _[bot only]_
* `status/1-needs-design-review`: need a design review. **(only for PR)**
* `status/2-needs-review`: need a code/documentation review. **(only for PR)**
* `status/3-needs-merge`: ready to merge. **(only for PR)**
* `status/4-merge-in-progress`: merge in progress. _[bot only]_
