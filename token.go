package brignext

type Token struct {
	TypeMeta `json:",inline" bson:",inline"`
	Value    string `json:"value" bson:"value"`
}

func NewToken(value string) Token {
	return Token{
		TypeMeta: TypeMeta{
			APIVersion: APIVersion,
			Kind:       "Token",
		},
		Value: value,
	}
}
