#!/bin/bash
#  Uhura Test Script
#  Validate the state changes in the server in the "happy path"
#  or 'no error' flow.  Do this as follows:

#  1. Starting up uhura in master mode, run it as a background task. Enable debug logging
#     so that state changes and other details are visible in the log
#  2. Use env.json as the environment descriptor (1 app, 1 test running on the same instance)
#  3. Use the script 'clientsim.sh' to send status messages to the uhura master.  Each status
#     message simulates the client moving through the following states:  INIT, READY, TEST, DONE
#  4. The local uhura slave also requests the map file
#  5. After all requests are finished, send a normal shutdown to the server
#  6. Compare the logfile from the uhura master to the known-good logfile and fail if there are
#     any functional differences
#
#  Author:  sman
#  Version: 0.1  Tue Sep  8 15:39:33 PDT 2015

UHURA_DIR="../.."
ENV_DESCR="env.json"
SCRIPTLOG="state_test1_script.log"
VERBOSE=0
UHOST=localhost
UPORT=8100
DRYRUN=0
SKIP_UHURA=0

if [ ! -e "${UHURA_DIR}/uhura" ]; then
	if [ -e "${GOPATH}/bin/uhura" ]; then
		UHURA_DIR="${GOPATH}/bin"
		echo "**** NOTICE **** uhura was not found in relative location:  ../../uhura"
		echo "**** NOTICE **** using ${GOPATH}/bin/uhura instead"
	else
		echo "**** ERROR **** uhura was not found in ../.. nor in ${GOPATH}/bin"
		exit 5
	fi
fi


shutdown() {
	bash ${TOOLS_DIR}/uhura_shutdown.sh -p {$UPORT} >>${SCRIPTLOG} 2>&1
	# Give the server a second to shutdown
	sleep 1
}

usage() {
	cat << ZZEOF
Usage: $0 options...
optons:
    -q           quiet mode
    -d uhuraDir  directory where uhura executable lives
    -t toolsDir  directory where test tools reside
    -e envDescr  environment descriptor
    -n           dry-run mode.  Do not create new cloud instances.
    -v           enables verbose mode
    -N 			 do not run uhura, assume it is already running
   [-p port]     default is 8080
   [-h host]     default is localhost
ZZEOF
	
	exit 1
}

#---------------------------------------------------------------------
# Function to send status to Uhura master
# $1 = UID
# $2 = status mode: {INIT|READY|TEST|DONE}
#---------------------------------------------------------------------
sendStatus() {
	if [ ${VERBOSE} -gt 0 ]; then
		echo "bash ${TOOLS_DIR}/clientsim.sh -h ${UHOST} -p ${UPORT} -n MainTestInstance -u $1 -s $2"
	fi
	echo -n "bash ${TOOLS_DIR}/clientsim.sh -h ${UHOST} -p ${UPORT} -n MainTestInstance -u $1 -s $2"  >>${SCRIPTLOG} 2>&1

	bash ${TOOLS_DIR}/clientsim.sh -h ${UHOST} -p ${UPORT} -n MainTestInstance -u $1 -s $2 >>${SCRIPTLOG} 2>&1
	echo >>${SCRIPTLOG} 2>&1
}

#---------------------------------------------------------------------
#  optspec begins with ':', option letters follow, if the
#  option takes a param then it is followed by ':'
#---------------------------------------------------------------------
while getopts ":vnNd:t:e:p:h:" o; do
    case "${o}" in
        v)
            VERBOSE=1
            ;;
        n)
            DRYRUN=1
            ;;
        d)
            UHURA_DIR=${OPTARG}
            ;;
        e)
            ENV_DESCR=${OPTARG}
            ;;
        t)
            TOOLS_DIR=${OPTARG}
            ;;
        p)
            UPORT=${OPTARG}
            ;;
        h)
			UHOST=${OPTARG}
			;;
        N)
			SKIP_UHURA=1
			;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ ${DRYRUN} -gt 0 ]; then
	UDRYRUN="-n"
fi

if [ ${VERBOSE} -gt 0 ]; then
	UVERBOSE="-D"
    echo "Current working directory = $(pwd)" >>${SCRIPTLOG} 2>&1
	echo "DRYRUN = ${DRYRUN}" >>${SCRIPTLOG} 2>&1
	echo "UDRYRUN = ${UDRYRUN}" >>${SCRIPTLOG} 2>&1
	echo "VERBOSE = ${VERBOSE}" >>${SCRIPTLOG} 2>&1
	echo "SKIP_UHURA = ${SKIP_UHURA}" >>${SCRIPTLOG} 2>&1
	echo "UHURA_DIR = ${UHURA_DIR}" >>${SCRIPTLOG} 2>&1
	echo "TOOLS_DIR = ${TOOLS_DIR}" >>${SCRIPTLOG} 2>&1
	echo "ENV_DESCR = ${ENV_DESCR}" >>${SCRIPTLOG} 2>&1
