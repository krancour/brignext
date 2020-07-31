package sdk

import "github.com/krancour/brignext/v2/sdk/meta"

type Token struct {
	meta.TypeMeta `json:",inline"`
	Value         string `json:"value"`
}

func NewToken(value string) Token {
	return Token{
		TypeMeta: meta.TypeMeta{
			APIVersion: meta.APIVersion,
			Kind:       "Token",
		},
		Value: value,
	}
}
