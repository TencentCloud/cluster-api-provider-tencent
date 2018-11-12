This is tencent cloud provider for [cluster-api](https://github.com/kubernetes-sigs/cluster-api)

## BUILD
#### Build Binaries
There're 3 binaries:
- clusterapi-generic-controller: machinedeployment controller and machineset controller. No depends on tencent provider.
- tke-cluster-controller: cluster controller for [TKE](https://cloud.tencent.com/product/tke)
- tke-machine-controller: machine controller for [TKE](https://cloud.tencent.com/product/tke)

To build these binaries, just type:
```
make bin
```

The binaries will be under **output/bin/**.

#### Build Docker Image
To build docker images from the binaries, try:
```
REPO=ccr.ccs.tencentyun.com/ccs-dev TAG=0.2 make img
```
3 images will be produced by above command:
- ccr.ccs.tencentyun.com/ccs-dev/clusterapi-generic-controller:0.2
- ccr.ccs.tencentyun.com/ccs-dev/tke-cluster-controller:0.2
- ccr.ccs.tencentyun.com/ccs-dev/tke-machine-controller:0.2

#### Build Yaml
To generate yaml files using above docker images, try:
```
REPO=ccr.ccs.tencentyun.com/ccs-dev TAG=0.2 make yaml
```

The generated yaml files will be under **output/yaml**

Usally we only need output/yaml/clusterapi-controllers-all-in-one.yaml, whick defines 3 deployments.

## Install on a TKE cluster
#### 1. Install CRDs
```
kubectl apply -f config/crds/clusterapi-crds-all-in-one.yaml
```

check:
```
# kubectl get CustomResourceDefinition
NAME                                    AGE
clusters.cluster.k8s.io                 1m
machines.cluster.k8s.io                 1m
machinesets.cluster.k8s.io              1m
machinedeployments.cluster.k8s.io       1m
```

#### 2. prepare and install secret

There's a templet yaml file for the secret:
```
# cat config/yaml/tencent-cloud-api-secret.yaml 
apiVersion: v1
kind: Secret
metadata:
  name: tencent-cloud-api-secret
type: Opaque
data:
  SecretId: '!!! Get your SecretId from https://console.cloud.tencent.com/cam/capi'
  SecretKey: '!!! Get your SecretKey from https://console.cloud.tencent.com/cam/capi'
```

Go to [Tencent Cloud API Token](https://console.cloud.tencent.com/cam/capi) to get your effective SecretId and SecretKey, fill them into the config/yaml/tencent-cloud-api-secret.yaml, and then `kubectl apply -f config/yaml/tencent-cloud-api-secret.yaml`


#### 3. install controllers
```
kubectl apply -f output/yaml/clusterapi-controllers-all-in-one.yaml
```

check:
```
# kubectl get deployment
NAME                            DESIRED   CURRENT   UP-TO-DATE   AVAILABLE   AGE
clusterapi-generic-controller   1         1         1            1           3d
tke-cluster-controller          1         1         1            1           3d
tke-machine-controller          1         1         1            1           3d
```

Now this TKE cluster is cluster-api enabled.

## Using cluster-api
#### Cluster resource
Modify the parameters in config/samples/cluster.yaml.
```
# cat config/samples/cluster.yaml
apiVersion: "cluster.k8s.io/v1alpha1"
kind: Cluster
metadata:
  name: test-tke-cluster
spec:
    clusterNetwork:
        services:
            cidrBlocks: ["10.96.0.0/12"]
        pods:
            cidrBlocks: ["192.168.0.0/16"]
        serviceDomain: "cluster.local"
    providerConfig:
      value:
        apiVersion: "tkeproviderconfig/v1alpha1"
        kind: "TKEClusterProviderConfig"
        clusterName: "cluster-test3"
        clusterCIDR: "172.30.0.0/19"
        clusterVersion: "1.10.5"
        vpcId: "vpc-iueiposq"
        region: "ap-beijing"
```

The create the cluster resource: 
```
kubectl apply -f config/samples/cluster.yaml
```

check: 
```
kubectl get clusters
```

Wait until the status turns from "Creating" to "Created", which means the cluster is ready.

#### machinedeployment resource
Modify the parameters in config/samples/machinedeployment.yaml.
```
# cat config/samples/machinedeployment.yaml
apiVersion: "cluster.k8s.io/v1alpha1"
kind: MachineDeployment
metadata:
  name: sample-machinedeployment
spec:
  replicas: 2
  selector:
    matchLabels:
      foo: bar
  template:
    metadata:
      labels:
        foo: bar
    spec:
      providerConfig:
        value:
          apiVersion: "tkeproviderconfig/v1alpha1"
          kind: "TKEMachineProviderConfig"
          zoneId: "800001"
          cpu: 1
          mem: 2
          bandwidthType: "PayByTraffic"
          bandwidth: 1
          subnetId: "subnet-nzi3a453"
          storageSize: 50
          rootSize: 20
          password: "123456789!"
          isVpcGateway: 0
          wanIp: 1
          osName: "ubuntu16.04.1 LTSx86_64"
      versions:
        kubelet: 1.10.5
  strategy:
    type: "RollingUpdate"
    rollingUpdate:
      maxUnavailable: "30%"
      maxSurge: "30%"
  minReadySeconds: 2
```

#### region
Get the region name from [Tencent Cloud Regions](https://cloud.tencent.com/document/api/213/15692#.E5.9C.B0.E5.9F.9F.E5.88.97.E8.A1.A8)

#### zoneId
Get the zoneId from [Tencent Cloud ZoneIds](https://cloud.tencent.com/document/api/213/1286)

#### vpcId
Get vpcId from [vpc console](https://console.cloud.tencent.com/vpc/vpc)

#### subnetId
Get subnetId from [subnet console](https://console.cloud.tencent.com/vpc/subnet)

---

Then create the machinedeployment resource, wait the machines ready.  You can scale up and down the machinedeployment, and then delete it.
