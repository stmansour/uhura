all: clean test
	@echo "*** COMPLETED in test **"

clean:
	cd ./stateflow_normal;make clean
	cd ./publicdns;make clean
	@echo "*** CLEAN COMPLETED in test ***"

test:
	# shutdown any running uhura on our port (8100)
	if [ `ps -ef | grep uhura | grep -v grep | grep "p 8100" | wc -l` -gt 0 ]; then RET=`/usr/local/accord/testtools/uhura_shutdown.sh -p 8100`;echo "uhura normal shutdown"; fi
	cd ./stateflow_normal;make test
	cd ./publicdns;make test
	@echo "*** TEST COMPLETED in test ***"

systest:
	cd sys0;make test
	cd sys1;make test
	cd sys2;make test
	cd sys3;make test
	cd sys4;make test
	@echo "*** SYSTEM TESTS COMPLETED ***"

install:
	@echo "*** INSTALL COMPLETED in test ***"
