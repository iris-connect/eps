#!/bin/sh

if [[ -z "${MOUNT_POINTS}" ]]
then
    echo "The environment variable MOUNT_POINTS can be set to a comma separated list of the mount points in the container. Use default now."
    READ_MOUNT_POINTS="/config,/tls,/app/settings"
    WRITE_MOUNT_POINTS="/storage,/app/db,/tmp"
    MOUNT_POINTS=$READ_MOUNT_POINTS,$WRITE_MOUNT_POINTS
fi

echo "MOUNT_POINTS is set to \"${MOUNT_POINTS}\""
echo "READ_MOUNT_POINTS is set to \"${READ_MOUNT_POINTS}\""
echo "WRITE_MOUNT_POINTS is set to \"${WRITE_MOUNT_POINTS}\""

for point in ${MOUNT_POINTS//,/ }
do
    if [ ! -d ${point} ]
    then
        echo "Skip non-existent directory: \"${point}\""
        continue
    fi

    echo "chown is made for \"${point}\""
    chown -R iris:iris ${point}
    
    echo "chmod u+r is made for \"${point}\""
    chmod -R u+r ${point}
done

for point in ${WRITE_MOUNT_POINTS//,/ }
do
    if [ ! -d ${point} ]
    then
        continue
    fi
    
    echo "chmod u+w is made for \"${point}\""
    chmod -R u+w ${point}
done

echo "Execute sd with user iris"

exec su iris -c "./sd $*"
