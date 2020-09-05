package authx

type PrincipalType string

const (
	PrincipalTypeUser           PrincipalType = "USER"
	PrincipalTypeServiceAccount PrincipalType = "SERVICE_ACCOUNT"
)
