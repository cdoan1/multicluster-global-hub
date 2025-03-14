# multicluster-global-hub-operator

The operator of multicluster global hub (see: https://github.com/stolostron/multicluster-global-hub)

## Prerequisites

1. Connect to a Kubernetes cluster with `kubectl`
2. ACM or OCM is installed on the Kubernetes cluster
3. PostgreSQL is installed and a database is created for multicluster global hub, also a secret with name `storage-secret` that contains the credential should be created in `open-cluster-management` namespace. The credential format like `postgres://<user>:<password>@<host>:<port>/<database>`:

```bash
kubectl create secret generic storage-secret -n "open-cluster-management" \
    --from-literal=database_uri=<postgresql-uri> 
```
> You can run this sample script `config/samples/storage/deploy_postgres.sh` to install postgres in `hoh-postgres` namespace and create the secret `storage-secret` in namespace `open-cluster-management` automatically. 

4. Kafka is installed and two topics `spec` and `status` are created, also a secret with name `transport-secret` that contains the kafka access information should be created in `open-cluster-management` namespace:

```bash
kubectl create secret generic transport-secret -n "open-cluster-management" \
    --from-literal=bootstrap_server=<kafka-bootstrap-server-address> \
    --from-literal=CA=<CA-for-kafka-server>
```
> As above, You can run this sample script `config/samples/transport/deploy_kafka.sh` to install kafka in kafka namespace and create the secret `transport-secret` in namespace `open-cluster-management` automatically. 

## Getting started

### Running on the cluster

1. Build and push your image to the location specified by `IMG`:

```bash
make docker-build docker-push IMG=<some-registry>/multicluster-global-hub-operator:<tag>
```

2. Deploy the controller to the cluster with the image specified by `IMG`:

```bash
make deploy IMG=<some-registry>/multicluster-global-hub-operator:<tag>
```

3. Install Instances of Custom Resource:

```bash
kubectl apply -k config/samples/
```

### Uninstall CRD

To delete the CRD from the cluster:

```bash
make uninstall
```

### Undeploy controller

Undeploy the controller from the cluster:

```bash
make undeploy
```

## Contributing

### Test It Out Locally

1. Install CRD and run operator locally:

```bash
make install run
```

2. Install Instances of Custom Resource:

```bash
kubectl apply -k config/samples/
```

### Modifying the API definitions

If you are editing the API definitions, generate the generated code and manifests such as CRs, CRDs, CSV using:

```bash
make generate manifests bundle
```

NOTE: Run `make --help` for more information on all potential make targets
