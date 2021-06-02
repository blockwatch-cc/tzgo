# Contributing to tzGO

We welcome all contributions to tzGO, be it extensions, bug fixes, documentation, or examples. This document outlines some of the conventions we put in place to keep the work on tzGO tidy and fun.

Before you start, please **read** and **sign** our Contributor License Agreement and send it to license@blockwatch.cc:

- use the [Individual CLA](https://github.com/blockwatch-cc/CLA/blob/master/ICLA.pdf) if you are an independent developer
- use the [Corporate CLA](https://github.com/blockwatch-cc/CLA/blob/master/CCLA.pdf) if you work for a company

The CLA is meant to protect you and us from legal trouble.

If you need any help or mentoring getting started or making a PR, please ask on [Discord](https://discord.gg/D5e98Hw).


## Contribution flow

This is a rough outline of what our contributor's workflow looks like:

- Create a Git branch from where you want to base your work. This is usually master.
- Write code, add test cases (optional right now), and commit your work (see below for message format).
- Run tests and make sure all tests pass (optional right now).
- Push your changes to a branch in your fork of the repository and submit a pull request.
- Your PR will be reviewed by a maintainer, who may request some changes.
  * Once you've made changes, your PR must be re-reviewed and approved.
  * If the PR becomes out of date, you can use GitHub's 'update branch' button.
  * If there are conflicts, you can rebase (or merge) and resolve them locally. Then force push to your PR branch.
    You do not need to get re-review just for resolving conflicts, but you should request re-review if there are significant changes.
- A maintainer will test and merge your pull request.

Thanks for your contributions!

### Format of the commit message

We follow a rough convention for commit messages that is designed to answer two
questions: what changed and why. The subject line should feature the what and
the body of the commit should describe the why.

```
rpc: add Granada constants

Blocks were counting all rewards and deposits including endorsements,
which was confusing and unclear what part was going to the block baker.
This update changes the rewards and deposits fields to only count
amounts related to baking.
```

The format can be described more formally as follows:

```
<subsystem>: <what changed>
<BLANK LINE>
<why this change was made>
<BLANK LINE>
Signed-off-by: <Name> <email address>
```

The first line is the subject and should be no longer than 50 characters, the other lines should be wrapped at 72 characters (see [this blog post](https://preslav.me/2015/02/21/what-s-with-the-50-72-rule/) for why).

The body of the commit message should describe why the change was made and at a high level, how the code works.

### Signing off the Commit

The project uses [DCO check](https://github.com/probot/dco#how-it-works) and the commit message must contain a `Signed-off-by` line for [Developer Certificate of Origin](https://developercertificate.org/).

Use option `git commit -s -m 'This is my commit message'` to sign off your commits.
