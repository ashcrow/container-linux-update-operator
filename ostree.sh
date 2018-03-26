#!/usr/bin/bash
# If we are told to reboot, then execute reboot instead allowing
# options to pass through
if [ "$1" == "reboot" ]; then
	/usr/sbin/reboot "${@:2}"
	exit
fi
mount --bind /lib64 /host/lib64
mount --bind /usr/bin/ostree /host/usr/bin/ostree
chroot /host /usr/bin/ostree $@
umount /host/lib64
umount /host/usr/bin/ostree
