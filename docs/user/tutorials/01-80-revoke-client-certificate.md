# Revoke a Client Certificate (RA)

> [!WARNING]
> Runtime Agent is currently not integrated into the Application Connector module. Proceed with caution and consult the respective documentation for the Application Connector module's current configuration and functionality.

After you have established a secure connection with Compass and generated a client certificate, you may want to revoke this certificate at some point. To revoke a client certificate, follow the steps in this tutorial.

> [!NOTE]
> A revoked client certificate remains valid until it expires, but it cannot be renewed.

## Prerequisites

- [OpenSSL toolkit](https://openssl-library.org/source/index.html) to create a Certificate Signing Request (CSR), keys, and certificates which meet high security standards
- [Compass](https://github.com/kyma-incubator/compass)
- Registered Application
- Runtime connected to Compass
- [Established secure connection with Compass](01-60-establish-secure-connection-with-compass.md)

> [!NOTE]
> See how to [maintain a secure connection with Compass and renew a client certificate](01-70-maintain-secure-connection-with-compass.md).

## Revoke the Certificate

To revoke a client certificate, make a call to the Certificate-Secured Connector URL using the client certificate.
The Certificate-Secured Connector URL is the `certificateSecuredConnectorURL` obtained when establishing a secure connection with Compass.
Send this mutation with the call:

```graphql
mutation { result: revokeCertificate }
```

A successful call returns the following response:

```json
{"data":{"result":true}}
```
