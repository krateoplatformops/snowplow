# `snowplow`

Snowplow is a web service that plays a key role within the suite of components that make up Krateo PlatformOps. 

It serves multiple purposes, primarily acting as a bridge between Krateo's custom resources and the UI. By enabling a dynamic and declarative approach to defining UI components and layouts, Snowplow ensures that the interface is interpreted and rendered seamlessly by Krateo Frontend.

Currently, Snowplow handles on-demand resolution of the [RESTAction](RESTAction.md) custom resource. For all other custom resources within Krateo, it delegates the task to the Kubernetes API server, which, in turn, leverages Krateo’s custom aggregation API layer to process them. 

Looking ahead, Snowplow will take full responsibility for managing all custom resources beyond RESTAction, eventually eliminating the need for delegation altogether.

As of today, the overall architecture is structured as follows:

![](_images/snowplow.png)

