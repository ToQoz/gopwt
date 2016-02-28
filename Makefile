TESTPKG = ./...
VERBOSE_FLAG = $(if $(VERBOSE),-v)

gopwt: *.go
	go build
test: gopwt
	./gopwt $(VERBOSE_FLAG) $(TESTPKG)
