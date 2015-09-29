all: clean test
	echo "*** COMPLETE in sys0 ***"

clean:
	rm -f *.log qmstr* descr*

test:	clean
	@echo "*** SYSTEM TEST 0 - full life cycle, 1 Tgo instance ***"
	./systest00.sh
	@echo "*** SYSTEM TEST 0 COMPLETE in test/sys0 ***"
