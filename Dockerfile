# AU image
FROM registry.fedoraproject.org/fedora:27

# Note, grub2-tools must have it's arch otherwise fedora won't install it (??)
RUN dnf update -y && \
    dnf install -y rpm-ostree grub2-tools-minimal grub2-tools.x86_64 grub2-tools-extra && \
    dnf clean all -y

# /:/host
VOLUME /host
# SOURCE:DEST
VOLUME /run/dbus
VOLUME /var/run/dbus

COPY ostree.sh /usr/bin/ostree.sh
# Examples:
# docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest pull fedora-atomic:fedora/27/x86_64/atomic-host
# docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest admin deploy fedora-atomic:fedora/27/x86_64/atomic-host
# docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest admin status
# docker run --privileged --rm -ti -v /:/host -v /var/run/dbus:/var/run/dbus -v /run/dbus:/run/dbus au:latest reboot
ENTRYPOINT ["/usr/bin/ostree.sh"]
