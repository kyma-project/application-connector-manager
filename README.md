[![REUSE status](https://api.reuse.software/badge/github.com/kyma-project/application-connector-manager)](https://api.reuse.software/info/github.com/kyma-project/application-connector-manager)

# Application Connector Manager TEST

## Status

![GitHub tag checks state](https://img.shields.io/github/checks-status/kyma-project/application-connector-manager/main?label=application-connector-operator&link=https%3A%2F%2Fgithub.com%2Fkyma-project%2Fapplication-connector-manager%2Fcommits%2Fmain)

## Overview

This repository contains the following projects:

- **Application Connector Manager** - operator compatible with Lifecycle Manager that manages the Application Connector module in Kyma.
- **Application Connector Module build configuration** - delivers the functionality of Application Connector as a Kyma module.

> **NOTE:** Docker images for the Application Connector binaries are built separately from the main GitHub [Kyma repository](https://github.com/kyma-project/kyma/).
 
## How It Works 
 
The Application Connector module represents a specific version of the Application Connector binaries.
The configuration of the released module is described as a ModuleTemplate custom resource (CR) and delivered as an OCI image.
You can install it in the Kyma cluster managed by Template Operator.

The installed Application Connector module is represented as the ApplicationConnector Kubernetes CR.

```yaml
apiVersion: operator.kyma-project.io/v1alpha1
kind: ApplicationConnector
metadata:
  labels:
    app.kubernetes.io/name: applicationconnector
    app.kubernetes.io/instance: applicationconnector-sample
  name: applicationconnector-sample
spec: {}
```

Any update to this CR is intercepted by Application Connector Manager and applied to the Application Connector binaries.

See also:
- [Lifecycle Manager documentation](https://github.com/kyma-project/lifecycle-manager#lifecycle-manager)
- [Application Connector module documentation](docs/user/README.md) 
- [Modularization of Kyma](https://kyma-project.io/#/?id=kyma-modules)
