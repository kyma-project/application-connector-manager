# Create a New Application

> [!NOTE]
> An Application represents a single connected external solution.

## Prerequisites

Before you start, export the name of your application as an environment variable:

```bash
export APP_NAME={YOUR_APP_NAME}
```

> [!NOTE]
> Read about the [Purpose and Benefits of Istio Sidecar Proxies](https://kyma-project.io/#/istio/user/00-00-istio-sidecar-proxies?id=purpose-and-benefits-of-istio-sidecar-proxies). Then, check how to [Enable Istio Sidecar Proxy Injection](https://kyma-project.io/#/istio/user/tutorials/01-40-enable-sidecar-injection). For more details, see [Default Istio Configuration](https://kyma-project.io/#/istio/user/00-15-overview-istio-setup) in Kyma.

## Create an Application

To create a new Application, run this command:

```bash
cat <<EOF | kubectl apply -f -
apiVersion: applicationconnector.kyma-project.io/v1alpha1
kind: Application
metadata:
  name: $APP_NAME
spec:
  description: Application description
  labels:
    region: us
    kind: production
EOF
```

## Get the Application Data

To get the data of the created Application and show the output in the `yaml` format, run this command:

   ```bash
   kubectl get app $APP_NAME -o yaml
   ```

A successful response returns the Application custom resource with the specified name.
This is an example response:

   ```yaml
   apiVersion: applicationconnector.kyma-project.io/v1alpha1
   kind: Application
   metadata:
     clusterName: ""
     creationTimestamp: 2018-11-22T13:53:20Z
     generation: 1
     name: test1
     namespace: ""
     resourceVersion: "30728"
     selfLink: /apis/applicationconnector.kyma-project.io/v1alpha1/applications/test1
     uid: f8ca5595-ee5d-11e8-acb2-000d3a443243
   spec:
     description: {APP_DESCRIPTION}
     labels:
       kind: "production"
       region: "us"
   ```

If there are registered services connected to your Application in Kyma, the response also shows them:

```yaml
...
spec:
  description: {APP_DESCRIPTION}
  labels:
    kind: "production"
    region: "us"
  services: {LIST_OF_REGISTERED_SERVICES}
```

> [!TIP]
> You can use Kyma dashboard to create and manage your Application. To do so, go to **Integration > Applications** from the **Cluster Details** view.
