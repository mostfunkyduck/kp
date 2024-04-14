# kp
This project is a reimplementation of [kpcli](http://kpcli.sourceforge.net/) with a few additional features thrown in.  It provides a shell-like interface for navigating a keepass database and manipulating entries.

This is a pure hobby project and is maintained when I feel up to it. To use it, install go >= 1.21 and run:

```sh
make test # run tests
make kp # build binary

./kp -h # print help

./kp -db /path/to/keepass.kdb # connect to a database
```
