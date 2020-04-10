package api

var jobStatusSchemaBytes = []byte(`
{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"$id": "github.com/lovethedrake/drakecore/config.schema.json",

	"title": "JobStatus",
	"type": "object",
	"required": ["phase"],
	"additionalProperties": false,
	"properties": {
		"started": {
			"type": [ "string", "null" ],
			"format": "date-time",
			"description": "The time at which the job started"
		},
		"ended": {
			"type": [ "string", "null" ],
			"format": "date-time",
			"description": "The time at which the job completed"
		},
		"phase": {
			"type": "string",
			"description": "The job's phase",
			"enum": [ "RUNNING", "ABORTED", "SUCCEEDED", "FAILED", "UNKNOWN" ]
		}
	}
}
`)
