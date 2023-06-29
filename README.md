# misim-k8s-adapter
Adapter for connecting MiSim to Kubernetes components.

## Usage

This adapter should be used with the [MiSim Orchestration Extension](https://github.com/DescartesResearch/misim-orchestration)
and handles connections to Kubernetes components (such as the kube-scheduler or cluster-autoscaler). Therefore, it 
implements selected party of the Kubernetes API. Consider the docs of the MiSim Orchestration Extension repository to
get examples.

## Common Pitfalls

Make sure to follow the correct order of starting the artifacts:
1. Start this adapter using the `run.sh` script
2. Start Kubernetes components
3. Start the MiSim Orchestration Extension

For Kubernetes components that use leader election mechanisms make sure to deactivate them at start by passing
`--leader-elect=false`, as, by now, we do not implement the Kubernetes Leases API.

## Cite us

The paper related to this repository is currently under review. We will add citation info as soon as available.

## Any questions?

For questions contact [Martin Straesser](https://se.informatik.uni-wuerzburg.de/software-engineering-group/staff/martin-straesser/).