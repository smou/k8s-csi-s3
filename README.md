# CSI for S3

This is a Container Storage Interface ([CSI](https://github.com/container-storage-interface/spec/blob/master/spec.md)) for S3 (or S3 compatible) storage. This can dynamically allocate buckets and mount them via a fuse mount into any container.
It implements the same mouting mechanism as the csi driver from aws and is not using any third party fs implementation (e.g. s3-fs)

## Kubernetes installation

### Requirements

* Kubernetes 1.33+
* Kubernetes has to allow privileged containers
* Docker daemon must allow shared mounts (systemd flag `MountFlags=shared`)

### Helm chart
//TODO Not available yet
Helm chart is published at `https://smou.github.io/k8s-csi-s3`:

```
helm repo add minio-csi-s3 https://smou.github.io/k8s-csi-s3/charts

helm install csi-s3 yandex-s3/csi-s3
```

### Configuration

The driver will load the configmap and secret via the kubernetes api and requires related permissions to be able to access the kubernetes resources

| Name             | Type         | Required | Default      | Description                                         |
| :--------------- | :----------- | :------: | :----------- | :-------------------------------------------------- |
| MINIO_ENDPOINT   | ConfigMap    | True     | -            | URL of the targeting minio instance. Https will automatically enable TLS |
| MINIO_REGION     | ConfigMap    | False    | us-east-1    | S3 region of the bucket. For compatibility only. Take no effekt for minio |
| MINIO_BUCKET_PREFIX | ConfigMap | False    | ""           | prefix for each bucket name. The bucket name will be equal to volume id 'pvc-UUID' |
| MINIO_ACCESSKEY  | Secret | True | - | Equal to AWS_ACCESS_KEY_ID |
| MINIO_SECRETKEY | Secret | True | - | Equal to AWS_SECRET_ACCESS_KEY |
| NAMESPACE | Env | True | "" | Namespace where is loading the configmap and secret from |
| CONFIGMAP_NAME | Env | True | "" | Name of the config map |
| SECRET_NAME | Env | True | "" | Name of the secret |

### Manual installation

#### 1. Create a secret with your S3 credentials

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: csi-s3-secret
  namespace: csi-s3
type: Opaque
stringData:
  MINIO_ACCESSKEY: <YOUR_ACCESS_KEY_ID>
  MINIO_SECRETKEY: <YOUR_SECRET_ACCESS_KEY>
```

#### 2. Updating ConfigMap with your configuration

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: csi-s3-config
  namespace: csi-s3
  labels:
    app.kubernetes.io/part-of: minio-csi-s3
data:
  MINIO_ENDPOINT: "https://minio.mydomain.com"
```

#### 3. Deploy the driver

```bash
kubectl apply -k ./k8s
```

##### Upgrading

If you're upgrading from yandex-cloud/k8s-csi-s3 - delete all resources:
- Deployment
- DeamonSet
- StorageClass
- CSIDriver
- RBAC (or Update)
  - ServiceAccount
  - ClusterRole
  - ClusterRoleBindings
Migrate Config and Secrets

#### 3. Create the storage class

```bash
kubectl create -f examples/storageclass.yaml
```

#### 4. Test the S3 driver

1. Create a pvc using the new storage class:

    ```bash
    kubectl apply -f k8s/test/pvc.yaml
    ```

1. Check if the PVC has been bound:

    ```bash
    $ kubectl get pvc s3-pvc
    NAME         STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
    s3-pvc       Bound     pvc-c5d4634f-8507-11e8-9f33-0e243832354b   1Gi        RWO            s3-standard    9s
    ```

1. Create a test pod which mounts your volume:

    ```bash
    kubectl apply -f k8s/test/pod.yaml
    ```

    If the pod can start, everything should be working.

1. Test the mount

    ```bash
    $ kubectl exec -ti csi-s3-test-nginx bash
    $ mount | grep fuse
    pvc-035763df-0488-4941-9a34-f637292eb95c: on /usr/share/nginx/html/s3 type fuse.geesefs (rw,nosuid,nodev,relatime,user_id=65534,group_id=0,default_permissions,allow_other)
    $ touch /usr/share/nginx/html/s3/hello_world
    ```

If something does not work as expected, check the troubleshooting section below.

## Additional configuration

### Bucket

minio-csi-s3 will create a new bucket per volume. The bucket name will match that of the volume ID. If you want to have the bucketname prefixed by a custom value
you need to set 'MINIO_BUCKET_PREFIX'

### Static Provisioning

If you want to mount a pre-existing bucket or prefix within a pre-existing bucket and don't want csi-s3 to delete it when PV is deleted, you can use static provisioning.

To do that you should omit `storageClassName` in the `PersistentVolumeClaim` and manually create a `PersistentVolume` with a matching `claimRef`, like in the following example: [deploy/kubernetes/examples/pvc-manual.yaml](deploy/kubernetes/examples/pvc-manual.yaml).

### Mounter

We **strongly recommend** to use the default mounter which is [GeeseFS](https://github.com/smou/geesefs).

However there is also support for two other backends: [s3fs](https://github.com/s3fs-fuse/s3fs-fuse) and [rclone](https://rclone.org/commands/rclone_mount).

The mounter can be set as a parameter in the storage class. You can also create multiple storage classes for each mounter if you like.

As S3 is not a real file system there are some limitations to consider here.
Depending on what mounter you are using, you will have different levels of POSIX compability.
Also depending on what S3 storage backend you are using there are not always [consistency guarantees](https://github.com/gaul/are-we-consistent-yet#observed-consistency).

You can check POSIX compatibility matrix here: https://github.com/smou/geesefs#posix-compatibility-matrix.

## Troubleshooting

### Issues while creating PVC

Check the logs of the controller:

```bash
kubectl logs -l app.kubernetes.io/part-of=minio-csi-s3 -c csi-s3-controller
```

### Issues creating containers

1. Ensure feature gate `MountPropagation` is not set to `false`
2. Check the logs of the s3-driver:

```bash
kubectl logs -l app.kubernetes.io/part-of=minio-csi-s3 -c csi-s3
```

## Development

This project can be built like any other go application.

```bash
go get -u github.com/smou/k8s-csi-s3
```

### Build executable

```bash
make build
```

### Tests

```bash
make test
```
