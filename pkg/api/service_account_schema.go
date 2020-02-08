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
			"maxLength": 50
		},

		"description": {
			"type": "string",
			"maxLength": 80
		}

	},

	"title": "ServiceAccount",
	"type": "object",
	"required": ["name", "description"],
	"additionalProperties": false,
	"properties": {
		"name": {
			"type": "string",
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The name of the service account"
		},
		"description": {
			"type": "string",
			"allOf": [{ "$ref": "#/definitions/description" }],
			"description": "A brief description of the service account"
		}
	}
}
`)
