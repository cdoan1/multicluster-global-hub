# Multicluster Global Hub Overview
[![Build](https://img.shields.io/badge/build-Prow-informational)](https://prow.ci.openshift.org/?repo=stolostron%2F${multicluster-global-hub})
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=open-cluster-management_hub-of-hubs&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=open-cluster-management_hub-of-hubs)
[![Go Reference](https://pkg.go.dev/badge/github.com/stolostron/multicluster-global-hub.svg)](https://pkg.go.dev/github.com/stolostron/multicluster-global-hub)
[![License](https://img.shields.io/github/license/stolostron/multicluster-global-hub)](/LICENSE)

This document attempts to explain how the different components in multicluster global hub come together to deliver multicluster management at very high scale. The multicluster-global-hub operator is the root operator which pulls in all things needed.

## Conceptual Diagram
 
![ArchitectureDiagram](doc/architecture/multicluster-global-hub-arch.png)

### Multicluster Global Hub Operator
Operator is for multicluster global hub. It is used to deploy all required components for multicluster management. The components include multicluster-global-hub-manager in the global hub cluster and multicluster-global-hub-agent in the regional hub clusters.

The Operator also leverages the manifestwork to deploy the Advanced Cluster Management for Kubernetes in the managed cluster. So the managed cluster is switched to a standard ACM Hub cluster (regional hub cluster).

### Multicluster Global Hub Manager
The manager is used to persist the data into the postgreSQL. The data is from Kafka transport. The manager is also used to post the data to Kafka transport so that it can be synced to the regional hub clusters.

### Multicluster Global Hub Agent
The agent is running in the regional hub clusters. It is responsible to sync-up the data between the global cluster hub and the regional hub clusters. For instance, sync-up the managed clusters' info from the regional hub clusters to the global hub cluster and sync-up the policy or application from the global hub cluster to the regional hub clusters.

## Quick Start Guide

### Prerequisites

1. Connect to a Kubernetes cluster with `kubectl`
2. ACM or OCM is installed on the Kubernetes cluster
3. PostgreSQL is installed and a database is created for the multicluster global hub. A secret `storage-secret` contains the credential is created in `open-cluster-management` namespace. The credential format like `postgres://<user>:<password>@<host>:<port>/<database>`:

```bash
kubectl create secret generic storage-secret -n "open-cluster-management" \
    --from-literal=database_uri=<postgresql-uri> 
```
> You can run this sample script `./operator/config/samples/postgres/deploy_postgres.sh` to install postgres in `hoh-postgres` namespace and create the secret `storage-secret` in namespace `open-cluster-management` automatically. 

4. Kafka is installed and two topics `spec` and `status` are created, also a secret with name `transport-secret` that contains the kafka access information should be created in `open-cluster-management` namespace:

```bash
kubectl create secret generic transport-secret -n "open-cluster-management" \
    --from-literal=bootstrap_server=<kafka-bootstrap-server-address> \
    --from-literal=CA=<CA-for-kafka-server>
```
> As above, You can run this sample script `./operator/config/samples/kafka/deploy_kafka.sh` to install kafka in kafka namespace and create the secret `transport-secret` in namespace `open-cluster-management` automatically. 

### Run the operator in the cluster

1. Check out the multicluster-global-hub repository
```bash
git clone git@github.com:stolostron/multicluster-global-hub.git
cd multicluster-global-hub/operator
```

2. Build and push your image to the location specified by `IMG`:

```bash
make docker-build docker-push IMG=<some-registry>/multicluster-global-hub-operator:<tag>
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```bash
make deploy IMG=<some-registry>/multicluster-global-hub-operator:<tag>
```

4. Install instance of custom resource:

```bash
kubectl apply -k config/samples/
```

### Uninstall the operator

1. Delete the multicluster-global-hub-operator CR:

```bash
kubectl delete mgh --all
```

2. Delete the multicluster-global-hub-operator:

```bash
make undeploy
```

3. To delete the multicluster global hub CRD from the cluster:

```bash
make uninstall
```

# Contributing

Go to the [Contributing guide](CONTRIBUTING.md) to learn how to get involved.
