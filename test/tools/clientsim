#!/bin/bash
#  optspec begins with ':', option letters follow, if the
#  option takes a param then it is followed by ':'
HOST=localhost
PORT=8080

usage() {
    echo "Usage:   clientsim -n instname -u instanceUid -s state [-h host] [-p port] [-i]"
    echo "                                state: INIT|READY|TEST|DONE"
    echo "Example: clientsim -p 8123 -n MainTestInstance -u prog1 -s READY"
}
while getopts ":n:u:s:p:ih:" o; do
    case "${o}" in
        n)
            INSTNAME=${OPTARG}
            ;;
        h)
            HOST=${OPTARG}
            ;;
        u)
            INSTUID=${OPTARG}
            ;;
        s)
            STATE=${OPTARG}
            ;;
        p)
            PORT=${OPTARG}
            ;;
        i)
            CURLOPTS="-i"
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))

if [ "z" == "z${INSTNAME}" ]; then
	echo "You must supply the instance name"
	usage
    exit 1
fi
if [ "z" == "z${INSTUID}" ]; then
	echo "You must supply the instance uid"
	usage
    exit 1
fi
if [ "z" == "z${STATE}" ]; then
	echo "You must supply the state"
	usage
    exit 1
fi

#  curl options:
#    -s  = don's show progress meter or error messages
#    -H  = headers
TSTAMP=$( date )
curl -s -H "content-Type: application/json" -X POST -d "{\"State\":\"${STATE}\", \"InstName\": \"${INSTNAME}\", \"UID\": \"${INSTUID}\", \"Tstamp\": \"$TSTAMP\"}" http://${HOST}:${PORT}/status/
