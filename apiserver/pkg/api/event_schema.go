package api

// nolint: lll
var eventSchemaBytes = []byte(`
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

		"url": {
			"type": "string",
			"pattern": "^[\\w:/\\-\\.\\?=]*$",
			"minLength": 5,
			"maxLength": 250
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
				}
			},
			"commit": {
				"type": "string",
				"description": "A git commit sha"
			},
			"ref": {
				"type": "string",
				"description": "A reference to a git branch or tag"
			},
			"initSubmodules": {
				"type": "boolean",
				"description": "Whether to initialize git submodules"
			}
		}

	},

	"title": "Event",
	"type": "object",
	"required": ["projectID", "source", "type"],
	"additionalProperties": false,
	"properties": {
		"projectID": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The ID of the project the event is for"
		},
		"source": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The name of the source that is sending the event"
		},
		"type": {
			"allOf": [{ "$ref": "#/definitions/identifier" }],
			"description": "The type of the event"
		},
		"shortTitle": {
			"type": "string",
			"description": "A succint description of the event",
			"maxLength": 50
		},
		"longTitle": {
			"type": "string",
			"description": "A detailed description of the event",
			"maxLength": 100
		},
		"git": { "$ref": "#/definitions/gitConfig" }
	}
}
`)
