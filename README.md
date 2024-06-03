Git-Decent is a small tool designed to help you, night owls 🦉, maintain appearances while working during unconventional hours.

Some individuals prefer to work during the night or other quiet hours, such as weekends.
However, this can create pressure on the rest of the team to work during those hours or
to be around to check pull requests and commits outside of their regular working schedule.

[![asciicast](https://asciinema.org/a/662392.svg)](https://asciinema.org/a/662392)

## Configuration
The configuration for git-decent is simple and intuitive. By using a configuration file, you can define the desired "decent time frames" for commits.

Here's an example of how the configuration seciton inside git config will look like:

```ini
[decent]
    Monday = 09:00/13:00, 14:00/17:00
    Tuesday = 09:00/13:00, 14:00/17:00
    Wednesday = 09:00/13:00, 14:00/17:00
    Thursday = 09:00/13:00, 14:00/17:00
    Friday = 09:00/13:00, 14:00/17:00
```
This is the default configuration which contains the typical 9 to 5 schedule.

## Commands
- **git decent**: Unpushed commits are amended if needed to fit the schedule
- **git decent amend**: Amend the last commit, if needed
- **git decent install**: Installs the pre-push and post-commit [1] hooks
- **git decent pre-psuh**: This is the hook that prevents pushes at undecent times
- **git decent post-commit**: This is the hook that automatically amends commits [1]


## Commit Amendment Example
Suppose a commit is made during off-hours on a weekend, such as Saturday at 02:00, git-decent will amend the commit to have a datetime corresponding to the next available "decent" time frame, which in this case is Monday from 09 to 13.

If another commit is done on Saturday, then it will be placed after the latest unpushed commit.

## Privacy Considerations
It is important to note that git-decent is not designed to preserve privacy. Its purpose is solely to make your working time less conspicuous to others.

## Pre-Push Hook
git-decent also provides a pre-push hook to prevent the pushing of commits made in the future.
It will also prevent pushes outside of decent time frames.

## Post-Commit hook
This commit will automatically amend the recently created commit. We are **abusing** the intent
of this hook which is just notification to do our business, so please be careful while using it
and use it under your responsability.

The commit tries really hard not to execute when it is not required (rebases, merges, cherry picks, etc).

Feel free to contribute to Git-Decent and make your nocturnal coding sessions a bit more "decent"!
