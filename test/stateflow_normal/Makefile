all: clean test
	@echo "*** TEST COMPLETE in test/master_statechg_test ***"

clean:
	rm -f qm* *.log *.out [u-z] descrinst.json
	@echo "*** CLEAN COMPLETE in test/master_statechg_test ***"

test:
	./state_test1.sh -p 8100 -n
	@echo "*** TESTS PASSED in test/master_statechg_test ***"
