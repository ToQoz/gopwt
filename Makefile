.DEFAULT_GOAL := test

TESTPKG = ./...
VERBOSE_FLAG = $(if $(VERBOSE),-v)

# lint
lint-deps:
	which golint || go get golang.org/x/lint/golint
lint: lint-deps
	go vet ./...
	rm -f .golint.txt
	golint ./... | grep -v 'internal.*exported\|translatedassert.*\(package comment\|exported\|underscores\)' | tee .golint.txt
	test ! -s .golint.txt

# test
test-deps:
	which dep || curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
test: test-deps
	go test $(VERBOSE_FLAG) $(TESTPKG)
test-integration: test
	cd _integrationtest && ./test.sh

# build gopwt binary
gopwt:
	go build

# run example
example: gopwt
	go test ./_example
# generate
gen-readme: gopwt
	_misc/gen-readme.bash
gen-op:
	_misc/gen-op.bash
gen: gen-op gen-readme
