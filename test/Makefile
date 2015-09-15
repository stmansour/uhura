all: clean test
	@echo "*** COMPLETED in test **"

clean:
	cd ./stateflow_normal;make clean
	cd ./publicdns;make clean
	@echo "*** CLEAN COMPLETED in test ***"

test:
	# shutdown any running uhura on our port (8123)
	if [ `ps -ef | grep uhura | grep -v grep | grep "p 8123" | wc -l` -gt 0 ]; then RET=`/usr/local/accord/testtools/uhura_shutdown.sh -p 8123`;echo "uhura normal shutdown"; fi
	cd ./stateflow_normal;make test
	cd ./publicdns;make test
	@echo "*** TEST COMPLETED in test ***"

install:
	@echo "*** INSTALL COMPLETED in test ***"
