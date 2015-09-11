all: clean test
	@echo "*** COMPLETED in test **"

clean:
	cd ./stateflow_normal;make clean
	cd ./publicdns;make clean
	@echo "*** CLEAN COMPLETED in test ***"

test:
	cd ./stateflow_normal;make test
	cd ./publicdns;make test
	@echo "*** TEST COMPLETED in test ***"
