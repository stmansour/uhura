all: clean test
	echo "*** COMPLETE in sys0 ***"

clean:
	rm -f *.log qmstr* descr* EnvShutdownStatus.json

test:	clean
	@echo "*** SYSTEM TEST 1 - full life cycle, 3 Tgo instances ***"
	./systest01.sh
	@echo "*** SYSTEM TEST 1 COMPLETE in test/sys1 ***"
