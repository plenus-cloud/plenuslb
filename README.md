# NOTES - TO BE REMOVED

https://itnext.io/how-to-generate-client-codes-for-kubernetes-custom-resource-definitions-crd-b4b9907769ba


kubectl patch daemonset plenuslb-operator -p '{"spec": {"template": {"spec": {"nodeSelector": {"non-existing": "true"}}}}}'

kubectl patch daemonset plenuslb-operator --type json -p='[{"op": "remove", "path": "/spec/template/spec/nodeSelector/non-existing"}]'

# PLENUSLB

## Description

PlenusLB is a bare metal and cloud load balancer for Kubernetes clusters.
It is designed for environments where it is not available a cloud load balancer already integrated for Kubernetes.

The current version support two scenarios:
- pure bare metal deployment
- integration with Hetzner cloud

## How it works

PlenusLB works by taking ip addresses, defined in a IPPool custom resource, and assigning them to kubernetes services of type LoadBalancer.
The allocation is managed with a IPAllocation custom resource; once assigned the ip the status of the service is updated, at this point Kubernetes will create the iptables/ipvs rules to route the ingress traffic to the service.

PlenusLB also take care of:
- if the ip pool declares a cloud provider then request an ip address to that cloud provider and, on the provider size, direct the routing to the server where the ip will be assigned
- choosing a cluster node to act as ingress node
- assigning the ip address to a given network interface of the node

## Build

To build and push a specified tag:

```
git checkout v0.2.1
./build.sh -v v0.2.1 -p
```
