#!/bin/sh

chown -R iris:iris ./settings

exec su iris -c "./eps $*"
