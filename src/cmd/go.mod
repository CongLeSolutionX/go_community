module cmd

go 1.16

require (
	github.com/google/pprof v0.0.0-20201203190320-1bf35d6f28c2
	golang.org/x/arch v0.0.0-20201008161808-52c3e6f60cff
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/mod v0.4.1
	golang.org/x/tools v0.1.1-0.20210218235257-f47cb783b141
)

replace golang.org/x/tools => ../../../src/golang.org/x/tools
