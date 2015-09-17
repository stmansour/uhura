#!/bin/bash

ACCORD_BIN=/usr/local/accord/bin
ACCORD_TOOLS=/usr/local/accord/testtools
UHURAPORT=8123
SYS0_TEST_DIR=$(pwd)
UHURAHOST=localhost

usage() {
	cat << ZZEOF
Usage: $0 options...
Optons:
    -n           dry-run mode.  Do not create new cloud instances.

Description:     Typically, this program is run with no options. If
                 you want to get at the uhura generated scripts without
                 having them launch aws instances, use the -n switch.

ZZEOF
	exit 1
}


while getopts ":n" o; do
    case "${o}" in
        n)
            DRYRUN=1
	    DOPTS="-n"
            ;;
	*)
	    usage
	    ;;
    esac
done

#
# Sometimes we test on a Mac, in which case "localhost" is OK. But if
# we're running on Linux, then it's probably on a build machine or in
# production. So, we should use the publicly visible dns name.
#
if [ $(uname) == "Linux" ]; then
	UHURAHOST=$(curl -s http://169.254.169.254/latest/meta-data/public-hostname)
fi

# Start up uhura with env descr: sys0.json
${ACCORD_BIN}/uhura ${DOPTS} -p ${UHURAPORT} -e ${SYS0_TEST_DIR}/sys0.json -t ${UHURAHOST} -d
