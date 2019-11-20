#!/bin/sh

systemctl daemon-reload
systemctl enable v2raya
systemctl start v2raya
