# PlenusLB

![PlenusLB Logo](img/logo.png "PlenusLB Logo")

## Description

PlenusLB is a bare metal and cloud load balancer for Kubernetes clusters.
It is designed for environments where a cloud load balancer already integrated for Kubernetes is not available.

The current version support two scenarios:
- pure bare metal deployment
- integration with Hetzner cloud

PlenusLB has been originally developed to be used on the [Plenus cloud platform](https://plenus.cloud) and in bare metal environments.

## How it works

PlenusLB works by taking IP addresses, defined in a IPPool custom resource, and assigning them to kubernetes services of type LoadBalancer.
The allocation is managed with a IPAllocation custom resource; once assigned the IP, the status of the service is updated, then Kubernetes will create the iptables/ipvs rules to route the ingress traffic to the service.

PlenusLB also takes care of:
- when the IP pool declares a cloud provider, requesting the IP address to that cloud provider and, on the provider side, direct the routing to the server where the IP will be assigned
- choosing a cluster node to act as ingress node
- assigning the IP address to a given network interface of the node, this way the node will accept traffic for the IP; usually the interface is an empty bridge.

## Prerequisites

For the bare metal scenario it will be necessary to reserve a pool of IP addresses for each cluster/PlenusLB: the addresses must be declared in a PersistentIPPool.

### Hetzner

To use PlenusLB with the Hetzner cloud provider you will need to have a project active on the cloud, create an API key in the "API TOKENS" sections of the interface and specify this token in the IP pools. The kubernetes cluster where PlenusLB is operating needs to be in the same Hetzner cloud project.

### Dedicated bridge interface

All cluster nodes need to have an interface which can be used to assign IP addresses to.
Since on that interface PlenusLB will remove all IP addresses during operator startup on each node, **do not use** any interface where normal IP addresses have been assigned to the node.
We strongly suggest to use a bridge specifically created for this purpose, for example ```pl0```

On Ubuntu/Debian nodes an empty (without physical interfaces) bridge ```pl0``` can be created with the following commands:

```
apt-get update && \
apt-get -q -y install bridge-utils bash && \
bash -c \"echo -e 'auto pl0\niface pl0 inet manual\n  bridge_ports none\n  bridge_stp off\n  bridge_fd 0\n  bridge_maxwait 0' > /etc/network/interfaces.d/90-bridge-pl0.cfg\" && \
/etc/init.d/networking restart
```

The bridge will be persistent and will be restarted automatically after reboots, it is sufficient to give the commands once.

Using cloud-init put the following lines in cloud-config:

```
#cloud-config
bootcmd:
 - [ cloud-init-per, once, plenuslb-init, sh, -xc, "apt-get update && apt-get -q -y install bridge-utils bash && bash -c \"echo -e 'auto pl0\niface pl0 inet manual\n  bridge_ports none\n  bridge_stp off\n  bridge_fd 0\n  bridge_maxwait 0' > /etc/network/interfaces.d/90-bridge-pl0.cfg\" && /etc/init.d/networking restart" ]
```

## Install and upgrade

PlenusLB can be installed with the helm chart in the Plenus helm chart repository.
Installation has been tested with helm 3.x and it is assumed a namespace plenuslb has been already created.

```yaml
helm repo add plenus https://plenus-charts.storage.googleapis.com/stable/
helm repo update
helm upgrade --install \
  --namespace=plenuslb \
  --set envs.CLUSTER_NAME=mycluster \
  --atomic --wait plenuslb plenus/plenuslb
```

The same command can be used to upgrade PlenusLB.

The value specified in the CLUSTER_NAME variable will be used as a part of the name given to the ephemeral IP addresses created on a cloud provider.
Set it to a value related to the cluster where PlenusLB is installed, so that the IP addresses can be easily identified with the cluster.

## Types of IP pool

PlenusLB supports two types of IP pools:
- ephemeral IP pools, useful when the life cycle of the IP follows the life cycle of the service
- persistent IP pools, targeted for all those cases where a static reservation of the IP is necessary

At the moment it is not possibile to create an IP address with an ephemeral IP pool and then migrate it to a persistent IP pool.

### Ephemeral IP

By creating a service with type: LoadBalancer and not specifying any externalIPs PlenusLB will provision an ephemeral IP:
the IP will be assigned to the service as long as the service exists, but there is no reservation; if the service is deleted the IP will be
released on the cloud provider. Ephemeral IP addresses cannot be used in the bare metal scenario.

To use ephemeral IP addresses it is necessary to create an EphemeralIPPool:

```
apiVersion: loadbalancing.plenus.io/v1alpha1
kind: EphemeralIPPool
metadata:
  name: hetzner-eph-pool-all
spec:
  cloudIntegration:
    hetzner:
      token: YOUR_HETZNER_API_TOKEN
  options:
    hostNetworkInterface:
      addAddressesToInterface: true
      interfaceName: pl0
```

```cloudIntegration``` declares the cloud provider where plenuslb will create the IP addresses. At the moment only ```hetzner``` is supported, and accepts a single parameter ```token``` which must contain an Hetzner API key; the IP addresses will be created in the project that the API keys are authorized for, this must be the same project where the kubernetes cluster has been created.

```options.hostNetworkInterface.interfaceName``` must be set to the interface name where PlenusLB will assign IP addresses.

To have an ephemeral IP assigned create a service with type: LoadBalancer and no externalIPs

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-kubernetes
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: hello-kubernetes
```

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-kubernetes
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hello-kubernetes
  template:
    metadata:
      labels:
        app: hello-kubernetes
    spec:
      containers:
      - name: hello-kubernetes
        image: paulbouwer/hello-kubernetes:1.8
        ports:
        - containerPort: 8080
```

### Persistent IP

Persistent IP addresses can be used in all those cases where a reservation for the IP is desiderable, regardless where there is a service requesting the IP or not.
Moreover at the moment it is the only type of IP supported in the bare metal scenario.

The idea is that the pool of available IP addresses is preassigned:
- in the bare metal scenario a pool of IP addresses is reserved for the kubernetes cluster
- on a cloud provider the IP addresses will be manually acquired, for example on Hetzner you will have to buy the IP addresses in the project accesible to PlenusLB throught the given token

Then it is necessary to create a PersistentIPPool containing the addresses.

The following example is for Hetzner:

```yaml
apiVersion: loadbalancing.plenus.io/v1alpha1
kind: PersistentIPPool
metadata:
  name: hetzner-persist-pool-all
spec:
  addresses:
    - "1.2.3.4"
    - "1.2.3.5"
  cloudIntegration: 
    hetzner:
      token: YOUR_HETZNER_API_TOKEN
  options:
    hostNetworkInterface: 
      addAddressesToInterface: true
      interfaceName: pl0
```

For the bare metal case omit the cloudIntegration section.

Every persistent IP can be bound to a single service.

To have PlenusLB assign the persistent IP to a service it is sufficient to specify it as an externalIPs in a LoadBalancer type service.

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-kubernetes-persistent
spec:
  type: LoadBalancer
  externalIPs:
  - 1.2.3.4
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: hello-kubernetes
```

Where 1.2.3.4 is the IP that you want assigned to the service.  

## Multitenancy

PlenusLB provides some degrees of multitenancy: if a cluster has multiple users, each one of them confined to a set of namespaces, it it possibile to create IP pools reserved for specific namespaces. This, combined with the use of persistent IP pools, allows to allocate some IP addresses for specific users/projects.

### Allowed namespaces

The following PersistentIPPool declares two IP addresses that can be requested only from services in the namespaces project1 and project2.


```yaml
apiVersion: loadbalancing.plenus.io/v1alpha1
kind: PersistentIPPool
metadata:
  name: hetzner-persist-pool-reserved
spec:
  allowedNamespaces:
    - project1
    - project2
  addresses:
    - "1.2.3.6"
    - "1.2.3.7"
  options:
    hostNetworkInterface: 
      addAddressesToInterface: true
      interfaceName: pl0
```

The same notation can be used with EphemeralIPPool. It could be useful when there are multiple projects on the same cluster and only some project must be allowed to request IP addresses from the cloud provider.

## Build

To build a specified tag:

```
git checkout v0.3.4
./build.sh -v v0.3.4
```
