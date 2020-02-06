#!/bin/sh

systemctl daemon-reload
systemctl enable v2raya
systemctl start v2raya

ICON_SOURCE=gui/public/img/icons
ICON_TARGET=/usr/share/icons
mkdir -p $ICON_TARGET || true
cp $ICON_SOURCE/android-chrome-512x512.png $ICON_TARGET/v2raya.png || true

cp install/v2raya.desktop /usr/share/applications/v2raya.desktop || true

ECHOLEN=$(echo -e|awk '{print length($0)}')
if [ ${ECHOLEN} = '0' ]
then
    ECHO='echo -e'
else
    ECHO='echo'
fi;
$ECHO "\033[36m**************************************\033[0m"
$ECHO "\033[36m*         Congratulations!           *\033[0m"
$ECHO "\033[36m* HTTPS demo: https://v2raya.mzz.pub *\033[0m"
$ECHO "\033[36m* HTTP  demo: http://v.mzz.pub       *\033[0m"
$ECHO "\033[36m**************************************\033[0m"
