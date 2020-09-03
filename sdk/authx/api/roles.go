package api

type UserRoleAssignment struct {
	UserID string `json:"userID"`
	Role   string `json:"role"`
}

type ServiceAccountRoleAssignment struct {
	ServiceAccountID string `json:"serviceAccountID"`
	Role             string `json:"role"`
}
