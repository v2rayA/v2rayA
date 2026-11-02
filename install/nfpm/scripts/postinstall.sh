#!/bin/sh
# Only the systemd distributions have a manager to notify. Restarting a
# running v2raya on upgrade keeps the core in step with the new binary; a
# package that is merely installed has nothing running yet.
if command -v systemctl >/dev/null 2>&1; then
	systemctl daemon-reload || true
	systemctl try-restart v2raya.service >/dev/null 2>&1 || true
fi
