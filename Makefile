all: clean uhura test install
	@echo "*** COMPLETED ***"

.PHONY:  test

install: uhura
	cp uhura /usr/local/accord/bin
	cd test;make install
	@echo "*** INSTALL COMPLETED ***"

uhura: *.go
	go fmt
	gl=$(which golint);if [ "x${gl}" != "x" ]; then golint; fi
	go vet
	go build
	@echo "*** BUILD COMPLETED ***"

clean:
	go clean
	rm -f *.log qmstr* *.out EnvShutdownStatus.json
	cd test;make clean
	@echo "*** CLEAN COMPLETE ***"

test:
	go test
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
