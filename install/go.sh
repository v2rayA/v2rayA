#!/bin/bash

echo -e "\e[37;44;4;1mThe old script is abandoned, we will redirect you to the new one in 5 seconds.\e[0m"
echo -e "\e[37;44;4;1mIf you want to download the new script and run it manually,
press Ctrl+C and visit https://v2raya.org/en/docs/prologue/installation/\e[0m"

sleep 5s
curl -Ls https://mirrors.v2raya.org/go.sh | sudo bash
