# `Endpoint`

## Overview

> `Endpoint` defines connection details and credentials for accessing an external API or service. 

It is stored as a Kubernetes [`Secret`](https://kubernetes.io/docs/concepts/configuration/secret/) and consumed by [`RESTAction`](restactions.md) to establish secure HTTP or HTTPS connections.

## `Endpoint` keys

| Key | Type | Description | Required |
|--------|------|-------------|-----------|
| `server-url` | `string` | Base URL of the target API or service. | ✅ |
| `proxy-url` | `string` | Optional proxy address used for outbound HTTP traffic. | ❌ | 
| `token` | `string` | Bearer token for authentication. | ❌ |
| `username` | `string` | Username for basic authentication. | ❌ |
| `password` | `string` | Password for basic authentication. | ❌ |
| `certificate-authority-data` | `string (base64)` | Base64-encoded PEM CA certificate used to verify the remote server. | ❌ |
| `client-certificate-data` | `string (base64)` | Base64-encoded PEM client certificate for mutual TLS authentication. | ❌ |
| `client-key-data` | `string (base64)` | Base64-encoded PEM private key associated with the client certificate. | ❌ |
| `debug` | `boolean` | Enables verbose logging of HTTP requests/responses. | ❌ |
| `insecure` | `boolean` | If `true`, disables TLS certificate verification. | ❌ |

## Storage in Kubernetes

`Endpoint` information is stored as a [`Secret`](https://kubernetes.io/docs/concepts/configuration/secret/).  

The secret’s `.data` or `.stringData` must contain the fields above with the exact key names.

Example:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-endpoint
  namespace: default
stringData:
  server-url: https://api.example.com
  proxy-url: http://proxy.internal:8080
  token: "abc123"
  username: "admin"
  password: "s3cret"
  certificate-authority-data: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCg=="
  client-certificate-data: "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCg=="
  client-key-data: "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQo="
  debug: "true"
  insecure: "false"
```

🔎 Note: Certificate and key values must be base64-encoded strings.

## Behavior Summary

- `server-url` is mandatory
- certificates and keys are expected to be base64-encoded, not raw PEM blocks
- when insecure is true, the TLS client skips server certificate validation
- if both clientCertificateData and clientKeyData are present, mutual TLS is enabled
- if certificateAuthorityData is provided, it’s added to the root CA pool
- when a proxy URL is set, outbound requests are routed through it
- boolean values (debug, insecure) are parsed from strings

