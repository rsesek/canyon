# canyon: Split Big Things

Canyon is a tool that can be used to split a git branch with a large number of changes into smaller, logically organized branches. This is helpful when performing a large refactoring in a project that uses a patch-based code review system.

## Workflow

The recommended workflow is as follows. Prepare your large refactoring on a single branch. This makes it easy to push/pull from different machines to test various build configurations. Once this single branch is in a working state, run canyon to split it up into small, reviewable units. Then go to each branch, upload it for review, merge in origin/master as necessary, and land. Iterate until you are done.

Canyon splits files by looking for common file paths. This can be done in either one of two ways: shared parent directory paths, or a file named the same amongst the child directories (e.g. an [OWNERS file](http://dev.chromium.org/developers/owners-files)). You can also limit the depth of subdirectories to be traversed, and canyon will merge changes in further nested directories into the parents. Each branch that canyon creates is named `canyon/original-branch/split-path`.

### For Chromium

This tool was first written with the [Chromium project](http://dev.chromium.org) in mind. The recommended workflow for that project is:

    ... hack on your big branch ...
    $ canyon -depth=2 -split-by=file -split-by-file=OWNERS -message 'Fix up callsites to FooBar(), in {{.SplitDirectory}}.

    BUG=12345`
    $ git checkout canyon/big-branch/foo-dir
    $ git cl upload
    ... edit the description and set R=someone
    ... the OWNERS file has been conveniently cat'd into the log ...
    ......
    ... get review ...
    $ git cl dcommit

The `git cl status` command will be very helpful during the process, as might the xargs command below.

## Options

Canyon operates *on the current branch* and *in the current repository*, and there are no options to control this. There are other options to control behavior though:

* `-message=S` set the commit message for every split. You can use `{{.SplitDirectory}}` to have canyon substitute the split directory path.
* `-depth=N` will control the maximum subdirectory depth for making splits.
* `-split-by=[dir|file]` controls the splitting behavior. `dir` will split by common shared parent directory paths, up to `-depth`. `file` will split by directories (or their parents) containing a file named `-split-by-file=S`, up to `-depth`.
* `-upstream=B` sets the branch that is "upstream" (typically "origin/master"), and is used as the diff-base.

The branches that canyon makes are just vanilla git branches -- there is no magic to them. The commit message you supplied will have an amendment listing the changed files, and if using `-split-by=file`, the contents of the file.

If you need to re-run the process for any reason, the following command will blow away all the split branches, like it never happened:

    git branch | grep canyon/ORIGINAL-BRANCH | xargs git branch -D

## F.A.Q.

### This is still a really f@#k!ng manual process.

Not a question, but yes. It's even worse without canyon. Unfortunately the tooling in this area is really lacking.

### Can canyon automate the upload/submit process more?

Perhaps. If you're coming from Chromium, note that a goal of this tool is not specific to that project and that can be reused elsewhere.

### Can canyon split files up any other ways?

Just the one that you're about to write yourself!

### Is there a way to control the extra data canyon adds to the commit log when splitting?

No, not at this time.

### Why wasn't this written in Python?

Why are you asking such a silly question?
