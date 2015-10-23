all: clean mysqltest test
	echo "*** COMPLETE in sys4 ***"

clean:
	rm -rf *.log ./tmp qmstr* descr* EnvShutdownStatus.json mysqltest sys4 *.txt

mysqltest:
	go build
	mv sys4 mysqltest; ./mysqltest -u ec2-user

test: clean
	@echo "*** SYSTEM TEST 4 - verify mysql database provisioning  ***"
	./systest.sh
	@echo "*** SYSTEM TEST 4 COMPLETE in test/sys4 ***"

package: mysqltest
	rm -rf ./tmp
	mkdir -p ./tmp/mysqltest
	cp mysqltest ./tmp/mysqltest/
	cp activate.* ./tmp/mysqltest/
	cd ./tmp;tar cvf mysqltest.tar mysqltest;gzip mysqltest.tar

publish: package
	cd ./tmp;deployfile.sh mysqltest.tar.gz jenkins-snapshot/mysqltest/latest

