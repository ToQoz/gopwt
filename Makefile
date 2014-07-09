gopwt: *.go
	go build
test: gopwt
	./gopwt ./...
