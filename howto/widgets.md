# Understanding the `Widget` Custom Resource

The `Widget` Custom Resource (CR) defines a dynamic, declarative model for UI components within Krateo PlatformOps.

It acts as the bridge between frontend-defined layout specifications and the backend logic provided by the `snowplow` web service.


## Overview

A `Widget` resource describes how a particular UI component (e.g., a button, a form, a chart) should be rendered in the Krateo frontend.

Unlike traditional custom resources that follow fixed schemas, the `Widget` CR is partially dynamic — it combines:

- a frontend-defined section, where UI developers can freely describe the widget’s behavior and layout
- a runtime-managed section, handled by `snowplow`, which performs resolution, evaluation, and enrichment tasks

This hybrid design allows frontend developers to define *what* a component should look like, while `snowplow` determines *how* it behaves at runtime.

Refer to the [Krateo Widgets documentation](https://github.com/krateoplatformops/frontend/blob/main/docs/docs.md) for a more detailed guide on Widgets.


## Runtime Flow

```text
Frontend Developer ──► defines JSON Schema (widgetData)
                          │
                          ▼
                  krateoctl generates CRD
                          │
                          ▼
                  Widget Custom Resource created (user)
                          │
                          ▼
                  snowplow resolves CR on demand
                          │
                          ▼
               evaluated results written to status
                          │
                          ▼
              Krateo Frontend reads status → renders component
```
