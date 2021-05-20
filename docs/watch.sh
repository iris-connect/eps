#!/bin/bash

venv/bin/python3 -m http.server --directory build --bind 0.0.0.0 8112 &
SERVE_PID=$!
trap 'kill $SERVE_PID' EXIT
which inotifywaits
INOTIFY_AVAILABLE=$!
which fswatch
FSWATCH_AVAILABLE=$!
if which inotifywait; then
    while true ; do \
        inotifywait -r src -e create,delete,move,modify || break; \
        make site || break; \
    done
elif which fswatch; then
	while true ; do \
		fswatch -1 src || break; \
		$(MAKE) || break; \
	done
else
    echo "please install inotify-watch or fswatch"
    exit -1
fi
