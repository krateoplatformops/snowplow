// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/health": {
            "get": {
                "description": "Health Check",
                "produces": [
                    "application/json"
                ],
                "summary": "Liveness Endpoint",
                "operationId": "health",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/health.ServiceInfo"
                        }
                    }
                }
            }
        },
        "/list": {
            "get": {
                "description": "Resources List",
                "produces": [
                    "application/json"
                ],
                "summary": "List resources by category in a specified namespace.",
                "operationId": "list",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Krateo User",
                        "name": "X-Krateo-User",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Krateo User Groups",
                        "name": "X-Krateo-Groups",
                        "in": "header",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Resource category",
                        "name": "category",
                        "in": "query",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Namespace",
                        "name": "ns",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "object",
                            "additionalProperties": true
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/status.Status"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/status.Status"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/status.Status"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/status.Status"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "health.ServiceInfo": {
            "type": "object",
            "properties": {
                "build": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "namespace": {
                    "type": "string"
                }
            }
        },
        "status.Status": {
            "type": "object",
            "properties": {
                "apiVersion": {
                    "type": "string"
                },
                "code": {
                    "description": "Suggested HTTP return code for this status, 0 if not set.",
                    "type": "integer"
                },
                "kind": {
                    "type": "string"
                },
                "message": {
                    "description": "A human-readable description of the status of this operation.",
                    "type": "string"
                },
                "reason": {
                    "description": "A machine-readable description of why this operation is in the\n\"Failure\" status. If this value is empty there\nis no information available. A Reason clarifies an HTTP status\ncode but does not override it.",
                    "allOf": [
                        {
                            "$ref": "#/definitions/status.StatusReason"
                        }
                    ]
                },
                "status": {
                    "description": "Status of the operation.\nOne of: \"Success\" or \"Failure\".",
                    "type": "string"
                }
            }
        },
        "status.StatusReason": {
            "type": "string",
            "enum": [
                "",
                "Unauthorized",
                "Forbidden",
                "NotFound",
                "Conflict",
                "Gone",
                "Invalid",
                "Timeout",
                "TooManyRequests",
                "BadRequest",
                "MethodNotAllowed",
                "NotAcceptable",
                "RequestEntityTooLarge",
                "UnsupportedMediaType",
                "InternalError",
                "ServiceUnavailable"
            ],
            "x-enum-varnames": [
                "StatusReasonUnknown",
                "StatusReasonUnauthorized",
                "StatusReasonForbidden",
                "StatusReasonNotFound",
                "StatusReasonConflict",
                "StatusReasonGone",
                "StatusReasonInvalid",
                "StatusReasonTimeout",
                "StatusReasonTooManyRequests",
                "StatusReasonBadRequest",
                "StatusReasonMethodNotAllowed",
                "StatusReasonNotAcceptable",
                "StatusReasonRequestEntityTooLarge",
                "StatusReasonUnsupportedMediaType",
                "StatusReasonInternalError",
                "StatusReasonServiceUnavailable"
            ]
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1.0",
	Host:             "",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "SnowPlow API",
	Description:      "This the total new Krateo backend.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}