fi


#---------------------------------------------------------------------
#  Find accord bin...
#---------------------------------------------------------------------
if [ -d /usr/local/accord/bin ]; then
	ACCORDBIN=/usr/local/accord/bin
	TOOLS_DIR="/usr/local/accord/testtools"
elif [ -d /c/Accord/bin ]; then
	ACCORDBIN=/c/Accord/bin
	TOOLS_DIR="/c/Accord/testtools"
else
	echo "*** ERROR: Required directory /usr/local/accord/bin or /c/Accord/bin does not exist."
	echo "           Please repair installation and try again."
	exit 2
fi
if [ ${VERBOSE} -gt 0 ]; then
	echo "ACCORDBIN = ${ACCORDBIN}"
fi

#---------------------------------------------------------------------
#  hard stance now... if uhura is running on our port, stop it first
#---------------------------------------------------------------------
if [ ${SKIP_UHURA} -eq 0 ]; then
	COUNT=$(ps -ef | grep uhura | grep -v grep | grep ${UPORT} | wc -l)
	if [ ${COUNT} -gt 0 ]; then
		echo "*** NOTICE: attempting to stop uhura already running on port ${UPORT}..."
		echo "SKIP_UHURA = ${SKIP_UHURA}"
		${TOOLS_DIR}/uhura_shutdown.sh -p ${UPORT}
		COUNT=$(ps -ef | grep uhura | grep -v grep | grep ${UPORT} | wc -l)
		if [ ${COUNT} -gt 0 ]; then
			echo "*** cannot stop it.  exiting..."
			exit 6
		fi
	fi

	#---------------------------------------------------------------------
	#  Validate that all the files that Uhura depends on are in place...
	#---------------------------------------------------------------------
	declare -a dependencies=('cr_linux_testenv.sh' 'cr_win_testenv.sh' 
				             'qmaster.scr1' 'qmaster.scr2' 'qmaster.sh')

	missing=0
	for dep in ${dependencies[@]}; do
		if [ ! -e ${ACCORDBIN}/${dep} ]; then
			((++missing))
			echo "*** ERROR: Required file ${ACCORDBIN}/${dep} was not found"
		fi
	done
	if [ $missing -gt 0 ]; then
		echo "           Please install the missing files and try again."
		exit 3
	fi

	rm -f qm* *.log *.out
	echo "${UHURA_DIR}/uhura -p ${UPORT} -d ${UVERBOSE} ${UDRYRUN} -e ${ENV_DESCR} >uhura.out 2>&1 &" >>${SCRIPTLOG} 2>&1
	${UHURA_DIR}/uhura -p ${UPORT} -d ${UVERBOSE} ${UDRYRUN} -e ${ENV_DESCR} >uhura.out 2>&1 &

	#---------------------------------------------------------------------
	# Give the server a second startup
	#---------------------------------------------------------------------
	sleep 1
fi

#---------------------------------------------------------------------
# This simulates the 2 clients contacting the server and walking
# through their states.  This is just a straight functional test.
# There are no random pauses or timing tricks.
#---------------------------------------------------------------------
sendStatus "prog1" "INIT"
sendStatus "prog2" "INIT"
sendStatus "prog1" "READY"
sendStatus "prog2" "READY"
sendStatus "prog1" "TEST"
sendStatus "prog2" "TEST"
sendStatus "prog1" "DONE"
sendStatus "prog2" "DONE"

