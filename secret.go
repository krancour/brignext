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
	Key      string `json:"key" bson:"value"`
	Value    string `json:"value" bson:"value"`
}

func NewSecret(key, value string) Secret {
	return Secret{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "Secret",
		},
		Key:   key,
		Value: value,
	}
}
