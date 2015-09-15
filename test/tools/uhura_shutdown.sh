#!/bin/bash
HOST=localhost
PORT=8080

usage() {
    echo "Usage:   uhura_shutdown [-h host] [-p port]"
    echo "         default host = localhost"
    echo "         default port = 8080"
    echo "Example: uhura_shutdown -p 8123 "
}
while getopts ":h:p:" o; do
    case "${o}" in
        h)
            HOST=${OPTARG}
            ;;
        p)
            PORT=${OPTARG}
            ;;
        *)
            usage
            ;;
    esac
done
shift $((OPTIND-1))


RET=$(curl -s http://${HOST}:${PORT}/shutdown/)
exit
