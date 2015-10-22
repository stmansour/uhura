#!/bin/bash
# tester for echosrv

TESTRESULTS="testresults.txt"
TESTSTART="teststart.txt"
HOST="$(hostname)"
LOGFILE="test.log"
USER="ec2-user"

usage() {
    cat << ZZEOF

Usage:   activate.sh [OPTIONS] COMMAND

OPTIONS
-u user    Set the user of the database to user

COMMAND
One of: start | stop | ready | test | teststatus | testresults
cmd is case insensitive

Examples:
Command to perform the test:
	bash$  activate.sh TEST 

Command to determine test results:
	bash$  activate.sh TestResults

ZZEOF

	exit 0
}

PassFail() {
	C=$(tail -n 1 ${LOGFILE} | grep Passed | wc -l) 
	if [ ${C} -eq 1 ]; then
		echo "PASS" > ${TESTRESULTS}
		rm -f ${LOGFILE}
	else
		echo "FAIL" > ${TESTRESULTS}
	fi
}

init() {
	date > ${TESTSTART}
	sleep 1
	rm -f ${TESTRESULTS}
}

testit() {
	init
	echo "./mysqltest" >${LOGFILE}
	./mysqltest  >${LOGFILE}
	PassFail
}

while getopts ":hu:" o; do
    case "${o}" in
        h)
            usage
            ;;
        u)
			USER=${OPTARG}
			echo "USER set to ${USER}"
			;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

for arg do
	cmd=$(echo ${arg}|tr "[:upper:]" "[:lower:]")
    case "${cmd}" in
	"start")
		echo "OK" 
		;;
	"stop")
		echo "OK"
		;;
	"ready")
		echo "OK"
		;;
	"test")
		testit
		echo "OK"
		;;
	"teststatus")
		if [ -f ${TESTSTART} ]; then
			if [ -f ${TESTRESULTS} ]; then
				if [[ ${TESTRESULTS} -nt ${TESTSTART} ]]; then
					echo "DONE"
					exit 0
				fi
			fi
			echo "TESTING"
		else
			echo "ERROR not testing"
		fi
		;;
	"testresults")
		if [ -f "${TESTRESULTS}" ]; then
			cat ${TESTRESULTS}
		else
			echo "ERROR No test results"
		fi
		;;
	*)
		echo "Unrecognized command: ${cmd}"
		exit 1
		;;
    esac
done
exit 0
