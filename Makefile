TESTPKG = ./...
VERBOSE_FLAG = $(if $(VERBOSE),-v)

test:
	go test $(VERBOSE_FLAG) $(TESTPKG)
test-all: gopwt
	_misc/test-all
example: gopwt
	go test ./_example
readme: gopwt
	_misc/gen
