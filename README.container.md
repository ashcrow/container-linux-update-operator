# OStree Container

## Components

- ``Dockerfile``: Container creation recipe
- ``ostree.sh``: Wrapper for ostree which uses ``chroot`` to enter ``/host`` and execute proper commands. This command also mounts the containers version of ``/lib64`` and ``/usr/bin/ostree`` over the ``/host``'s versions.


## Expected Mounts

- ``/host``: This mount is the hosts root file system
- ``/var/run/dbus/``: Allows access to the hosts dbus
- ``/run/dbus/``: Allows access to the hosts system bus


## Examples
```
# Pull an update
docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest pull fedora-atomic:fedora/27/x86_64/atomic-host
# Deploy the update
docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest admin deploy fedora-atomic:fedora/27/x86_64/atomic-host
# Look at the change
docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest admin status
# Reboot the node
docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest reboot
```
