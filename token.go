package brignext

type Token struct {
	TypeMeta `json:",inline" bson:",inline"`
	Value    string `json:"value" bson:"value"`
}
