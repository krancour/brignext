package api

// nolint: lll
var serviceAccountSchemaBytes = []byte(`
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
		}

	},

	"title": "ServiceAccount",
	"type": "object",
	"required": ["id", "description"],
	"additionalProperties": false,
	"properties": {
		"id": {
			"type": "string",
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "A meaningful identifier for the service account"
		},
		"description": {
			"type": "string",
			"allOf": [{ "$ref": "#/definitions/description" }],
			"description": "A brief description of the service account"
		}
	}
}
`)
