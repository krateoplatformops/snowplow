# `snowplow`

Snowplow is a web service that plays a key role within the suite of components that make up Krateo PlatformOps. 

It serves multiple purposes, primarily acting as a bridge between Krateo's custom resources and the UI. By enabling a dynamic and declarative approach to defining UI components and layouts, Snowplow ensures that the interface is interpreted and rendered seamlessly by Krateo Frontend.

It handles on-demand resolution of the `RESTAction` custom resource and all Krateo Frontend `Widgets` custom resources.

## Learn More

### Developer Guide

- [Building `snowplow`](howto/developer-guide.md)
- [ADR: Decoupling `authn` from `snowplow` for Testing and Operations](howto/decoupling-authn-from-snowplow-for-testing.md)

### User Guide

- [`Endpoint` reference](howto/endpoints.md)
- [`RESTAction` reference](howto/restactions.md)
- [Understanding the `Widget` Custom Resource](howto/widgets.md)
- [Installing `snowplow` on Kind](howto/install.md)

### Examples

- [RESTAction: list _cluster namespaces_](howto/restactions/example-cluster-namespaces.md)
- [RESTAction: invoke _external API_](howto/restactions/example-external-api.md)
