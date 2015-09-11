all: clean test
	@echo "*** TEST COMPLETE in test/master_statechg_test ***"

clean:
	rm -f qm* *.log *.out x y
	@echo "*** CLEAN COMPLETE in test/master_statechg_test ***"

test:
	./state_test1.sh -p8123
	@echo "*** TESTS PASSED in test/master_statechg_test ***"
