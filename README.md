# Promtail Operator for Kubernetes

## Getting started

There are two ways of running the opeator: 1) run locally outside the cluster (for quick dev/test) or 2) run as a deployment inside the cluster. For either approach, the first step is to install the promtail CRD

```
make install
```

1. Run locally outside the cluster

Deploy the required ClusterRole, serviceAccount:

```
make deploy
```

Start the controller:

```
make run
```

Create an instance

```
kubectl apply -f ./config/samples/logging_v1_promtail.yaml 
```

2. Run as a deployment inside the cluster

```
make deploy
```
