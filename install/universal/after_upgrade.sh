#!/usr/bin/env bash

systemctl daemon-reload

if [ "$(systemctl is-active v2raya)" = "active" ]; then
    systemctl restart v2raya
fi
