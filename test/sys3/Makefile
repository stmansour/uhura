all: clean test
	echo "*** COMPLETE in sys3 ***"

clean:
	rm -f *.log qmstr* descr* EnvShutdownStatus.json

test:   clean
	@echo "*** SYSTEM TEST 3 - full life cycle, 3 instances, 6 apps, dynamic substitution  ***"
	./systest.sh
	@echo "*** SYSTEM TEST 3 COMPLETE in test/sys3 ***"
