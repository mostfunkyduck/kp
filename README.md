# kp
This project is a reimplementation of [kpcli](http://kpcli.sourceforge.net/) with a few additional features thrown in.  It provides a shell-like interface for navigating a keepass database and manipulating entries.

Currently, a full reimplementation of all of kpcli's features is still under development, but other features, such as `search` have been added.

keepassv2+ has also been implemented, but is still very incomplete - treat it as alpha software.  It is also advisable to not use v2 at all outside of development as it's incomplete and may have edge cases where corruption can occur.

keepassv1, however, is stable and fit for use
# usage

## command line
```
> ./kp -help
Usage of ./kp:
  -db string
        the db to open
  -key string
        a key file to use to unlock the db
  -kpversion int
        which version of keepass to use (1 or 2) (default 1)
  -n string
        execute a given command and exit
  -version
        print version and exit
```

## program shell
```
/ > help

Commands:
  attach       attach <get|show|delete> <entry> <filesystem location>
  cd           cd <path>
  clear        clear the screen
  edit         edit <entry>
  exit         exit the program
  help         display help
  ls           ls [path]
  mkdir        mkdir <group name>
  mv           mv <soruce> <destination>
  new          new <path>
  pwd          pwd
  rm           rm <entry>
  save         save
  saveas       saveas <file.kdb> [file.key]
  search       search <term>
  select       select [-f] <entry>
  show         show [-f] <entry>
  version      version
  xp           xp <entry>
  xu           xu
  xw           xw
  xx           xx
```
Running a command with the argument `help` will display a more detailed usage message
```
/ > attach help

manages the attachment for a given entry

Commands:
  create       attach create <entry> <name> <filesystem location>
  details      attach details <entry>
  get          attach get <entry> <filesystem location>
```

# Building
Running `make` will deploy the git hooks, lint the repo, run all the tests, then compile `kp`.  To deploy to `/usr/local/bin`, run `sudo make install`
```
> make
git config core.hooksPath .githooks
./scripts/lint.sh || exit 1
KPVERSION=1 go test . ./keepass/keepassv1 -coverprofile coverage.out -coverpkg=.,./keepass,./keepass/common,./keepass/keepassv1
ok      github.com/mostfunkyduck/kp     0.093s  coverage: 36.6% of statements in ., ./keepass, ./keepass/common, ./keepass/keepassv1
ok      github.com/mostfunkyduck/kp/keepass/keepassv1   0.799s  coverage: 19.3% of statements in ., ./keepass, ./keepass/common, ./keepass/keepassv1
KPVERSION=2 go test . ./keepass/keepassv2 -coverprofile coverage.out -coverpkg=.,./keepass,./keepass/common,./keepass/keepassv2
ok      github.com/mostfunkyduck/kp     0.091s  coverage: 41.5% of statements in ., ./keepass, ./keepass/common, ./keepass/keepassv2
ok      github.com/mostfunkyduck/kp/keepass/keepassv2   0.087s  coverage: 25.0% of statements in ., ./keepass, ./keepass/common, ./keepass/keepassv2
go build -gcflags "-N -I ." -ldflags "-X main.VersionRevision=`git show | head -1 | awk '{print $NF}' | cut -c 1-5` -X main.VersionBuildDate=`date -u +%Y-%m-%d-%H-%M` -X main.VersionBuildTZ=UTC -X main.VersionBranch=`git branch 2>/dev/null | grep '\*' | sed "s/* //"` -X main.VersionRelease=0.1 -X main.VersionHostname=`hostname`"
```

## Overview
There are two main components, the shell and the libraries that interact with the database directly.  The shell interfaces with the database through those abstractionsso that the actual logic is the same for v1 and v2.  The shell works by having individual files for each command which are strung together in `main.go`.

# Future plans
In priority order:

1. finish v2 by adding feature to add and remove fields as well as throwing in some more testing
1. fill in test coverage
1. implement for other systems, such as `pass`, because why not?
