package sdk

import "github.com/krancour/brignext/v2/apiserver/internal/sdk/meta"

type SecretList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []Secret `json:"items"`
}

func NewSecretList() SecretList {
	return SecretList{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "SecretList",
		},
		ListMeta: meta.ListMeta{},
		Items:    []Secret{},
	}
}

type Secret struct {
	meta.TypeMeta `json:",inline" bson:",inline"`
	Key           string `json:"key" bson:"value"`
	Value         string `json:"value" bson:"value"`
}

func NewSecret(key, value string) Secret {
	return Secret{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Secret",
		},
		Key:   key,
		Value: value,
	}
}
