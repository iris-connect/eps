#!/bin/sh

if ! su iris -c "find ./settings -type f -exec cat {} > /dev/null +"; then
    chown -R iris:iris ./settings
fi

exec su iris -c "./eps $*"
