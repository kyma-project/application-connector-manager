# Network Policies

Learn about the network policies for the Application Connector module.

## Default Network Policies

To increase security, the Application Connector module creates network policies that control traffic to and from its Pods. All the policies are enabled by default and cannot be disabled.

The following tables list all network policies for the Application Connector module. Each table presents network polices for a particular Application Connector module component. The core policies apply to all Application Connector module components, namely Application Connector Manager, Compass Runtime Agent, Central Application Gateway, and Central Application Connectivity Validator.

<!-- on Help Portal, all network policies are in one table as the Help Portal documentation for the Application Connector module is limited and doesn't introduce detailed architecture with the module components --!>

**Core Policies**

These policies apply to all Application Connector module components (Application Connector Manager, Compass Runtime Agent, Central Application Gateway, and Central Application Connectivity Validator):

| Policy Name | Type | Description |
|-------------|------|-------------|
| `kyma-project.io--acm-module-to-api-server` | Egress | Allows egress from the Application Connector module Pods to any destination on TCP port 443, such as the Kubernetes API server) |
| `kyma-project.io--acm-module-to-dns` | Egress | Allows egress from the Application Connector module Pods to DNS services on UDP/TCP ports 53 and 8053) for cluster and external DNS resolution |
| `kyma-project.io--acm-module-allow-to-sidecar` | Egress | Allows egress from the Application Connector module Pods to Istio's istiod on TCP port 15012 for sidecar configuration |
| `kyma-project.io--acm-module-allow-metrics` | Ingress | Allows ingress to the Application Connector Manager Pods on TCP port 8080 from Pods in the `kyma-system` namespace labeled `networking.kyma-project.io/metrics-scraping: allowed` for metrics scraping |
| `kyma-project.io--acm-module-allow-to-external-system` | Egress | Allows egress from the Compass Runtime Agent and Central Application Gateway Pods to external systems at 0.0.0.0/0, excluding link-local addresses |

**Central Application Gateway Policies**

These policies apply specifically to the Central Application Gateway component:

| Policy Name | Type | Description |
|-------------|------|-------------|
| `kyma-project.io--acm-gateway-allow-from-cluster-workload` | Ingress | Allows ingress to the Central Application Gateway Pods on TCP ports 8080 and 8082 from any source within the cluster |
| `kyma-project.io--acm-gateway-allow-from-health-check` | Ingress | Allows ingress to the Central Application Gateway Pods on TCP port 8081 for health checks |
| `kyma-project.io--acm-gateway-allow-to-external-system` | Egress | Allows egress from the Central Application Gateway Pods to external systems on TCP port 443 |

**Central Application Connectivity Validator Policies**

| Policy Name | Type | Description |
|-------------|------|-------------|
| `kyma-project.io--acm-validator-allow-from-istio-ingressgateway` | Ingress | Allows ingress to the Validator Pods on TCP port 8080 from Istio Ingress Gateway in the `istio-system` namespace |
| `kyma-project.io--acm-validator-allow-from-health-check` | Ingress | Allows ingress to the Validator Pods on TCP port 8081 for health checks |
| `kyma-project.io--acm-validator-allow-to-eventing` | Egress | Allows egress from the Validator Pods to the Eventing Publisher Proxy on TCP port 8080 |

**Compass Runtime Agent Policies**

| Policy Name | Type | Description |
|-------------|------|-------------|
| `kyma-project.io--acm-compass-allow-from-health-check` | Ingress | Allows ingress to the Compass Runtime Agent Pods on TCP port 8090 for health checks |

## Verify Status

To check if the network policies are active, run the following command:

```bash
kubectl get networkpolicies -n kyma-system -l kyma-project.io/module=application-connector
```
