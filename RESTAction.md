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


