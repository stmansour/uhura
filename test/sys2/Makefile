all: clean test
	echo "*** COMPLETE in sys2 ***"

clean:
	rm -f *.log qmstr* descr* EnvShutdownStatus.json

test:   clean
	@echo "*** SYSTEM TEST 2 - full life cycle, 1 TGO instance, 3 applications  ***"
	./systest.sh
	@echo "*** SYSTEM TEST 2 COMPLETE in test/sys2 ***"
