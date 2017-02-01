# parsy
POSIX-style command line argument parsing

## Motivation for this package:
Go's flag package is very strict and expects arguments after options ONLY
otherwise it treats everything after an argument as an argument e.g
```shell
go run main.go gotham --depth 10 --no-prompt true
```

instead discards all the initialized options on the CLI. This is a problem
that I've encountered quite a lot from users of drive who are used to Posix
style command line argument parsing.
With this package, that order is accepted and can be used alright.
