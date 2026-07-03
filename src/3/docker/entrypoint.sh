#!/usr/bin/env bash

set -e

/setup.sh
nginx
/watch-nginx.sh &
exec /usr/sbin/sshd -D
