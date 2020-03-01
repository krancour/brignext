package brignext

// nolint: lll
type GitConfig struct {
	CloneURL       string `json:"cloneURL" bson:"cloneURL"`
	Commit         string `json:"commit,omitempty" bson:"commit"`
	Ref            string `json:"ref,omitempty" bson:"ref"`
	InitSubmodules *bool  `json:"initSubmodules,omitempty" bson:"initSubmodules"`
	// // TODO: We MUST encrypt this!
	// SSHKey  string `json:"sshKey,omitempty" bson:"sshKey"`
	// SSHCert string `json:"sshCert,omitempty" bson:"sshCert"`
}
