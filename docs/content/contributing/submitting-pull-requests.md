# Submitting Pull Requests

A Quick Guide for Efficient Contributions
{: .subtitle }

So you've decided to improve Traefik? 
Thank You! 
Now the last step is to submit your Pull Request in a way that makes sure it gets the attention it deserves.

Let's go through the classic pitfalls to make sure everything is right. 

## Title

The title must be short and descriptive. (~60 characters)

## Description

Follow the [pull request template](https://github.com/traefik/traefik/blob/master/.github/PULL_REQUEST_TEMPLATE.md) as much as possible.

Explain the conditions which led you to write this PR: give us context.
The context should lead to something, an idea or a problem that you’re facing.

Remain clear and concise.

Take time to polish the format of your message so we'll enjoy reading it and working on it.
Help the readers focus on what matters, and help them understand the structure of your message (see the [Github Markdown Syntax](https://help.github.com/articles/github-flavored-markdown)).

## PR Content

- Make it small.
- One feature per Pull Request.
- Write useful descriptions and titles.
- Avoid re-formatting code that is not on the path of your PR.
- Make sure the [code builds](building-testing.md).
- Make sure [all tests pass](building-testing.md).
- Add tests.
- Address review comments in terms of additional commits (and don't amend/squash existing ones unless the PR is trivial).

!!! note "Third-Party Dependencies"

    If a PR involves changes to third-party dependencies, the commits pertaining to the vendor folder and the manifest/lock file(s) should be committed separated.

!!! tip "10 Tips for Better Pull Requests"

    We enjoyed this article, maybe you will too! [10 tips for better pull requests](https://blog.ploeh.dk/2015/01/15/10-tips-for-better-pull-requests/).
