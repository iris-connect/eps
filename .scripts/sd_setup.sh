#!/bin/bash

reset=--reset

for entry in `ls ${1}`; do
	if [ "${entry: -5}" == ".json" ]; then
		echo "Importing ${1}/${entry}..."
		EPS_SETTINGS=settings/dev/roles/hd-1 eps sd submit-records ${reset} settings/dev/directory/${entry}
		# we only call reset on the first file
		reset=''
	fi
done
