#!/bin/bash

ACCORD_BIN=/usr/local/accord/bin
ACCORD_TOOLS=/usr/local/accord/testtools
UHURAPORT=8123
SYS0_TEST_DIR=$(pwd)
UHURAHOST=localhost

if [ $(uname) == "Linux" ]; then
	UHURAHOST=$(curl http://169.254.169.254/latest/meta-data/public-hostname)
fi

# Start up uhura with env descr: sys0.json
${ACCORD_BIN}/uhura -p ${UHURAPORT} -e ${SYS0_TEST_DIR}/sys0.json -t ${UHURAHOST} -d
