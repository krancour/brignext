package api

// nolint: lll
var projectSchemaBytes = []byte(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "github.com/lovethedrake/drakecore/config.schema.json",

	"definitions": {

		"identifier": {
			"type": "string",
			"pattern": "^\\w[\\w-]*$",
			"minLength": 3,
			"maxLength": 50
		},

		"description": {
			"type": "string",
			"minLength": 3,
			"maxLength": 80
		},

		"url": {
			"type": "string",
			"pattern": "^[\\w:/\\-\\.\\?=]*$",
			"minLength": 5,
			"maxLength": 250
		},

		"tag": {
			"type": "string",
			"pattern": "^\\w[\\w-\\.]*$",
			"minLength": 3,
			"maxLength": 50
		},

		"image": {
			"type": "object",
			"description": "An OCI image",
			"required": ["repository"],
			"additionalProperties": false,
			"properties": {
				"repository": {
					"allOf": [{ "$ref": "#/definitions/url" }],
					"description": "The OCI image repository"
				},
				"tag": {
					"allOf": [{ "$ref": "#/definitions/tag" }],
					"description": "The tag for the correct OCI image from the repository"
				},
				"pullPolicy": {
					"type": "string",
					"description": "Pull policy for the OCI image",
					"enum": [ "IfNotPresent", "Always" ]
				}
			}
		},

		"triggeringEvents": {
			"type": "object",
			"description": "Describes a set of events that trigger a worker",
			"required": ["provider"],
			"additionalProperties": false,
			"properties": {
				"provider": {
					"allOf": [{ "$ref": "#/definitions/identifier" }],
					"description": "The name of the event provider"
				},
				"types": {
					"type": "array",
					"description": "Types of events from the provider",
					"items": { "$ref": "#/definitions/identifier" }
				}
			}
		},

		"containerConfig": {
			"type": "object",
			"description": "Configuration for an OCI container",
			"properties": {
				"image": { "$ref": "#/definitions/image" },
				"command": {
					"type": "string",
					"description": "The command to execute within the container"
				}
			}
		},

		"jobsConfig": {
			"type": "object",
			"description": "Configuration for any job containers the worker container might fan out to",
			"properties": {
				"allowPrivileged": {
					"type": "boolean",
					"description": "Whether job containers are permitted to be run as privileged"
				},
				"allowHostMounts": {
					"type": "boolean",
					"description": "Whether job containers are permitted to mount files or directories from the container host"
				},
				"kubernetes": { "$ref": "#/definitions/jobsKubernetesConfig" }
			}
		},

		"jobsKubernetesConfig": {
			"type": "object",
			"description": "Jobs configuration pertaining specifically to Kubernetes",
			"properties": {
				"cacheStorageClass": {
					"allOf": [{ "$ref": "#/definitions/identifier" }],
					"description": "The Kubernetes storage class that a worker may use when provisioning a volume to cache artifacts across multiple executions of a job"
				},
				"allowSecretKeyRef": {
					"type": "boolean",
					"description": "Whether to permit the worker to reference Kubernetes secrets when defining environment variables for jobs"
				},
				"serviceAccount": {
					"allOf": [{ "$ref": "#/definitions/identifier" }],
					"description": "A specific service account the worker should use for any jobs it launches"
				}
			}
		},

		"kubernetesConfig": {
			"type": "object",
			"description": "Worker configuration pertaining specifically to Kubernetes",
			"properties": {
				"workspaceStorageClass": {
					"allOf": [{ "$ref": "#/definitions/identifier" }],
					"description": "The Kubernetes storage class that a worker may use when provisioning a volume to be shared across multiple jobs"
				},
				"serviceAccount": {
					"allOf": [{ "$ref": "#/definitions/identifier" }],
					"description": "A specific service account to use for the worker"
				}
			}
		},

		"workerConfig": {
			"type": "object",
			"description": "Configuration for a single Brigade worker",
			"additionalProperties": false,
			"properties": {
				"events": {
					"type": "array",
					"description": "The events that trigger this worker",
					"items": { "$ref": "#/definitions/triggeringEvents" }
				},
				"initContainer": {
					"allOf": [{ "$ref": "#/definitions/containerConfig" }],
					"description": "Configuration for the worker's init container"		
				},
				"container": {
					"allOf": [{ "$ref": "#/definitions/containerConfig" }],
					"description": "Configuration for the worker's main container"
				},
				"workspaceSize": {
					"type": "string",
					"description": "The amount of storage to be provisioned for a worker"
				},
				"kubernetes": { "$ref": "#/definitions/kubernetesConfig" },
				"jobs": { "$ref": "#/definitions/jobsConfig" }
			}
		}

	},

	"title": "Project",
	"type": "object",
	"required": ["id"],
	"additionalProperties": false,
	"properties": {
		"id": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "A meaningful identifier for the project"
		},
		"description": {
			"allOf": [{ "$ref": "#/definitions/description" }],
			"description": "A brief description of the project"
		},
		"tags": {
			"type": "object",
			"additionalProperties": true,
			"patternProperties": {
				"^\\w[\\w-]*$": { "$ref": "#/definitions/identifier" }
			}
		},
		"secrets": {
			"type": "object",
			"description": "A map of project secrets"
		},
		"workers": {
			"type": "object",
			"description": "A map of worker configurations indexed by unique names",
			"additionalProperties": false,
			"patternProperties": {
				"^\\w[\\w-]*$": { "$ref": "#/definitions/workerConfig" }
			}
		}
	}
}
`)
