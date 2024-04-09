# Application Connector Module

## What is Application Connectivity in Kyma?

Application Connectivity in Kyma is an area that:

- Simplifies and secures the connection between external systems and Kyma
- Stores and handles the metadata of external systems
- Provides certificate handling for the [Eventing](https://kyma-project.io/#/eventing-manager/user/README) flow in the Compass scenario (mode)
- Manages secure access to external systems
- Provides monitoring and tracing capabilities to facilitate operational aspects

Depending on your use case, Application Connectivity works in one of the two modes:

- **Standalone mode** (default) - a standalone mode where Kyma is not connected to [Compass](https://github.com/kyma-incubator/compass)
- **Compass mode** - using [Runtime Agent](00-30-runtime-agent-overview.md) and integration with [Compass](https://github.com/kyma-incubator/compass) to automate connection and registration of services using mTLS certificates

> [!WARNING]
> Runtime Agent is currently not integrated into the Application Connector module. Proceed with caution and consult the respective documentation for the Application Connector module's current configuration and functionality.

## Application Connector Module

The Application Connector module allows you to install, uninstall, and configure all features of Kyma's Application Connector on your Kubernetes cluster using the Application Connector Manager service.

Application Connector allows you to connect your workflows deployed on Kyma with external solutions. No matter if you want to integrate an on-premise or a cloud system, the integration process does not change, which allows you to avoid any configuration or network-related problems.

The external solution you connect to Kyma using Application Connector is represented as an Application. There is always a one-to-one relationship between a connected solution and an Application, which helps to ensure the highest level of security and separation. This means that you must create five separate Applications in your cluster to connect five different external solutions and use their APIs and event catalogs in Kyma.

Application Connector secures Eventing with a client certificate verified by Istio Ingress Gateway in the Compass scenario.

### Features

Application Connector:

- Simplifies and secures the connection between external systems and Kyma
- Stores and handles the metadata of external APIs
- Proxies calls sent from internals of the Kyma cluster (for example, from [Function](https://kyma-project.io/#/serverless-manager/user/resources/06-10-function-cr)) to external APIs registered as the connected external solution (Application)
- Provides certificate handling for the [Eventing](https://kyma-project.io/#/eventing-manager/user/README) flow in the Compass scenario
- Delivers events from the connected external solution to the internal Kyma Event Publisher in the Compass scenario
- Manages secure access to external systems

All the Application Connector components scale independently, which allows you to adjust it to fit the needs of the implementation built using Kyma.

### Supported APIs

Application Connector supports secured REST APIs exposed by the connected external solution. Application Connector supports a variety of authentication methods to ensure smooth integration with a wide range of APIs.

The following authentication methods for your secured APIs are supported:

- Basic Authentication
- OAuth
- OAuth 2.0 mTLS
- Client Certificates

> [!NOTE]
> Non-secured APIs are supported too, however, they are not recommended in the production environment.

In addition to authentication methods, Application Connector supports Cross-Site Request Forgery (CSRF) Tokens.

Application Connector supports any API that adheres to the REST principles and is available over the HTTP protocol.

## Application Connector Manager

When you enable the Application Connector module, Application Connector Manager takes care of installation and configuration of all the Application Connector module components on your cluster. It manages the lifecycle of the Application Connector module based on the dedicated ApplicationConnector custom resource (CR).

## Useful Links

If you want to perform some simple and more advanced tasks, check the [Application Connectivity tutorials](tutorials/README.md).

To learn more about the architecture, the configuration parameters, and any other references, visit [technical reference](technical-reference/README.md).

For more information on the CRs, see [Resources](resources/README.md).
