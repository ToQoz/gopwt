TESTPKG = ./...
VERBOSE_FLAG = $(if $(VERBOSE),-v)

gopwt:
	go build
test:
	GOPWT_DEBUG=1 go test $(VERBOSE_FLAG) $(TESTPKG)
	cd _integrationtest && ./test.sh
test-all: gopwt
	_misc/test-all
example: gopwt
	go test ./_example
readme: gopwt
	_misc/gen-readme.bash
op:
	_misc/gen-op.bash
gen: op readme
