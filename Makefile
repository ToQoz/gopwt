TESTPKG = ./...
VERBOSE_FLAG = $(if $(VERBOSE),-v)

gopwt: *.go
	go build
test: gopwt
	./gopwt $(VERBOSE_FLAG) $(TESTPKG)
test-all: gopwt
	_misc/test-all
example: gopwt
	./gopwt ./_example
readme: gopwt
	_misc/gen
