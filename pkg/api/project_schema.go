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

		"repo": {
			"type": "object",
			"description": "A source code repository for the project",
			"required": ["cloneURL"],
			"additionalProperties": false,
			"properties": {
				"cloneURL": {
					"allOf": [{ "$ref": "#/definitions/url" }],
					"description": "The URL where the respository can be cloned from"
				}
			}
		}

	},

	"title": "Project",
	"type": "object",
	"required": ["name", "repo"],
	"additionalProperties": false,
	"properties": {
		"name": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The name of the project"
		},
		"description": {
			"allOf": [{ "$ref": "#/definitions/description" }],
			"description": "A brief description of the project"
		},
		"repo": { "$ref": "#/definitions/repo" }
	}
}
`)
