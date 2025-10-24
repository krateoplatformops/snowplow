# `RESTAction`

**API Group:** `templates.krateo.io`  
**Kind:** `RESTAction`  
**Version:** `v1`  
**Scope:** Namespaced  

## Overview

The `RESTAction` is a Krateo PlatformOps resource that enables users to **declaratively define one or more REST API calls** within Kubernetes.

It allows you to chain HTTP requests, handle dependencies between them, extract data, and use filters to process results — all through a Kubernetes-native manifest.

This approach is particularly useful for integrating external systems or Kubernetes APIs into workflows managed by Krateo PlatformOps.

> `RESTAction` defines one or more declarative HTTP (REST) calls that can optionally depend on other calls.

It allows you to orchestrate a chain of API requests across multiple endpoints using Kubernetes resources.

A `RESTAction` resource declaratively defines one or more HTTP calls (`spec.api`) that can depend on each other.

Each call can produce a JSON response that becomes part of a **shared global context**, enabling subsequent calls to reference previous results using **JQ expressions**, iterators, and filters.

To fully leverage these advanced capabilities — such as resolving JQ expressions, using custom JQ functions or modules, and managing interdependent API calls — the `RESTAction` must be executed through the `snowplow` service endpoint (`/call`).  

Only this endpoint implements the orchestration logic that:
- Executes all HTTP requests defined under `spec.api`, respecting their declared dependencies (`dependsOn`).
- Stores all API responses in a global JSON context.
- Evaluates and resolves any JQ expressions or iterators defined within the resource.
- Returns the computed output in the resource’s `status` field.

When a `RESTAction` is retrieved directly via Kubernetes (e.g. `kubectl get restaction <name>`), the resource is shown **as-is**, without JQ resolution or execution of any API calls.


## `spec`

The `spec` field defines the configuration for the REST action workflow.

| Field | Type | Description | Required |
|--------|------|-------------|-----------|
| `api` | `array` | List of API requests to execute. Each item defines one HTTP call. | ✅ |
| `filter` | `string` | Optional filter to apply to the overall output or results. | ❌ |


### `spec.api[]`

Defines a single HTTP request and its dependencies.

| Field | Type | Description | Required |
|--------|------|-------------|-----------|
| `name` | `string` | A unique identifier for this API call. | ✅ |
| `verb` | `string` | The HTTP method (e.g., `GET`, `POST`, `PUT`, `DELETE`). Defaults to `GET`. | ❌ |
| `path` | `string` | The URI path of the request. | ❌ |
| `payload` | `string` | The request body (for methods like `POST`, `PUT`, etc.). | ❌ |
| `headers` | `array` | Array of custom request headers to include in the request. | ❌ |
| `filter` | `string` | Optional filter to process or extract data from the response. | ❌ |
| `errorKey` | `string` | Key to identify error fields in the response. | ❌ |
| `exportJwt` | `boolean` | If `true`, exports a JWT token from this request for later use. | ❌ |
| `continueOnError` | `boolean` | If `true`, continues execution even if this call fails. | ❌ |
| `endpointRef` | `object` | Reference to a Kubernetes [`Endpoint`](endpoints.md) object defining the target service. | ✅ |
| `dependsOn` | `object` | Declares a dependency on another API call defined in this spec. | ❌ |

### `spec.api[].endpointRef`

Defines the reference to an [`Endpoint`](endpoints.md) resource that this API should call.

| Field | Type | Description | Required |
|--------|------|-------------|-----------|
| `name` | `string` | Name of the referenced [`Endpoint`](endpoints.md) object. | ✅ |
| `namespace` | `string` | Namespace of the referenced [`Endpoint`](endpoints.md) object. | ✅ |

### `spec.api[].dependsOn`

Defines a dependency on another API call within the same `RESTAction` definition.  
Useful for chaining calls where one must complete before another.

| Field | Type | Description | Required |
|--------|------|-------------|-----------|
| `name` | `string` | Name of another API call in the list that this call depends on. | ✅ |
| `iterator` | `string` | Optional field on which to iterate (used for loop-like behavior). | ❌ |

## Example

```yaml
apiVersion: templates.krateo.io/v1
kind: RESTAction
metadata:
  name: example-restaction
spec:
  api:
    - name: get-user
      endpointRef:
        name: user-endpoint
        namespace: default
      verb: GET
      path: /users
      headers:
        - "Authorization: Bearer $(TOKEN)"
      continueOnError: false

    - name: update-user
      dependsOn:
        name: get-user
      endpointRef:
        name: user-endpoint
        namespace: default
      verb: PUT
      path: /users/123
      payload: '{"status":"active"}'

  filter: ""