if [ ${SKIP_UHURA} -eq 0 ]; then

	sleep 1
	shutdown  ## should not need to shutdown, when all are in DONE state it will shutdown automatically
	if [ ! -e uhura.log ]; then
		echo "*** ERROR:  could not find uhura.log."
		exit 1
	fi

	mv uhura.log state_test1.log

	#---------------------------------------------------------------------
	#  Files produced:
	#     * state_test1.log  - uhura log file from this test run
	#     * state_test1_slave.log   - output from the client simulator (currently not used)

	#  Other file definitions:
	#     * state_test1.gold - uhura log file from this test case where the information
	#                          in the log file is known to be correct.

	#  Compare the "gold" output to log file output from this run
	#     *  ignore differences in timestamps
	#     *  ignore differences in the startup port. On different systems
	#        uhura may need to listen on different ports. This is not a
	#        a functional issue.
	#     *  ignore differences in the current workng directory. It will be
	#        different on different machines and different operating systems
	#     *  fail if there are any other differences
	#---------------------------------------------------------------------


	#---------------------------------------------------------------------
	#  Deal with the timestamps by essentially filtering out timestamps in 
	#  both gold log and this run's log:

	#  1. The log files contain the date and time at the beginning of each line.
	#     The first perl invocation removes it. 

	#  2. Part of the client status message to the uhura master includes a timestamp.  
	#     In the logfiles, these show up in this form: Tstamp: Tue Sep  8 14:17:39 PDT 2015
	#     We replace all timestamps in this form with "TIMESTAMP".

	#  Note A: if the timestamp becomes critical we'll have to adopt a different approach
	#  Note B: for the year the regexp constrains to the following range 2010 - 2049, and a few
	#          other constraints on days, hrs, mins, secs are in these regexps.  They're probably
	#          fine, but if we see any miscompares in timestamps, look closely at the regexps.
	#---------------------------------------------------------------------

	declare -a uhura_filters=(
		's/(20[1-4][0-9]\/[0-1][0-9]\/[0-3][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] )(.*)/$2/'	
		's/Tstamp:.*/Tstamp: TIMESTAMP/'
		's/master mode on port [0-9]+/Current working directory = SOMEDIRECTORY/'
		's/^Current working directory = [\/a-zA-Z0-9]+/master mode on port SOMEPORT/'
		's/^exec [\/_\.a-zA-Z0-9]+ [\/_\.\-a-zA-Z0-9]+ [\/\._a-zA-Z0-9]+.*/exec SOMEPATH/g'
		's/^Uhura starting on:.*/URL: somehost:someport/'
	)
	echo "Validating output..." >>${SCRIPTLOG} 2>&1
	cp state_test1.gold x
	cp state_test1.log y
	for f in "${uhura_filters[@]}"
	do
		perl -pe "$f" x > x1; mv x1 x
		perl -pe "$f" y > y1; mv y1 y
	done

	#---------------------------------------------------------------------
	#  Now see how they compare...
	#---------------------------------------------------------------------
	DIFFS=$(diff x y | wc -l)

	#---------------------------------------------------------------------
	# some randomness in where go routines get their timeslice...
	#---------------------------------------------------------------------
	if [ ${DIFFS} -gt 0 -a ${DIFFS} -lt 5 ]; then
		diff x y | grep "^[<>]" | perl -pe "s/^[<>]//" >z
		DIFFS=$(grep -v "Comms: sending TESTNOW" z | wc -l)
	fi

	if [ ${VERBOSE} -gt 0 ]; then
		echo "Functional differences between reference log and this test log: ${DIFFS}"
	fi

	if [ ${DIFFS} -eq 0 ]; then
		if [ ${VERBOSE} -gt 0 ]; then
			echo "VALIDATION 1 PASSED"
		fi
	else
		if [ ${VERBOSE} -gt 0 ]; then
			echo "VALIDATION 1 FAILED:  differences are as follows:"
			diff x y
		fi
		exit 1
	fi

	#---------------------------------------------------------------------
	#  Now check the script output...
	#---------------------------------------------------------------------
	declare -a scriptout_filters=(
		's/\"Timestamp\":.*/Timestamp: TIMESTAMP/'
	)
	cp state_test1_script.gold v
	cp state_test1_script.log w
	for f in "${scriptout_filters[@]}"
	do
		perl -pe "$f" v > v1; mv v1 v
		perl -pe "$f" w > w1; mv w1 w
	done

	DIFFS=$(diff v w | wc -l)

	if [ ${VERBOSE} -gt 0 ]; then
		echo "Functional differences between reference log and this test log: ${DIFFS}"
	fi

	if [ ${DIFFS} -eq 0 ]; then
		if [ ${VERBOSE} -gt 0 ]; then
			echo "VALIDATION 2 PASSED"
		fi
	else
		if [ ${VERBOSE} -gt 0 ]; then
			echo "VALIDATION 2 FAILED:  differences are as follows:"
			diff v w
		fi
		exit 1
	fi

	if [ ${VERBOSE} -gt 0 ]; then
		echo "ALL VALIDATIONS PASSED"
	fi
	exit 0

fi
