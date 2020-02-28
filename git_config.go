package brignext

// nolint: lll
type GitConfig struct {
	CloneURL       string `json:"cloneURL" bson:"cloneURL"`
	Commit         string `json:"commit,omitempty" bson:"commit,omitempty"`
	Ref            string `json:"ref,omitempty" bson:"ref,omitempty"`
	InitSubmodules *bool  `json:"initSubmodules,omitempty" bson:"initSubmodules,omitempty"`
	// // TODO: We MUST encrypt this!
	// SSHKey  string `json:"sshKey,omitempty" bson:"sshKey,omitempty"`
	// SSHCert string `json:"sshCert,omitempty" bson:"sshCert,omitempty"`
}

// type Github struct {
// 	// TODO: We MUST encrypt this!
// 	Token     string `json:"token,omitempty" bson:"token,omitempty"`
// 	BaseURL   string `json:"baseURL,omitempty" bson:"baseURL,omitempty"`
// 	UploadURL string `json:"uploadURL,omitempty" bson:"uploadURL,omitempty"`
// }
