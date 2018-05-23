all: clean test
	@echo "*** TEST COMPLETE in test/master_statechg_test ***"

clean:
	rm -f qm* *.log *.out [u-z] descrinst.json EnvShutdownStatus.json
	@echo "*** CLEAN COMPLETE in test/master_statechg_test ***"

test:
	./functest.sh -p 8100 -n
	@echo "*** TESTS PASSED in test/master_statechg_test ***"
