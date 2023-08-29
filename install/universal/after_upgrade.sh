#!/usr/bin/env bash

systemctl daemon-reload

if [ "$(systemctl is-active v2raya)" = "active" ]; then
    echo "Restarting v2rayA service..."
    systemctl restart v2raya
    echo "v2rayA service restarted"
fi

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
