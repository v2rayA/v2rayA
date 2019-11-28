#!/bin/sh

systemctl daemon-reload
systemctl restart v2raya
echo -e "\033[36m**************************************\033[0m"
echo -e "\033[36m*         Congratulations!           *\033[0m"
echo -e "\033[36m* HTTP  demo: http://v.mzz.pub       *\033[0m"
echo -e "\033[36m* HTTPS demo: https://v2raya.mzz.pub *\033[0m"
echo -e "\033[36m**************************************\033[0m"
