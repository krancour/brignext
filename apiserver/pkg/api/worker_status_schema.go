package api

var workerStatusSchemaBytes = []byte(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "github.com/lovethedrake/drakecore/config.schema.json",

	"title": "WorkerStatus",
	"type": "object",
	"required": ["phase"],
	"additionalProperties": false,
	"properties": {
		"started": {
			"type": [ "string", "null" ],
			"format": "date-time",
			"description": "The time at which the worker started"
		},
		"ended": {
			"type": [ "string", "null" ],
			"format": "date-time",
			"description": "The time at which the worker completed"
		},
		"phase": {
			"type": "string",
			"description": "The worker's phase",
			"enum": [ "PENDING", "RUNNING", "CANCELED", "ABORTED", "SUCCEEDED", "FAILED", "UNKNOWN" ]
		}
	}
}
`)
