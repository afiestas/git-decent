# Git-Decent
Git-Decent is a small tool designed to help you, night owls, maintain appearances while working during unconventional hours.

## Problem Statement
Some individuals prefer to work during the night or other quiet hours, such as weekends.
However, this can create pressure on the rest of the team to work during those hours or
to be around to check pull requests and commits outside of their regular working schedule.

## Solution
Git-Decent aims to alleviate this burden for everyone involved.
With this tool, committers can make their commits at any time, and Git-Decent will automatically
adjust the commit timestamps to reflect a more "decent" datetime. This way, the rest of the team can
enjoy their free time without feeling obligated to be available or work outside their preferred schedule.

## Configuration
The configuration for Git-Decent is simple and intuitive. By using a configuration file, you can define
the desired "decent time slots" for commits.

Here's an example of how the configuration file (git-decent.ini) would look:

```ini
[Decent]
Mon = 9-5
Tue = 9-5
Wed = 9-5
```
In this example, commits made outside of the specified days and time slots will be automatically amended
to fit the desired "decent" time slots.

## Commit Amendment Example
Suppose a commit is made during off-hours on a weekend, such as Saturday at 02:00. Git-Decent will amend the
commit to have a datetime corresponding to the next available "decent" time slot, such as Monday at 09:XX:XX.
The specific time within the time slot will be somewhat randomly selected.

Then, if another commit is made on Saturday at 03:30 it will be amended with a datetime that falls after the
last amended commit but not too far into Monday. For example, it could be set to 09:50 or a similar time.

## Privacy Considerations
It is important to note that Git-Decent is not designed to preserve privacy. Its purpose is solely to make
your working time less conspicuous to others.

## Pre-Push Hook
Git-Decent also provides a pre-push hook to prevent the pushing of commits made in the future.
This behavior can be configured on a per-remote-branch basis. For instance, you may want to allow pushing
commits to your personal work-in-progress branch (myWIPBranch), as doing so will not trigger continuous
integration systems or send messages to the team's chat.

Feel free to contribute to Git-Decent and make your nocturnal coding sessions a bit more "decent"!