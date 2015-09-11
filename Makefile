all: clean uhura test
	@echo "*** COMPLETED ***"

.PHONY:  test

install: uhura
	cp uhura /usr/local/accord/bin
	@echo "*** INSTALL COMPLETED ***"

uhura: *.go
	go fmt
	go vet
	go build
	@echo "*** BUILD COMPLETED ***"

clean:
	go clean
	cd ./test;make clean
	@echo "*** CLEAN COMPLETE ***"

test: coverage
	cd ./test;make test
	@echo "*** TEST COMPLETE - ALL TESTS PASSED ***"

coverage:
	go test -coverprofile=c.out
	go tool cover -func=c.out
	go tool cover -html=c.out
