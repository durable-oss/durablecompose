# Contributing to DurableCompose

Want to help improve DurableCompose? We welcome contributions from the community!

This page contains information about reporting issues as well as some tips and
guidelines useful to experienced open source contributors. Finally, make sure
you read our [community guidelines](#durablecompose-community-guidelines) before you
start participating.

## About Durable Programming

DurableCompose is a product of **Durable Programming**, which specializes in helping organizations maintain and modernize enterprise applications. We believe in:

- **Long-term sustainability**: Software that remains maintainable over years
- **Pragmatic solutions**: Addressing real problems with proven approaches
- **Quality and reliability**: Thorough testing and careful design
- **Community collaboration**: Open development while supporting commercial viability

## Topics

- [Contributing to DurableCompose](#contributing-to-durablecompose)
  - [About Durable Programming](#about-durable-programming)
  - [Topics](#topics)
  - [Reporting security issues](#reporting-security-issues)
  - [Reporting other issues](#reporting-other-issues)
  - [Quick contribution tips and guidelines](#quick-contribution-tips-and-guidelines)
    - [Pull requests are always welcome](#pull-requests-are-always-welcome)
    - [Talking to other contributors](#talking-to-other-contributors)
    - [Conventions](#conventions)
    - [Merge approval](#merge-approval)
    - [Sign your work](#sign-your-work)
    - [How can I become a maintainer?](#how-can-i-become-a-maintainer)
  - [DurableCompose community guidelines](#durablecompose-community-guidelines)
  - [Coding Style](#coding-style)

## Reporting security issues

We take security seriously. If you discover a security issue, please bring it to our attention right away!

Please **DO NOT** file a public issue. Instead, send your report privately to
[security@durableprogramming.com](mailto:security@durableprogramming.com).

Security reports are greatly appreciated and we will publicly thank you for them.


## Reporting other issues

A great way to contribute to the project is to send a detailed report when you
encounter an issue. We always appreciate a well-written, thorough bug report,
and will thank you for it!

Check that [our issue database](https://github.com/durable_oss/durablecompose/issues)
doesn't already include that problem or suggestion before submitting an issue.
If you find a match, you can use the "subscribe" button to get notified of
updates. Do *not* leave random "+1" or "I have this too" comments, as they
only clutter the discussion, and don't help to resolve it. However, if you
have ways to reproduce the issue or have additional information that may help
resolve the issue, please leave a comment.

When reporting issues, always include:

* The output of `durablecompose version`.
* The output of `docker version`.
* The output of `docker context show`.
* The output of `docker info`.

Also, include the steps required to reproduce the problem if possible and
applicable. This information will help us review and fix your issue faster.
When sending lengthy log files, consider posting them as a gist
(https://gist.github.com).
Don't forget to remove sensitive data from your log files before posting (you
can replace those parts with "REDACTED").

_Note:_ 
Maintainers might request additional information to diagnose an issue,
if initial reporter doesn't answer within a reasonable delay (a few weeks),
issue will be closed.

## Quick contribution tips and guidelines

This section gives the experienced contributor some tips and guidelines.

### Pull requests are always welcome

Not sure if that typo is worth a pull request? Found a bug and know how to fix
it? Do it! We will appreciate it. Any significant change, like adding a backend,
should be documented as
[a GitHub issue](https://github.com/durable_oss/durablecompose/issues)
before anybody starts working on it.

We are always thrilled to receive pull requests. We do our best to process them
quickly. If your pull request is not accepted on the first try,
don't get discouraged!

### Talking to other contributors

<table class="tg">
  <col width="45%">
  <col width="65%">
  <tr>
    <td>GitHub Issues</td>
    <td>
      The primary place for discussion is our <a href="https://github.com/durable_oss/durablecompose/issues" target="_blank">GitHub issue tracker</a>. This is where we track bugs, feature requests, and discuss implementation details.
    </td>
  </tr>
  <tr>
    <td>GitHub Discussions</td>
    <td>
      For general questions, design discussions, and community support, use <a href="https://github.com/durable_oss/durablecompose/discussions" target="_blank">GitHub Discussions</a>.
    </td>
  </tr>
  <tr>
    <td>Website</td>
    <td>
      Visit <a href="https://durableprogramming.com" target="_blank">Durable Programming's website</a>
      to learn more about our philosophy and other projects.
    </td>
  </tr>
</table>


### Conventions

Fork the repository and make changes on your fork in a feature branch:

- If it's a bug fix branch, name it XXXX-something where XXXX is the number of
    the issue.
- If it's a feature branch, create an enhancement issue to announce
    your intentions, and name it XXXX-something where XXXX is the number of the
    issue.

Submit unit tests for your changes. Go has a great test framework built in; use
it! Take a look at existing tests for inspiration. Also, end-to-end tests are
available. Run the full test suite, both unit tests and e2e tests on your
branch before submitting a pull request. See [BUILDING.md](BUILDING.md) for
instructions to build and run tests.

Write clean code. Universally formatted code promotes ease of writing, reading,
and maintenance. Always run `gofmt -s -w file.go` on each changed file before
committing your changes. Most editors have plug-ins that do this automatically.

Pull request descriptions should be as clear as possible and include a reference
to all the issues that they address.

Commit messages must start with a capitalized and short summary (max. 50 chars)
written in the imperative, followed by an optional, more detailed explanatory
text which is separated from the summary by an empty line.

Code review comments may be added to your pull request. Discuss, then make the
suggested modifications and push additional commits to your feature branch. Post
a comment after pushing. New commits show up in the pull request automatically,
but the reviewers are notified only when you comment.

Pull requests must be cleanly rebased on top of the base branch without multiple branches
mixed into the PR.

**Git tip**: If your PR no longer merges cleanly, use `rebase master` in your
feature branch to update your pull request rather than `merge master`.

Before you make a pull request, squash your commits into logical units of work
using `git rebase -i` and `git push -f`. A logical unit of work is a consistent
set of patches that should be reviewed together: for example, upgrading the
version of a vendored dependency and taking advantage of its now available new
feature constitute two separate units of work. Implementing a new function and
calling it in another file constitute a single logical unit of work. The very
high majority of submissions should have a single commit, so if in doubt: squash
down to one.

After every commit, make sure the test suite passes. Include documentation
changes in the same pull request so that a revert would remove all traces of
the feature or fix.

Include an issue reference like `Closes #XXXX` or `Fixes #XXXX` in the pull
request description that closes an issue. Including references automatically
closes the issue on a merge.

Please do not add yourself to the `AUTHORS` file, as it is regenerated regularly
from the Git history.

Please see the [Coding Style](#coding-style) for further guidelines.

### Merge approval

Docker maintainers use LGTM (Looks Good To Me) in comments on the code review to
indicate acceptance.

A change requires at least 2 LGTMs from the maintainers of each
component affected.

For more details, see the [MAINTAINERS](MAINTAINERS) page.

### Sign your work

The sign-off is a simple line at the end of the explanation for the patch. Your
signature certifies that you wrote the patch or otherwise have the right to pass
it on as an open-source patch. The rules are pretty simple: if you can certify
the below (from [developercertificate.org](https://developercertificate.org/)):

```
Developer Certificate of Origin
Version 1.1

Copyright (C) 2004, 2006 The Linux Foundation and its contributors.
660 York Street, Suite 102,
San Francisco, CA 94110 USA

Everyone is permitted to copy and distribute verbatim copies of this
license document, but changing it is not allowed.

Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

Then you just add a line to every git commit message:

    Signed-off-by: Joe Smith <joe.smith@email.com>

Use your real name (sorry, no pseudonyms or anonymous contributions.)

If you set your `user.name` and `user.email` git configs, you can sign your
commit automatically with `git commit -s`.

### How can I become a maintainer?

We welcome dedicated contributors to join as maintainers. To become a maintainer:

1. Make consistent, quality contributions over time
2. Demonstrate understanding of the project's goals and philosophy
3. Show good judgment in code reviews and discussions
4. Be available to help maintain the project

Being a maintainer is a time investment. Make sure you will have time to make yourself available. You don't have to be a maintainer to make a difference on the project!

## DurableCompose community guidelines

We want to keep the DurableCompose community welcoming, collaborative, and productive. To help with this, we've established guidelines for the community:

* Be nice: Be courteous, respectful and polite to fellow community members:
  no regional, racial, gender or other abuse will be tolerated. We like
  nice people way better than mean ones!

* Encourage diversity and participation: Make everyone in our community feel
  welcome, regardless of their background and the extent of their
  contributions, and do everything possible to encourage participation in
  our community.

* Keep it legal: Basically, don't get us in trouble. Share only content that
  you own, do not share private or sensitive information, and don't break
  the law.

* Stay on topic: Make sure that you are posting to the correct channel and
  avoid off-topic discussions. Remember when you update an issue or respond
  to an email you are potentially sending it to a large number of people. Please
  consider this before you update. Also, remember that nobody likes spam.

* Don't send emails to the maintainers: There's no need to send emails to the
  maintainers to ask them to investigate an issue or to take a look at a
  pull request. Instead of sending an email, GitHub mentions should be
  used to ping maintainers to review a pull request, a proposal or an
  issue.

## Coding Style

Unless explicitly stated, we follow all coding guidelines from the Go
community. While some of these standards may seem arbitrary, they somehow seem
to result in a solid, consistent codebase.

It is possible that the code base does not currently comply with these
guidelines. We are not looking for a massive PR that fixes this, since that
goes against the spirit of the guidelines. All new contributors should make their
best effort to clean up and make the code base better than they left it.
Obviously, apply your best judgement. Remember, the goal here is to make the
code base easier for humans to navigate and understand. Always keep that in
mind when nudging others to comply.

The rules:

1. All code should be formatted with `gofmt -s`.
2. All code should pass the default levels of
   [`golint`](https://github.com/golang/lint).
3. All code should follow the guidelines covered in [Effective
   Go](https://go.dev/doc/effective_go) and [Go Code Review
   Comments](https://go.dev/wiki/CodeReviewComments).
4. Include code comments. Tell us the why, the history and the context.
5. Document _all_ declarations and methods, even private ones. Declare
   expectations, caveats and anything else that may be important. If a type
   gets exported, having the comments already there will ensure it's ready.
6. Variable name length should be proportional to its context and no longer.
   `noCommaALongVariableNameLikeThisIsNotMoreClearWhenASimpleCommentWouldDo`.
   In practice, short methods will have short variable names and globals will
   have longer names.
7. No underscores in package names. If you need a compound name, step back,
   and re-examine why you need a compound name. If you still think you need a
   compound name, lose the underscore.
8. No utils or helpers packages. If a function is not general enough to
   warrant its own package, it has not been written generally enough to be a
   part of a util package. Just leave it unexported and well-documented.
9. All tests should run with `go test` and outside tooling should not be
   required. No, we don't need another unit testing framework. Assertion
   packages are acceptable if they provide _real_ incremental value.
10. Even though we call these "rules" above, they are actually just
    guidelines. Since you've read all the rules, you now know that.

If you are having trouble getting into the mood of idiomatic Go, we recommend
reading through [Effective Go](https://go.dev/doc/effective_go). The
[Go Blog](https://go.dev/blog/) is also a great resource. Drinking the
kool-aid is a lot easier than going thirsty.
