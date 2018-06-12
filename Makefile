TESTPKG = ./...
VERBOSE_FLAG = $(if $(VERBOSE),-v)

gopwt:
	go build
test:
	go test -tags=integration $(VERBOSE_FLAG) $(TESTPKG)
coverage:
	GOPWT_OFF=1 go test -tags=integration -race -coverprofile=coverage.txt -covermode=atomic $(TESTPKG)
test-all: gopwt
	_misc/test-all
example: gopwt
	go test ./_example
readme: gopwt
	_misc/gen-readme.bash
op:
	_misc/gen-op.bash
gen: op readme
