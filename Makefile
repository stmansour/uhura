all: clean uhura test install
	@echo "*** COMPLETED ***"

.PHONY:  test

install: uhura
	cp uhura /usr/local/accord/bin
	cd test;make install
	@echo "*** INSTALL COMPLETED ***"

uhura: *.go
	go fmt
	go vet
	go build
	@echo "*** BUILD COMPLETED ***"

clean:
	go clean
	rm -f *.log qmstr* *.out
	cd ./test;make clean
	@echo "*** CLEAN COMPLETE ***"

test:
	go test
	cp uhura /usr/local/accord/bin/
	cd ./test;make test
	@echo "*** TEST COMPLETE - ALL TESTS PASSED ***"

coverage:
	go test -coverprofile=c.out
	go tool cover -func=c.out
	go tool cover -html=c.out
	@echo "*** COVERAGE COMPLETE - ALL UNIT TESTS PASSED ***"
