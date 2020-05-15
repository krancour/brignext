package brignext

type SecretList struct {
	TypeMeta `json:",inline"`
	ListMeta `json:"metadata"`
	Items    []Secret `json:"items"`
}

type Secret struct {
	TypeMeta   `json:",inline" bson:",inline"`
	ObjectMeta `json:"metadata" bson:"metadata"`
	Value      string `json:"value" bson:"value"`
}
