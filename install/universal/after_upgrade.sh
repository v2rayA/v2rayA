#!/bin/sh

systemctl daemon-reload

ECHOLEN=$(echo -e|awk '{print length($0)}')
if [ ${ECHOLEN} = '0' ]
then
    ECHO='echo -e'
else
    ECHO='echo'
fi;
    $ECHO "\033[36m******************************\033[0m"
    $ECHO "\033[36m*         Completed!         *\033[0m"
    $ECHO "\033[36m******************************\033[0m"
    $ECHO
    $ECHO "\033[36mWARN: v2raya@.service was deprecated; please use user service v2raya-lite.service instead.\033[0m"
    $ECHO "\033[36m      This does NOT impact v2raya.service users.\033[0m"
