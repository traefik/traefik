# Contribution Guidelines

## Pre-Development
- Look for an existing Github issue describing the bug you have found/feature request you would like to see getting implemented.
- If no issue exists and there is reason to believe that your (non-trivial) contribution might be subject to an up-front design discussion, file an issue first and propose your idea.

## Development
- Fork the repository.
- Create a feature branch (`git checkout -b my-new-feature master`).
- Commit your changes, preferring one commit per logical unit of work. Often times, this simply means having a single commit.
- If applicable, update the documentation in the [README file](README.md).
- In the vast majority of cases, you should add/amend a (regression) test for your bug fix/feature.
- Push your branch (`git push origin my-new-feature`).
- Create a new pull request.
- Address any comments your reviewer raises, pushing additional commits onto your branch along the way. In particular, refrain from amending/force-pushing until you receive an LGTM (Looks Good To Me) from your reviewer. This will allow for a better review experience.
