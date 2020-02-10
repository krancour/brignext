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

		"event": {
			"type": "string",
			"pattern": "^(?:\\*)|(?:\\w[\\w-]+\\:(?:(?:\\*)|(?:\\w[\\w-]+)))$",
			"minLength": 1,
			"maxLength": 40
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

		"workerConfig": {
			"type": "object",
			"description": "Configuration for a single Brigade worker",
			"additionalProperties": false,
			"properties": {
				"events": {
					"type": "array",
					"description": "The events that trigger this worker",
					"items": { "$ref": "#/definitions/event" }
				},
				"image": { "$ref": "#/definitions/image" },
				"command": {
					"type": "string",
					"description": "The command to execute within the worker container"
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
			"allOf": [{ "$ref": "#/definitions/description" }],
			"description": "A brief description of the project"
		},
		"workerConfigs": {
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
