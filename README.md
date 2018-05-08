# CoreOS v2 Update Operator (rpm-ostree version)

CoreOS Update Operator is a version of
the [Container Linux Update Operator](https://github.com/coreos/container-linux-update-operator) adapted
for [rpm-ostree](https://github.com/projectatomic/rpm-ostree).

It synchronizes host updates, integrating with Kubernetes, helping
to ensure seamless upgrades of the base operating system in a cluster.

There are two parts, an `update-agent` which runs as a DaemonSet on each node,
and an `update-operator` runs as a Deployment, watching changes to node
annotations and reboots the nodes as needed. It coordinates the reboots of
multiple nodes in the cluster, ensuring that not too many are rebooting at once.

Currently, `update-operator` only reboots one node at a time.

This operator fulfills the same purpose
as [locksmith](https://github.com/coreos/locksmith). And for people familiar
with `yum` based systems, one could think of `rpm-ostree` as on the same level
as `yum`, the `update-agent` as similar to `yum-cron` . The operator equivalent
might be implemented via configuration management or scripting.

The advantage of this operator over both is its integration with Kubernetes.

## More detailed design

[Original proposal](https://docs.google.com/document/d/1DHiB2UDBYRU6QSa2e9mCNla1qBivZDqYjBVn_DvzDWc/edit#)

The CoreOS v2 `update-agent` more directly tells rpm-ostree to upgrade, rather than
relying on a timer in `rpm-ostree` itself.  Then, if an update is available `rpm-ostree`
will expose a "cached update" property via DBus.  The agent will then
indicate via [node annotations](./pkg/constants/constants.go) that it needs a reboot.

The operator watches all node annotations, and takes care of choosing which
node to drain and reboot.

## Requirements

- A Kubernetes cluster (>= 1.6) running on a system using `rpm-ostree` such as today's Fedora Atomic

## Usage

Create the `update-operator` deployment and `update-agent` daemonset.

```
kubectl apply -f examples/deploy -R
```

## Test

To test that it is working, you could use `rpm-ostree deploy` to reset hosts
to previous versions, then enable the operator.
