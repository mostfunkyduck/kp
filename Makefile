REVISION=`git show | head -1 | awk '{print $$NF}' | cut -c 1-5`
HOSTNAME=`hostname`
DATE=`date -u +%Y-%m-%d-%H-%M`
BRANCH=`git branch 2>/dev/null | grep '\*' | sed "s/* //"`
RELEASE=0.1

.PHONY: test cscope install tidy fix lint testv1 testv2 vet

# default to having lint be a prereq to build

all: init lint nolint

# sets up the git hooks in the repo
init:
	git config core.hooksPath .githooks

# allow nolint from when bad stuff creeps in and needs a separate commit
nolint: test kp

kp: *.go keepass/*.go keepass/*/*.go
	go build -gcflags "-N -I ." -ldflags "-X main.VersionRevision=$(REVISION) -X main.VersionBuildDate=$(DATE) -X main.VersionBuildTZ=UTC -X main.VersionBranch=$(BRANCH) -X main.VersionRelease=$(RELEASE) -X main.VersionHostname=$(HOSTNAME)" 

cscope:
	# this will add a local index called 'cscope.out' based on a collection of source files in 'cscope.files'
	# adding local source code
	find . -name "*.go" -print > cscope.files
	# running cscope, the -b and -k flags will keep things narrowly scoped
	cscope -b -k

modtidy:
	go mod tidy

# non blocking linter run that fixes mistakes
fix: modtidy
	./scripts/lint.sh fix

lint:
	./scripts/lint.sh || exit 1

install: kp
	cp ./kp /usr/local/bin/kp

# allow testing v1 and v2 separately or together
coveragecmd := -coverprofile coverage.out -coverpkg=./internal/commands,./keepass,./keepass/common

testv1:
	KPVERSION=1 go test ./internal/commands ./keepass/keepassv1 $(coveragecmd),./keepass/keepassv1

testv2:
	KPVERSION=2 go test ./internal/commands ./keepass/keepassv2 $(coveragecmd),./keepass/keepassv2

test: testv1 testv2

# quick command to vet the entire source tree
vet:
	go vet . ./internal/* ./keepass/keepassv1 ./keepass/keepassv2
