REVISION=`git show | head -1 | awk '{print $$NF}' | cut -c 1-5`
HOSTNAME=`hostname`
DATE=`date -u +%Y-%m-%d-%H-%M`
BRANCH=`git branch 2>/dev/null | grep '\*' | sed "s/* //"`
RELEASE=0.1

.PHONY: test cscope goimports install

all: test kp
kp: *.go keepass/*.go keepass/*/*.go
	go build -gcflags "-N -I ." -ldflags "-X main.VersionRevision=$(REVISION) -X main.VersionBuildDate=$(DATE) -X main.VersionBuildTZ=UTC -X main.VersionBranch=$(BRANCH) -X main.VersionRelease=$(RELEASE) -X main.VersionHostname=$(HOSTNAME)" 

cscope:
	# this will add a local index called 'cscope.out' based on a collection of source files in 'cscope.files'
	# adding local source code
	find . -name "*.go" -print > cscope.files
	# running cscope, the -b and -k flags will keep things narrowly scoped
	cscope -b -k

goimports:
	goimports -w *.go
	goimports -w ./keepass/*.go
	goimports -w ./keepass/*/*.go

gofmt:
	go fmt
	go fmt ./keepass
	go fmt ./keepass/tests
	go fmt ./keepass/common
	go fmt ./keepass/keepassv1
	go fmt ./keepass/keepassv2

install: kp
	cp ./kp /usr/local/bin/kp

test:
	go test . ./keepass/keepassv2 ./keepass/keepassv1 -coverprofile coverage.out -coverpkg=.,./keepass,./keepass/keepassv2,./keepass/common,./keepass/keepassv1
