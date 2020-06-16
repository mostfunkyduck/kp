kp: *.go
	go build

cscope:
	# this will add a local index called 'cscope.out' based on a collection of source files in 'cscope.files'
	# adding local source code
	find . -name "*.go" -print > cscope.files
	# running cscope, the -b and -k flags will keep things narrowly scoped
	cscope -b -k
