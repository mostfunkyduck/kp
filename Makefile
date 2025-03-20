REVISION=`git show | head -1 | awk '{print $$NF}' | cut -c 1-5`
HOSTNAME=`hostname`
DATE=`date -u +%Y-%m-%d-%H-%M`
BRANCH=`git branch 2>/dev/null | grep '\*' | sed "s/* //"`
RELEASE=0.1.0

.PHONY: test cscope install tidy fix lint testv1 testv2 vet

# default to having lint be a prereq to build

all: init lint nolint

# sets up the git hooks in the repo
init:
	git config core.hooksPath .githooks

# allow nolint from when bad stuff creeps in and needs a separate commit
nolint: test kp

prepForDebug: kp
	test -d ./local_artifacts || mkdir ./local_artifacts

kp: *.go internal/*/*.go internal/backend/*/*.go
	go build -mod=readonly -gcflags "-N -I ." -ldflags "-X main.VersionRevision=$(REVISION) -X main.VersionBuildDate=$(DATE) -X main.VersionBuildTZ=UTC -X main.VersionBranch=$(BRANCH) -X main.VersionRelease=$(RELEASE) -X main.VersionHostname=$(HOSTNAME)" 

modtidy:
	go mod tidy

# non blocking linter run that fixes mistakes
fix: modtidy
	./scripts/lint.sh fix

lint:
	./scripts/lint.sh || exit 1

install: kp
	install ./kp $(HOME)/.local/bin
# allow testing v1 and v2 separately or together
coveragecmd := -coverprofile coverage.out -coverpkg=./internal/commands,./internal/backend/types,./internal/backend/common
internalpkgs := ./internal/commands
testv1:
	KPVERSION=1 go test -mod=readonly $(internalpkgs) ./internal/backend/keepassv1 $(coveragecmd)

testv2:
	KPVERSION=2 go test -mod=readonly $(internalpkgs) ./internal/backend/keepassv2 $(coveragecmd)

test: testv1 testv2

# quick command to vet the entire source tree, need to enumerate all targets because of linter pickiness
vet:
	go vet . ./internal/commands ./internal/backend/types ./internal/backend/keepassv1 ./internal/backend/keepassv2 ./internal/backend/common ./internal/backend/tests
