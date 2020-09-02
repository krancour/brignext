package api

// Role represents a set of permissions, with domain-specific meaning, held by a
// principal, such as a User or ServiceAccount.
type Role struct {
	// Name is the name of a Role and has domain-specific meaning.
	Name string `json:"name"`
	// Scope qualifies the scope of the Role. The value is opaque and has meaning
	// only in relation to a specific RoleName.
	Scope string `json:"scope"`
}

type RoleAssignment struct {
	Role             string `json:"role"`
	UserID           string `json:"userID"`
	ServiceAccountID string `json:"serviceAccountID"`
}
