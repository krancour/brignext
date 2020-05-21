package brignext

type SecretList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []Secret `json:"items"`
}

func NewSecretList() SecretList {
	return SecretList{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "SecretList",
		},
		ListMeta: ListMeta{},
		Items:    []Secret{},
	}
}

type Secret struct {
	TypeMeta `json:",inline" bson:",inline"`
	// TODO: Secret isn't really a sub-resource of a project-- it might not be
	// appropriate for it to utilize ObjectMeta. Certainly, the constraints we
	// normally place on an ObjectMeta.ID are too narrow for all reasonable
	// secret names, and the IDs we're currently using don't uniquely identify
	// a secret either. So, a secret is, perhaps, more of a "synthetic" resource.
	// Perhaps remove ObjectMeta and add a Key field.
	ObjectMeta `json:"metadata" bson:"metadata"`
	Value      string `json:"value" bson:"value"`
}

func NewSecret(key, value string) Secret {
	return Secret{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "Secret",
		},
		ObjectMeta: ObjectMeta{
			ID: key,
		},
		Value: value,
	}
}
