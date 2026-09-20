#!/usr/bin/env bash

systemctl daemon-reload

ECHOLEN=$(echo -e|awk '{print length($0)}')
if [ "${ECHOLEN}" = '0' ]
then
    ECHO='echo -e'
else
    ECHO='echo'
fi;
    $ECHO "\033[36m******************************\033[0m"
    $ECHO "\033[36m*         Completed!         *\033[0m"
    $ECHO "\033[36m******************************\033[0m"
