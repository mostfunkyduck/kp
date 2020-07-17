module github.com/mostfunkyduck/kp

go 1.14

require (
	github.com/abiosoft/ishell v2.0.1-0.20190723053747-1b6ad7eb4d5e+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db
	github.com/atotto/clipboard v0.1.2
	github.com/chzyer/logex v1.1.10 // indirect
	github.com/chzyer/test v0.0.0-20180213035817-a1ea475d72b1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.9.0 // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/sethvargo/go-password v0.2.0
	github.com/stretchr/testify v1.6.1 // indirect
	github.com/tobischo/gokeepasslib/v3 v3.1.0
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sys v0.0.0-20200625212154-ddb9806d33ae // indirect
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	zombiezen.com/go/sandpass v1.1.0
)

replace zombiezen.com/go/sandpass => github.com/mostfunkyduck/sandpass v1.1.1-0.20200617090953-4e7550e75911

replace github.com/atotto/clipboard => github.com/mostfunkyduck/clipboard v0.1.3-0.20190428164314-bdea50a7aaf0
