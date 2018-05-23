TOP=..
THISDIR=test
TESTREPORT="testreport.txt"
DIRS=stateflow_normal publicdns
SYSDIRS=sys0 sys1 sys2 sys3 sys4

all: clean test
	@echo "*** COMPLETED in ${THISDIR} **"

clean:
	for dir in $(DIRS); do make -C $$dir clean;done
	@echo "*** CLEAN COMPLETED in ${THISDIR} ***"

test:
	# shutdown any running uhura on our port (8100)
	if [ `ps -ef | grep uhura | grep -v grep | grep "p 8100" | wc -l` -gt 0 ]; then RET=`/usr/local/accord/testtools/uhura_shutdown.sh -p 8100`;echo "uhura normal shutdown"; fi
	for dir in $(DIRS); do make -C $$dir test;done
	@echo "*** TEST COMPLETED in ${THISDIR} ***"

systest:
	for dir in $(SYSDIRS); do make -C $$dir test;done
	@echo "*** SYSTEM TESTS COMPLETED ***"

install:
	@echo "*** INSTALL COMPLETED in ${THISDIR} ***"
