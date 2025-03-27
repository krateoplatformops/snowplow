# RESTAction 

> This document provides an overview of the `RESTAction` CRD and its properties to facilitate its usage within Kubernetes environments.

## Overview
The `RESTAction` Custom Resource Definition (CRD) allows users to declaratively define calls to APIs that may depend on other API calls.

## Schema `spec` Details

| Property  | Type  | Description |
|-----------|-------|-------------|
| `api` | array | Defines API requests to an HTTP service. |
| `filter` | string | A JQ filter that can be applied to the global response. |

#### `api` Array Item Properties

> A single `api` item defines an HTTP REST call. 
> The invoked API **must produce a `JSON` content type**

| Property  | Type  | Description |
|-----------|-------|-------------|
| `dependsOn` | object | Defines dependencies on other APIs. |
| `endpointRef` | object | References an Endpoint object. |
| `filter` | string | A JQ expression for response processing. |
| `headers` | array of strings | Custom request headers (each header can be a JQ expression). |
| `name` | string | Unique identifier for the API request. |
| `path` | string | Request URI path (can be a JQ expression). |
| `payload` | string | Request body payload (can be a JQ expression). |
| `verb` | string | HTTP method (defaults to GET if omitted). |
| `continueOnError` | bool | Controls behavior on HTTP errors; if true, it continues processing other APIs, otherwise (default: false), it stops. |
| `errorKey` | string | Used when continueOnError=true, defines the key name in the JSON results for storing error details (default: "error"). |

#### `dependsOn` Object Properties

| Property  | Type  | Description |
|-----------|-------|-------------|
| `iterator` | string | A JQ expression that returns a JSON array on which to iterate. |
| `name` | string | Name of another API on which this depends. |

#### `endpointRef` Object Properties

> Reference to a Kubernetes secret that describes the HTTP REST API endpoint.

| Property  | Type  | Description |
|-----------|-------|-------------|
| `name` | string | Name of the referenced object. |
| `namespace` | string | Namespace of the referenced object. |

---

### `status` Properties
The `status` field is an open-ended object that preserves unknown fields for storing results of all the `api` calls.


## Examples

Some `RESTAction` examples can be found [here](testdata/restactions/). Lets check the one that query [jsonplaceholder.typicode.com](https://jsonplaceholder.typicode.com) API:

```yaml
---
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: typicode-endpoint
  namespace: demo-system
stringData:
  server-url: https://jsonplaceholder.typicode.com
---
apiVersion: templates.krateo.io/v1
kind: RESTAction
metadata:
  name: typicode
  namespace: demo-system
spec:
  filter: "[.todos[] as $todo | .users[] | select(.id == $todo.userId) | { name: .name, id: $todo.id, title: $todo.title, completed: $todo.completed }]"
  api:
  - name: users
    path: "/users"
    endpointRef:
      name: typicode-endpoint
      namespace: demo-system
    filter: map(select(.email | endswith(".biz")))
  - name: todos
    dependsOn: 
      name: users
      iterator: .users
    path: ${ "/todos?userId=" + (.id|tostring) }
    headers:
      - ${ "X-UserID:" + (.id|tostring) }
    endpointRef:
      name: typicode-endpoint
      namespace: demo-system
```