Related github issue:

[Statefulset pods not rescheduled when node is powered off #74689](https://github.com/kubernetes/kubernetes/issues/74689)

[EKS Configurable or shorter delay for Node Auto Repair node termination](https://github.com/aws/containers-roadmap/issues/2587)

The controller force deletes statefulsets pods which stuck on terminating status when their nodes are unhealthy or not ready(Normally, due to node az network issues, and the statefulset pod can be stuck up to 50 mins), This ensures that the pods are rescheduled on healthy nodes, maintaining application availability.
