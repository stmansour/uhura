#!/bin/bash
#  Uhura startup script - for use only when driving events from 
#  another source or from state_test1.sh -N
#  Uhura runs logging to the screen so we can watch what happens
#
#  Author:  sman
#  Version: 0.1  Tue Sep 18 18:31:33 PDT 2015

UHURA_DIR="../.."
ENV_DESCR="env.json"
SCRIPTLOG="solo_uhura.log"
VERBOSE=0
UHOST=localhost
UPORT=8100
DRYRUN=1
SKIP_UHURA=0

if [ ! -e "${UHURA_DIR}/uhura" ]; then
	if [ -e "/usr/local/accord/bin/uhura" ]; then
		UHURA_DIR="/usr/local/accord/bin"
		echo "**** NOTICE **** uhura was not found in relative location:  ../../uhura"
		echo "**** NOTICE **** using /usr/local/accord/bin/uhura instead"
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
    -e envDescr  environment descriptor
    -n           dry-run mode.  Do not create new cloud instances. this defaults to on
    [-p port]    default is 8100
    [-h host]    default is localhost
ZZEOF
	
	exit 1
}


#---------------------------------------------------------------------
#  optspec begins with ':', option letters follow, if the
#  option takes a param then it is followed by ':'
#---------------------------------------------------------------------
while getopts ":ne:p:h:" o; do
    case "${o}" in
        n)
            DRYRUN=1
            ;;
        e)
            ENV_DESCR=${OPTARG}
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

if [ ${DRYRUN} -gt 0 ]; then
	UDRYRUN="-n"
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
COUNT=$(ps -ef | grep uhura | grep -v grep | grep ${UPORT} | wc -l)
if [ ${COUNT} -gt 0 ]; then
	echo "*** NOTICE: attempting to stop uhura already running on port ${UPORT}..."
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
echo "../../uhura -p ${UPORT} -d ${UVERBOSE} ${UDRYRUN} -e ${ENV_DESCR}"
../../uhura -p ${UPORT} -d -D ${UVERBOSE} ${UDRYRUN} -e ${ENV_DESCR}
