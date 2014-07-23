TESTPKG = ./...

gopwt: *.go
	go build
test: gopwt
	./gopwt -v $(TESTPKG)
