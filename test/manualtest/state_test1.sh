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
VERBOSE=1
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
# This simulates the 2 clients contacting the server and walking
# through their states.  This is just a straight functional test.
# There are no random pauses or timing tricks.
#---------------------------------------------------------------------
# sendStatus "prog1" "INIT"
# sendStatus "prog2" "INIT"
# sendStatus "prog1" "READY"
# sendStatus "prog2" "READY"
# sendStatus "prog1" "TEST"
# sendStatus "prog2" "TEST"
# sendStatus "prog1" "DONE"
# sendStatus "prog2" "DONE"

echo "we're using env.json"
default="prog1-INIT"
while [ 1 ]; do
	echo
	read -p "Enter uid-status [${default}] " ps
	if [ "x" == "x${ps}" ]; then
		ps=${default}
	fi
	P="${ps#*-}"
	U="${ps%-*}"
	${TOOLS_DIR}/clientsim.sh -h ${UHOST} -p ${UPORT} -n MainTestInstance -u $U -s $P
	default=${ps}
done