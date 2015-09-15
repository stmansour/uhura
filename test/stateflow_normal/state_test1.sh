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
TOOLS_DIR="/usr/local/accord/testtools"
ENV_DESCR="env.json"
SLAVELOG="state_test1_script.log"
VERBOSE=0
UHOST=localhost
UPORT=8080

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
	bash ${TOOLS_DIR}/uhura_shutdown.sh -p {$UPORT} >>${SLAVELOG} 2>&1
	# Give the server a second to shutdown
	sleep 1
}

usage() {
	echo "Usage: $0 options..." 
	echo "optons:"
	echo "    -q           quiet mode"
	echo "    -d uhuraDir  directory where uhura executable lives"
	echo "    -t toolsDir  directory where test tools reside"
	echo "    -e envDescr  environment descriptor"
	echo "    -v           enables verbose mode"
    echo "   [-p port]     default is 8080"
    echo "   [-h host]     default is localhost"
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
	echo -n "bash ${TOOLS_DIR}/clientsim.sh -h ${UHOST} -p ${UPORT} -n MainTestInstance -u $1 -s $2"  >>${SLAVELOG} 2>&1

	bash ${TOOLS_DIR}/clientsim.sh -h ${UHOST} -p ${UPORT} -n MainTestInstance -u $1 -s $2 >>${SLAVELOG} 2>&1
	echo >>${SLAVELOG} 2>&1
}

#---------------------------------------------------------------------
#  optspec begins with ':', option letters follow, if the
#  option takes a param then it is followed by ':'
#---------------------------------------------------------------------
while getopts ":vd:t:e:p:h:" o; do
    case "${o}" in
        v)
            VERBOSE=1
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
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ ${VERBOSE} -gt 0 ]; then
    echo "Current working directory = $(pwd)"
	UVERBOSE="-D"
	echo "VERBOSE = ${VERBOSE}"
	echo "UHURA_DIR = ${UHURA_DIR}"
	echo "TOOLS_DIR = ${TOOLS_DIR}"
	echo "ENV_DESCR = ${ENV_DESCR}"
fi

#---------------------------------------------------------------------
#  Stop early if uhura is already running on this box...
#---------------------------------------------------------------------
# COUNT=$(ps -ef | grep uhura | grep -v grep | wc -l)
# if [ ${COUNT} -gt 0 ]; then
# 	echo "*** ERROR: There is another instance of uhura already running..."
# 	ps -ef | grep uhura | grep -v grep 
# 	echo "***        Stop this instance and try again."
# 	exit 1
# fi

#---------------------------------------------------------------------
#  Find accord bin...
#---------------------------------------------------------------------
if [ -d /usr/local/accord/bin ]; then
	ACCORDBIN=/usr/local/accord/bin
elif [ -d /c/Accord/bin ]; then
	ACCORDBIN=/c/Accord/bin
else
	echo "*** ERROR: Required directory /usr/local/accord/bin or /c/Accord/bin does not exist."
	echo "           Please repair installation and try again."
	exit 2
fi
if [ ${VERBOSE} -gt 0 ]; then
	echo "ACCORDBIN = ${ACCORDBIN}"
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
${UHURA_DIR}/uhura -p ${UPORT} -d ${UVERBOSE} -e ${ENV_DESCR} >uhura.out 2>&1 &

#---------------------------------------------------------------------
# Give the server a second startup
#---------------------------------------------------------------------
sleep 1

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

shutdown
if [ ! -e uhura.log ]; then
	echo "*** ERROR:  could not find uhura.log."
	exit 1
fi

mv uhura.log state_test1_master.log

#---------------------------------------------------------------------
#  Files produced:
#     * state_test1_master.log  - uhura log file from this test run
#     * state_test1_slave.log   - output from the client simulator (currently not used)

#  Other file definitions:
#     * state_test1_master.gold - uhura log file from this test case where the information
#                                in the log file is known to be correct.

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
#           |     year     /   month   /   day  | |   hr    :    min   :    sec  |   everything else
perl -pe 's/(20[1-4][0-9]\/[0-1][0-9]\/[0-3][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] )(.*)/$2/' state_test1_master.gold   \
| perl -pe 's/Tstamp: [A-Z][a-z]{2} [A-Z][a-z]{2} [ 0-1][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] [A-Z]{3} 20[1-4][0-9]/Tstamp: TIMESTAMP/' > x
#                     |     dow    |   month      |   dom    |    hour  :    min   :   sec    |   TZ   |  year     |

perl -pe 's/(20[1-4][0-9]\/[0-1][0-9]\/[0-3][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] )(.*)/$2/' state_test1_master.log   \
| perl -pe 's/Tstamp: [A-Z][a-z]{2} [A-Z][a-z]{2} [ 0-1][0-9] [0-2][0-9]:[0-5][0-9]:[0-5][0-9] [A-Z]{3} 20[1-4][0-9]/Tstamp: TIMESTAMP/' > y

#---------------------------------------------------------------------
#  Now deal with any port differences
#---------------------------------------------------------------------
perl -pe 's/master mode on port [0-9]+/Current working directory = SOMEdirectory/' x > x1; mv x1 x
perl -pe 's/master mode on port [0-9]+/Current working directory = SOMEdirectory/' y > y1; mv y1 y

#---------------------------------------------------------------------
#  Now deal with any working directory differences
#---------------------------------------------------------------------
perl -pe 's/^Current working directory = [\/a-zA-Z0-9]+/master mode on port SOMEPORT/' x > x1; mv x1 x
perl -pe 's/^Current working directory = [\/a-zA-Z0-9]+/master mode on port SOMEPORT/' y > y1; mv y1 y

#---------------------------------------------------------------------
#  Now see how they compare...
#---------------------------------------------------------------------
DIFFS=$(diff x y | wc -l)

if [ ${VERBOSE} -gt 0 ]; then
	echo "Functional differences between reference log and this test log: ${DIFFS}"
fi

if [ ${DIFFS} -eq 0 ]; then
	if [ ${VERBOSE} -gt 0 ]; then
		echo "PASSED"
	fi
	exit 0
else
	if [ ${VERBOSE} -gt 0 ]; then
		echo "FAILED:  differences are as follows:"
		diff x y
	fi
	exit 1
fi
