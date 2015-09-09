all: clean uhura test
	@echo "*** COMPLETED ***"

.PHONY:  test

install: uhura
	cp uhura /usr/local/accord/bin
	@echo "*** INSTALL COMPLETED ***"

uhura: *.go
	go build
	@echo "*** BUILD COMPLETED ***"

clean:
	rm -f uhura
	cd ./test;make clean
	@echo "*** CLEAN COMPLETE ***"

test:
	cd ./test;make test
	@echo "*** TEST COMPLETE - ALL TESTS PASSED ***"
