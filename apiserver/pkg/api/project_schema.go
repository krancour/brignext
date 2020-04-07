package api

// nolint: lll
var projectSchemaBytes = []byte(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "github.com/lovethedrake/drakecore/config.schema.json",

	"definitions": {

		"empty": {
			"type": "string",
			"enum": [ "" ]
		},

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
					"type": [ "array", "null" ],
					"description": "Types of events from the provider",
					"items": { "$ref": "#/definitions/identifier" }
				}
			}
		},

		"containerConfig": {
			"type": "object",
			"description": "Configuration for an OCI container",
			"required": ["image"],
			"properties": {
				"image": {
					"type": "string",
					"description": "A URI for an OCI image"
				},
				"imagePullPolicy": {
					"type": "string",
					"description": "Pull policy for the OCI image",
					"enum": [ "", "IfNotPresent", "Always" ]
				},
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
				"allowSecretKeyRef": {
					"type": "boolean",
					"description": "Whether to permit the worker to reference Kubernetes secrets when defining environment variables for jobs"
				},
				"imagePullSecrets": {
					"oneOf": [
						{ "$ref": "#/definitions/empty" },
						{ "$ref": "#/definitions/identifier" }
					],
					"description": "A Kubernetes secret that can be used as an image pull secret for job images"
				}
			}
		},

		"gitConfig": {
			"type": "object",
			"description": "Worker configuration pertaining specifically to git",
			"properties": {
				"cloneURL": {
					"oneOf": [
						{ "$ref": "#/definitions/empty" },
						{ "$ref": "#/definitions/url" }
					],
					"description": "The URL for cloning a git project"
				},
				"initSubmodules": {
					"type": "boolean",
					"description": "Whether to initialize git submodules"
				}
			}
		},

		"kubernetesConfig": {
			"type": "object",
			"description": "Worker configuration pertaining specifically to Kubernetes",
			"properties": {
				"workspaceStorageClass": {
					"oneOf": [
						{ "$ref": "#/definitions/empty" },
						{ "$ref": "#/definitions/identifier" }
					],
					"description": "The Kubernetes storage class that a worker may use when provisioning a volume to be shared across multiple jobs"
				},
				"imagePullSecrets": {
					"oneOf": [
						{ "$ref": "#/definitions/empty" },
						{ "$ref": "#/definitions/identifier" }
					],
					"description": "A Kubernetes secret that can be used as an image pull secret for the worker's image"
				}
			}
		},

		"workerConfig": {
			"type": "object",
			"description": "Configuration for a single Brigade worker",
			"additionalProperties": false,
			"properties": {
				"events": {
					"type": [ "array", "null" ],
					"description": "The events that trigger this worker",
					"items": { "$ref": "#/definitions/triggeringEvents" }
				},
				"container": {
					"allOf": [{ "$ref": "#/definitions/containerConfig" }],
					"description": "Configuration for the worker's main container"
				},
				"workspaceSize": {
					"type": "string",
					"description": "The amount of storage to be provisioned for a worker"
				},
				"git": { "$ref": "#/definitions/gitConfig" },
				"kubernetes": { "$ref": "#/definitions/kubernetesConfig" },
				"jobsConfig": { "$ref": "#/definitions/jobsConfig" },
				"logLevel": {
					"type": "string",
					"description": "Log level to be observed by the worker",
					"enum": [ "", "DEBUG", "INFO", "WARN", "ERROR" ]
				}
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
			"oneOf": [
				{ "$ref": "#/definitions/empty" },
				{ "$ref": "#/definitions/description" }
			],
			"description": "A brief description of the project"
		},
		"tags": {
			"type": [ "object", "null" ],
			"additionalProperties": true,
			"patternProperties": {
				"^\\w[\\w-]*$": { "$ref": "#/definitions/identifier" }
			}
		},
		"secrets": {
			"type": [ "object", "null" ],
			"description": "A map of project secrets"
		},
		"workerConfigs": {
			"type": [ "object", "null" ],
			"description": "A map of worker configurations indexed by unique names",
			"additionalProperties": false,
			"patternProperties": {
				"^\\w[\\w-]*$": { "$ref": "#/definitions/workerConfig" }
			}
		}
	}
}
`)
