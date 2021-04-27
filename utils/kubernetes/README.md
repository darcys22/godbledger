# Deploying to Kubernetes

[Kubernetes](https://kubernetes.io/) (often abbreviated as `k8s`) is a common orchestration platform for automating the deployment, scaling, and management of containerized applications.

## Files

This directory contains example spec yaml which can get you up and running locally in minikube or the kubernetes instance offered by Docker Desktop.  They can also serve as examples to be modified for deployment to a kuberentes cluster hosted remotely.

| file                     | description                                                                                |
| ------------------------ | ------------------------------------------------------------------------------------------ |
| k8s.00.namespace.yml     | creates a `godbledger` namespace                                                           |
| k8s.01.mysql.secrets.yml | creates a secret resource with MYSQL credentials                                           |
| k8s.02.mysql.volume.yml  | creates a persistent storage volume for the `mysql` database                               |
| k8s.03.mysql.yml         | spins up `mysql` deployment and service                                                    |
| k8s.04.godbledger.yml    | spins up a `godbledger` deployment (running the `godbledger` app on port 8080) and service |
| k8s.config.toml          | a config file with connection details to `mysql` and `godbledger`                          |

Couple of notes on these files:

- they are all configured with a single [namespace](https://kubernetes.io/docs/reference/kubernetes-api/cluster-resources/namespace-v1/) called `godbledger`

- there is nothing here which requires these resources to be in a single namespace

- how you organize your kubernetes deployments is up to you, but notice that there is a `metadata.namespace` property in each of these resources which may need to updated if you make different choices.

## Apply

You can run these commands from any directory; simply adjust the pathed arguments accordingly.

1. Build the `godbledger:latest` container image:

    ```
    make build-docker
    ```

1. Create a [Namespace](https://kubernetes.io/docs/reference/kubernetes-api/cluster-resources/namespace-v1/) to contain the godbledger resources:

    ```
    kubectl apply -f ./k8s.00.namespace.yml
    ```

1. Create a [Secret]() resource to store the mysql root password and godbledger user credentials:

    ```
    kubectl apply -f ./k8s.01.mysql.secrets.yml
    ```

1. Create a [Persistent Volume](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/persistent-volume-v1/) and [Persistent Volume Claim](https://kubernetes.io/docs/reference/kubernetes-api/config-and-storage-resources/persistent-volume-claim-v1/) to store the mysql data files:

    ```
    kubectl apply -f ./k8s.02.mysql.volume.yml 
    ```

    A PersistentVolume (PV) is "a storage resource provisioned by an administrator".  It maps some storage location from outside the cluster to be accessable to (i.e. mountable by) workloads running inside the cluster.

    Persistent storage is important for a database server running inside kubernetes because by design, most resources in kubernetes are ephemeral and can be destroyed and recreated automatically by the orchestration layer for a variety of reasons.  By storing the msyql data files outside the ephemeral container you ensure that it is more durable and will survive across mysql pod restarts.

    A PersistentVolumeClaim (PVC) is "a user's request for and claim to a Persistent Volume".
    
    In this case we are creating a `mysql-pv-claim` claim for the `mysql` server to own the `mysql-pv-volume` volume.

1. Create the mysql server deployment:

    ```
    kubectl apply -f ./k8s.03.mysql.yml
    ```

    This creates a `mysql` [Deployment](https://kubernetes.io/docs/reference/kubernetes-api/workloads-resources/deployment-v1/) and `mysql` [Service](https://kubernetes.io/docs/reference/kubernetes-api/services-resources/service-v1/) resource which is listening inside the cluster on `mysql:3306`.

    Root password and credentials for the `godbledger` user are pulled dynamically from the `mysql-creds` secret resources configured in step 3.

    The `nodePort:30036` setting binds the service to receive traffic directed to port `30036` on the host node (i.e. `localhost:30036` from your host machine).
    
    The godbledger user credentials and This `nodePort` value is configured also in `k8s.config.toml` to allow `reporter` running on the host machine to connect directly to the mysql server:

    ```toml
    DatabaseType = "mysql"
    DatabaseLocation = "godbledger:password@tcp(localhost:30036)/ledger?charset=utf8mb4,utf8"
    ```

    ```toml
    Host = "127.0.0.1"
    RPCPort = "30080"
    ```

1. Create the `godbledger` Deployment and `godbledger` Service:

    ```
    kubectl apply -f ./k8s.04.godbledger.yml
    ```

    This creates a `godbledger` deployment and service able to receive traffic internally via cluster DNS at `godbledger:8080` and over the nodePort on port `30080`.

    This `nodePort` value is configured also in `k8s.config.toml` to allow `ledger_cli` running on the host machine to connect directly to the godbledger server:

    ```toml
    Host = "127.0.0.1"
    RPCPort = "30080"
    ```

1. Run apps locally and connect to your kubernetes deployments:

    - build the apps from source (if you don't have them installed locally):

        ```
        make build-native
        ```

    - `ledger_cli` connects to the godbledger server API over gRPC:

        ```
        ledger_cli --config ./k8s.config.toml
        ```

    - `reporter` connects directly to the database:

        ```
        reporter --config ./k8s.config.toml
        ```