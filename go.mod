module github.com/mostfunkyduck/kp

go 1.18

require (
	github.com/AlecAivazis/survey/v2 v2.1.1
	github.com/abiosoft/ishell v2.0.1-0.20190723053747-1b6ad7eb4d5e+incompatible
	github.com/abiosoft/readline v0.0.0-20180607040430-155bce2042db
	github.com/atotto/clipboard v0.1.4
	github.com/sethvargo/go-password v0.2.0
	github.com/tobischo/gokeepasslib/v3 v3.1.0
	zombiezen.com/go/sandpass v1.1.0
)

require (
	github.com/aead/argon2 v0.0.0-20180111183520-a87724528b07 // indirect
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/chzyer/logex v1.2.1 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/flynn-archive/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b // indirect
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/sys v0.0.0-20210917161153-d61c044b1678 // indirect
	golang.org/x/text v0.3.6 // indirect
)

replace github.com/abiosoft/ishell => ./vendor/github.com/abiosoft/ishell

replace zombiezen.com/go/sandpass => github.com/mostfunkyduck/sandpass v1.1.1-0.20200617090953-4e7550e75911
