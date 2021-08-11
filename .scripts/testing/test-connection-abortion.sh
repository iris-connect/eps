#!/bin/bash
while true
do
    EPS_SETTINGS=settings/dev/roles/private-proxy-eps-1 eps --level debug server run &
    RUNNING_PID=$!
    sleep 0.1
    kill ${RUNNING_PID}
done
