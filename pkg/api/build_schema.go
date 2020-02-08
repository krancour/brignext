package api

// nolint: lll
var buildSchemaBytes = []byte(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "github.com/lovethedrake/drakecore/config.schema.json",

	"definitions": {

		"identifier": {
			"type": "string",
			"pattern": "^\\w[\\w-]*$",
			"minLength": 3,
			"maxLength": 50
		}

	},

	"title": "Build",
	"type": "object",
	"required": ["projectName", "provider", "type"],
	"additionalProperties": false,
	"properties": {
		"projectName": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The name of the project the build is for"
		},
		"provider": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The name of the provider that triggered the build"
		},
		"type": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The type of event that triggered the build"
		}
	}
}
`)
