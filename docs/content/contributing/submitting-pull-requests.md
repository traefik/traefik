---
title: "Traefik Pull Requests Documentation"
description: "Looking to contribute to Traefik Proxy? This guide will show you the guidelines for submitting a PR in our contributors guide repository."
---

# Before You Submit a Pull Request

This guide is for contributors who already have a pull request to submit.
If you are looking for information on setting up your developer environment
and creating code to contribute to Traefik Proxy or related projects,
see the [development guide](https://docs.traefik.io/contributing/building-testing/).

Looking for a way to contribute to Traefik Proxy?
Check out this list of [Priority Issues](https://github.com/traefik/traefik/labels/contributor%2Fwanted),
the [Good First Issue](https://github.com/traefik/traefik/labels/contributor%2Fgood-first-issue) list,
or the list of [confirmed bugs](https://github.com/traefik/traefik/labels/kind%2Fbug%2Fconfirmed) waiting to be remedied.

## How We Prioritize

We wish we could review every pull request right away.
Unfortunately, our team has to prioritize pull requests (PRs) for review
(but we are welcoming new [maintainers](https://github.com/traefik/traefik/blob/master/docs/content/contributing/maintainers-guidelines.md) to speed this up,
if you are interested, check it out and apply).

The PRs we are able to handle fastest are:

* Documentation updates.
* Bug fixes.
* Enhancements and Features with a `contributor/wanted` tag.

PRs that take more time to address include:

* Enhancements or Features without the `contributor/wanted` tag.

If you have an idea for an enhancement or feature that you would like to build,
[create an issue](https://github.com/traefik/traefik/issues/new/choose) for it first
and tell us you are interested in writing the PR.
If an issue already exists, definitely comment on it to tell us you are interested in creating a PR.

This will allow us to communicate directly and let you know if it is something we would accept.

It also allows us to make sure you have all the information you need during the design phase
so that it can be reviewed and merged quickly.

Read more about the [Triage process](https://github.com/traefik/contributors-guide/blob/master/issue_triage.md) in the docs.

## The Pull Request Submit Process

Merging a PR requires the following steps to be completed before it is merged automatically.

* Make sure your pull request adheres to our best practices. These include:
    * [Following project conventions](https://github.com/traefik/traefik/blob/master/docs/content/contributing/maintainers-guidelines.md); including using the PR Template.
    * Make small pull requests.
    * Solve only one problem at a time.
    * Comment thoroughly.
    * Do not open the PR from an organization repository.
    * Keep "allows edit from maintainer" checked.
    * Use semantic line breaks for documentation.
    * Ensure your PR is not a draft. We do not review drafts, but do answer questions and confer with developers on them as needed.
* Pass the validation check.
* Pass all tests.
* Receive 3 approving reviews maintainers.

## Pull Request Review Cycle

Learn about our [Triage Process](https://github.com/traefik/contributors-guide/blob/master/issue_triage.md),
in short, it looks like this:

* We triage every new PR or comment before entering it into the review process.
    * We ensure that all prerequisites for review have been met.
    * We check to make sure the use case meets our needs.
    * We assign reviewers.
* Design Review.
    * This takes longer than other parts of the process.
    * We review that there are no obvious conflicts with our codebase.
* Code Review.
    * We review the code in-depth and run tests.
    * We may ask for changes here.
    * During code review, we ask that you be reasonably responsive,
      if a PR languishes in code review it is at risk of rejection,
      or we may take ownership of the PR and the contributor will become a co-author.
* Merge.
    * Success!

!!! note

    Occasionally, we may freeze our codebase when working towards a specific feature or goal that could impact other development.
    During this time, your pull request could remain unmerged while the release work is completed.

## Run Local Verifications

You must run these local verifications before you submit your pull request to predict the pass or failure of continuous integration.
Your PR will not be reviewed until these are green on the CI.

* `make validate`
* `make pull-images`
* `make test`

## The Testing and Merge Workflow

Pull Requests are managed by the bot [Myrmica Lobicornis](https://github.com/traefik/lobicornis).
This bot is responsible for verifying GitHub Checks (CI, Tests, etc), mergability, and minimum reviews.
In addition, it rebases or merges with the base PR branch if needed.
It performs several other housekeeping items
and you can read more about those on the [README](https://github.com/traefik/lobicornis) for Lobicornis.

The maintainer giving the final LGTM must add the `status/3-needs-merge` label to trigger the merge bot.

By default, a squash-rebase merge will be carried out.

The status `status/4-merge-in-progress` is only used by the bot.

If the bot is not able to perform the merge, the label `bot/need-human-merge` is added.
In such a situation, solve the conflicts/CI/... and then remove the label `bot/need-human-merge`.

To prevent the bot from automatically merging a PR, add the label `bot/no-merge`.

The label `bot/light-review` decreases the number of required LGTM from 3 to 1.

This label can be used when:

* Updating a dependency.
* Merging branches back into the next version branch.
* Submitting minor documentation changes.
* Submitting changelog PRs.

## Why Was My Pull Request Closed?

Traefik Proxy is made by the community for the community,
as such the goal is to engage the community to make Traefik the best reverse proxy available.
Part of this goal is maintaining a lean codebase and ensuring code velocity.
unfortunately, this means that sometimes we will not be able to merge a pull request.

Because we respect the work you did, you will always be told why we are closing your pull request.
If you do not agree with our decision, do not worry; closed pull requests are effortless to recreate,
and little work is lost by closing a pull request that subsequently needs to be reopened.

Your pull request might be closed if:

* Your PR's design conflicts with our existing codebase in such a way that merging is not an option
  and the work needed to make your pull request usable is too high.
    * To prevent this, make sure you created an issue first
      and think about including Traefik Proxy maintainers in your design phase to minimize conflicts.
* Your PR is for an enhancement or feature that we will not use.
    * Please remember to create an issue for any pull request **before** you create a PR
      to ensure that your goal is something we can merge and that you have any design insight you might need from the team.
* Your PR has been waiting for feedback from the contributor for over 90 days.

## Why is My Pull Request Not Getting Reviewed

A few factors affect how long your pull request might wait for review.

We must prioritize which PRs we focus on.
Our first priority is PRs we have identified as having high community engagement and broad applicability.
We put our top priorities on our roadmap, and you can identify them by the `contributor/wanted` tag.
These PRs will enter our review process the fastest.

Our second priority is bug fixes.
Especially for bugs that have already been tagged with `bug/confirmed`.
These reviews enter the process quickly.

If your PR does not meet the criteria above,
it will take longer for us to review, as any PRs that do meet the criteria above will be prioritized.

Additionally, during the last few weeks of a milestone, we stop reviewing PRs to reduce churn and stabilize.
We will resume after the release.

The second major reason that we deprioritize your PR is that you are not following best practices.

The most common failures to follow best practices are:

* You did not create an issue for the PR you wish to make.
  If you do not create an issue before submitting your PR,
  we will not be able to answer any design questions and let you know how likely your PR is to be merged.
* You created pull requests that are too large to review.
    * Break your pull requests up.
      If you can extract whole ideas from your pull request and send those as pull requests of their own,
      you should do that instead.
      It is better to have many pull requests addressing one thing than one pull request addressing many things.
    * Traefik Proxy is a fast-moving codebase — lock in your changes ASAP with your small pull request,
      and make merges be someone else's problem.
      We want every pull request to be useful on its own,
      so use your best judgment on what should be a pull request vs. a commit.
* You did not comment well.
    * Comment everything.

Please remember that we are working internationally, cross-culturally, and with different use-cases.
Your reviewer will not intuitively understand the problem the same way you do or solve it the same way you would.
This is why every change you make must be explained, and your strategy for coding must also be explained.

* Your tests were inadequate or absent.
    * If you do not know how to test your PR, please ask!
      We will be happy to help you or suggest appropriate test cases.

If you have already followed the best practices and your PR still has not received a response,
here are some things you can do to move the process along:

* If you have fixed all the issues from a review,
  remember to re-request a review (using the designated button) to let your reviewer know that you are ready.
  You can choose to comment with the changes you made.
* Ping `@tfny` if you have not been assigned to a reviewer.

For more information on best practices, try these links:

* [How to Write a Git Commit Message - Chris Beams](https://chris.beams.io/posts/git-commit/)
* [Distributed Git - Contributing to a Project (Commit Guidelines)](https://git-scm.com/book/en/v2/Distributed-Git-Contributing-to-a-Project)
* [What’s with the 50/72 rule? - Preslav Rachev](https://preslav.me/2015/02/21/what-s-with-the-50-72-rule/)
* [A Note About Git Commit Messages - Tim Pope](https://tbaggery.com/2008/04/19/a-note-about-git-commit-messages.html)

## It's OK to Push Back

Sometimes reviewers make mistakes.
It is OK to push back on changes your reviewer requested.
If you have a good reason for doing something a certain way, you are absolutely allowed to debate the merits of a requested change.
Both the reviewer and reviewee should strive to discuss these issues in a polite and respectful manner.

You might be overruled, but you might also prevail.
We are pretty reasonable people.

Another phenomenon of open-source projects (where anyone can comment on any issue) is the dog-pile -
your pull request gets so many comments from so many people it becomes hard to follow.
In this situation, you can ask the primary reviewer (assignee) whether they want you to fork a new pull request
to clear out all the comments.
You do not have to fix every issue raised by every person who feels like commenting,
but you should answer reasonable comments with an explanation.

## Common Sense and Courtesy

No document can take the place of common sense and good taste.
Use your best judgment, while you put a bit of thought into how your work can be made easier to review.
If you do these things, your pull requests will get merged with less friction.
