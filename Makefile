TESTPKG = ./...
VERBOSE_FLAG = $(if $(VERBOSE),-v)

gopwt:
	go build
test:
	go test $(VERBOSE_FLAG) $(TESTPKG)
test-integration: test
	cd _integrationtest && ./test.sh
example: gopwt
	go test ./_example
readme: gopwt
	_misc/gen-readme.bash
op:
	_misc/gen-op.bash
gen: op readme
