#!/bin/sh

if [[ -z "${MOUNT_POINTS}" ]]
then
    echo "The environment variable MOUNT_POINTS can be set to a comma separated list of the mount points in the container. Use default now."
    MOUNT_POINTS="/config,/tls,/app/settings"
fi

echo "MOUNT_POINTS is set to \"${MOUNT_POINTS}\""

for point in ${MOUNT_POINTS//,/ }
do
    if [ ! -d ${point} ]
    then
        echo "Skip non-existent directory: \"${point}\""
        continue
    fi

    if ! su iris -c "find ${point} -type f -exec cat {} > /dev/null +"
    then
        echo "chown is made for \"${point}\""
        chown -R iris:iris ${point}
    else
        echo "Skip readable directory: \"${point}\""
    fi
done

echo "Execute eps with user iris"

exec su iris -c "./eps $*"
