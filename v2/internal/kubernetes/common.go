package kubernetes

const (
	LabelComponent = "brigade.sh/component"
	LabelProject   = "brigade.sh/project"
	LabelEvent     = "brigade.sh/event"
	LabelJob       = "brigade.sh/job"

	SecretTypeProjectSecrets = "brigade.sh/project-secrets"
	SecretTypeEvent          = "brigade.sh/event"
	SecretTypeJobSecrets     = "brigade.sh/job-secrets"
)
