all: clean uhura test
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
	cd test;make clean
	@echo "*** CLEAN COMPLETE ***"

test:
	go test -run aws_test
	go test -run state_test
	go test -run http_test
	@./uhura -u -d -D;echo "Internal stress tests PASS"
	cd test;make test
	@echo "*** TEST COMPLETE - ALL TESTS PASSED ***"

systest:
	cd test;make systest
	@echo "*** SYSTEM TESTS COMPLETE - ALL TESTS PASSED ***"

coverage:
	go test -coverprofile=c.out
	go tool cover -func=c.out
	go tool cover -html=c.out
