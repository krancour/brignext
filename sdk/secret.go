package sdk

import "github.com/krancour/brignext/v2/sdk/meta"

type Secret struct {
	meta.TypeMeta `json:",inline"`
	Key           string `json:"key"`
	Value         string `json:"value"`
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

type SecretReference struct {
	meta.TypeMeta `json:",inline"`
	Key           string `json:"key"`
}

type SecretReferenceList struct {
	meta.TypeMeta `json:",inline"`
	meta.ListMeta `json:"metadata"`
	Items         []SecretReference `json:"items"`
}
